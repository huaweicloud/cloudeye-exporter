package collector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

type BaseHuaweiCloudExporter struct {
	From            int64
	To              int64
	Namespaces      []string
	Prefix          string
	ClientConfig    *Config
	Region          string
	txnKey          string
	MaxRoutines     int
	ScrapeBatchSize int
}

func replaceName(name string) string {
	newName := strings.Replace(name, ".", "_", -1)
	newName = strings.ToLower(newName)

	return newName
}

func GetMonitoringCollector(namespaces []string) *BaseHuaweiCloudExporter {
	exporter := &BaseHuaweiCloudExporter{
		Namespaces:      namespaces,
		Prefix:          CloudConf.Global.Prefix,
		MaxRoutines:     CloudConf.Global.MaxRoutines,
		ClientConfig:    conf,
		ScrapeBatchSize: CloudConf.Global.ScrapeBatchSize,
	}
	return exporter
}

type PrometheusMetricMap = struct {
	sync.RWMutex
	MetricMap map[string]bool // key:txnKey value: metric map for deduplicate label key:label
}

// Describe simply sends the two Descs in the struct to the channel.
func (exporter *BaseHuaweiCloudExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

func (exporter *BaseHuaweiCloudExporter) listMetrics(namespace string) ([]model.MetricInfoList, map[string]labelInfo) {
	allResourcesInfo, metrics := exporter.listAllResources(namespace)
	logs.Logger.Debugf("[%s] Resource number of %s: %d", exporter.txnKey, namespace, len(allResourcesInfo))

	if len(metrics) > 0 {
		return metrics, allResourcesInfo
	}
	// 用户指定了EPID但是metrics为空的场景下,则不应返回数据
	if CloudConf.Global.EpIds != "" {
		return nil, nil
	}
	logs.Logger.Debugf("[%s] Start to getAllMetric from CES", exporter.txnKey)
	allMetrics, err := listAllMetrics(namespace)
	if err != nil {
		logs.Logger.Errorf("[%s] Get all metrics error: %s", exporter.txnKey, err.Error())
		return nil, nil
	}
	logs.Logger.Debugf("[%s] End to getAllMetric, Total number of of metrics: %d", exporter.txnKey, len(allMetrics))
	return allMetrics, allResourcesInfo
}

func (exporter *BaseHuaweiCloudExporter) setProData(ctx context.Context, ch chan<- prometheus.Metric,
	dataList []model.BatchMetricData, allResourcesInfo map[string]labelInfo, metricMap *PrometheusMetricMap) {
	defer func() {
		if err := recover(); err != nil {
			logs.Logger.Errorf("[%s] SetProData error: %+v", exporter.txnKey, err)
		}
	}()
	for _, metric := range dataList {
		exporter.debugMetricInfo(metric)
		data, err := getLatestData(metric.Datapoints)
		if err != nil {
			logs.Logger.Warnf("[%s] Get data point error: %s, metric_name: %s, dimension: %+v", exporter.txnKey, err.Error(), metric.MetricName, metric.Dimensions)
			continue
		}

		label := getLabel(metric, allResourcesInfo)
		fqName := prometheus.BuildFQName(exporter.Prefix, replaceName(*metric.Namespace), metric.MetricName)

		ip := "-" // 添加常量标签，默认值为-，有对应的映射关系就是对应的ip
		for i, name := range label.Name {
			if strings.HasSuffix(name, "cluster_id") || strings.HasSuffix(name, "node_id") {
				if value, exists := MIPId[label.Value[i]]; exists {
					ip = value
				}
			}
		}
		constLabels := prometheus.Labels{
			"ip": ip,
		}

		proMetric, err := prometheus.NewConstMetric(
			prometheus.NewDesc(fqName, fqName, label.Name, constLabels),
			prometheus.GaugeValue, data, label.Value...)
		if err != nil {
			logs.Logger.Errorf("[%s] New const metric error: %s, fqName: %s, label: %+v",
				exporter.txnKey, err.Error(), fqName, label)
			continue
		}
		dimArray := make([]string, 0, len(*metric.Dimensions))
		for _, dimension := range *metric.Dimensions {
			dimArray = append(dimArray, dimension.Name)
		}
		dimNameStr := strings.Join(dimArray, ",")
		if isAgentMetric(*metric.Namespace) && isMetricLabelConflict(fqName, label, metricMap) {
			logs.Logger.Warnf("[%s] Metric label conflict, namespace: %s , dimension: %s, metric name: %s", exporter.txnKey, *metric.Namespace, dimNameStr, metric.MetricName)
			continue
		}
		if err := sendMetricData(ctx, ch, proMetric); err != nil {
			logs.Logger.Errorf("[%s] Context has canceled, no need to send metric data, metric name: %s", exporter.txnKey, fqName)
		}
	}
}

func isAgentMetric(namespace string) bool {
	return namespace == "AGT.ECS" || namespace == "SERVICE.BMS"
}

func getLabel(metric model.BatchMetricData, info map[string]labelInfo) labelInfo {
	label := getDimLabel(metric)
	if extendLabel, exist := info[GetResourceKeyFromMetricData(metric)]; exist {
		label.Name = append(label.Name, extendLabel.Name...)
		label.Value = append(label.Value, extendLabel.Value...)
	}
	return label
}

func getDimLabel(metric model.BatchMetricData) labelInfo {
	var label labelInfo
	for _, dim := range *metric.Dimensions {
		label.Name = append(label.Name, strings.ReplaceAll(dim.Name, "-", "_"))
		label.Value = append(label.Value, getDimValue(*metric.Namespace, dim.Name, dim.Value))
	}
	label.Name = append(label.Name, "unit")
	label.Value = append(label.Value, *metric.Unit)
	return label
}

func getDimValue(namespaces, dimName, dimValue string) string {
	if !isContainsInStringArr(namespaces, []string{"AGT.ECS", "SERVICE.BMS"}) {
		return dimValue
	}

	if !isContainsInStringArr(dimName, []string{"mount_point", "disk", "proc", "gpu", "raid"}) {
		return dimValue
	}

	return getAgentOriginValue(dimValue)
}

func isContainsInStringArr(target string, array []string) bool {
	for index := range array {
		if target == array[index] {
			return true
		}
	}
	return false
}

func (exporter *BaseHuaweiCloudExporter) collectMetricByNamespace(ctx context.Context, ch chan<- prometheus.Metric, namespace string, proMap *PrometheusMetricMap) {
	defer func() {
		if err := recover(); err != nil {
			logs.Logger.Errorf("[%s] recover error: %+v", exporter.txnKey, err)
		}
	}()

	allMetrics, allResourcesInfo := exporter.listMetrics(namespace)
	if len(allMetrics) == 0 {
		logs.Logger.Warnf("[%s] Metrics of %s are not found, skip.", exporter.txnKey, namespace)
		return
	}

	logs.Logger.Debugf("[%s] Start to scrape metric data", exporter.txnKey)
	workChan := make(chan struct{}, exporter.MaxRoutines)
	defer close(workChan)
	var wg sync.WaitGroup
	count := 0
	tmpMetrics := make([]model.MetricInfo, 0, exporter.ScrapeBatchSize)
	metricsMap := make(map[string]bool, 0)
	for _, metric := range allMetrics {
		dimsValueKey := fmt.Sprintf("%s,%s", getDimsValueKey(metric.Dimensions), metric.MetricName)
		if _, ok := metricsMap[dimsValueKey]; ok {
			continue
		}
		metricsMap[dimsValueKey] = true
		count++
		tmpMetrics = append(tmpMetrics, transMetric(metric))
		if (len(tmpMetrics) == exporter.ScrapeBatchSize) || (count == len(allMetrics)) {
			workChan <- struct{}{}
			wg.Add(1)
			go func(tmpMetrics []model.MetricInfo) {
				defer func() {
					<-workChan
					wg.Done()
				}()
				logs.Logger.Debugf("[%s] Start to getBatchMetricData, metric count: %d", exporter.txnKey, len(tmpMetrics))
				dataList, err := batchQueryMetricData(&tmpMetrics, exporter.From, exporter.To)
				if err != nil {
					return
				}
				exporter.setProData(ctx, ch, *dataList, allResourcesInfo, proMap)
			}(tmpMetrics)
			tmpMetrics = make([]model.MetricInfo, 0, exporter.ScrapeBatchSize)
		}
	}

	wg.Wait()
	logs.Logger.Debugf("[%s] End to scrape all metric data", exporter.txnKey)
}

func transMetric(metricInfoList model.MetricInfoList) model.MetricInfo {
	return model.MetricInfo{
		Dimensions: metricInfoList.Dimensions,
		Namespace:  metricInfoList.Namespace,
		MetricName: metricInfoList.MetricName,
	}
}

func (exporter *BaseHuaweiCloudExporter) Collect(ch chan<- prometheus.Metric) {
	duration, err := time.ParseDuration("-10m")
	if err != nil {
		logs.Logger.Error("ParseDuration -10m error:", err.Error())
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	now := time.Now()
	exporter.From = now.Add(duration).UnixNano() / 1e6
	exporter.To = now.UnixNano() / 1e6
	exporter.txnKey = fmt.Sprintf("%s-%d-%d", strings.Join(exporter.Namespaces, "-"), exporter.From, exporter.To)

	logs.Logger.Debugf("[%s] Start to collect data", exporter.txnKey)
	var wg sync.WaitGroup
	proMap := PrometheusMetricMap{
		RWMutex:   sync.RWMutex{},
		MetricMap: make(map[string]bool),
	}
	for _, namespace := range exporter.Namespaces {
		wg.Add(1)
		go func(ctx context.Context, ch chan<- prometheus.Metric, namespace string) {
			defer wg.Done()
			exporter.collectMetricByNamespace(ctx, ch, namespace, &proMap)
		}(ctx, ch, namespace)
	}
	wg.Wait()
	logs.Logger.Debugf("[%s] End to collect data", exporter.txnKey)
}

func sendMetricData(ctx context.Context, ch chan<- prometheus.Metric, metric prometheus.Metric) error {
	// Check whether the Context has canceled
	select {
	case _, ok := <-ctx.Done():
		if !ok {
			return ctx.Err()
		}
	default: // continue
	}
	// If no, send the metric
	ch <- metric
	return nil
}

func (exporter *BaseHuaweiCloudExporter) debugMetricInfo(md model.BatchMetricData) {
	dataJson, err := json.Marshal(md)
	if err != nil {
		logs.Logger.Errorf("[%s] Marshal metricData error: %s", exporter.txnKey, err.Error())
		return
	}
	logs.Logger.Debugf("[%s] Get data points of metric are: %s", exporter.txnKey, string(dataJson))
}

func getLatestData(data []model.DatapointForBatchMetric) (float64, error) {
	if len(data) == 0 {
		return 0, errors.New("data not found")
	}

	return *data[len(data)-1].Average, nil
}

func isMetricLabelConflict(fqName string, label labelInfo, metricMap *PrometheusMetricMap) bool {
	labelArray := make([]string, 0, len(label.Name))
	for i := range label.Name {
		labelTmp := fmt.Sprintf("%s=%s", label.Name[i], label.Value[i])
		labelArray = append(labelArray, labelTmp)
	}
	labelResult := fmt.Sprintf("%s{%s}", fqName, strings.Join(labelArray, ","))

	metricMap.RLock()
	_, txnMapOk := metricMap.MetricMap[labelResult]
	metricMap.RUnlock()
	if txnMapOk {
		return true
	} else {
		metricMap.Lock()
		metricMap.MetricMap[labelResult] = true
		metricMap.Unlock()
	}
	return false
}
