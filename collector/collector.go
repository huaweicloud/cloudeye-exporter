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
		datapoint, err := getLatestData(metric.Datapoints)
		if err != nil {
			logs.Logger.Warnf("[%s] Get data point error: %s, metric_name: %s, dimension: %+v", exporter.txnKey, err.Error(), metric.MetricName, metric.Dimensions)
			continue
		}

		label := getLabel(metric, allResourcesInfo)
		label = transformUnit(&metric, label)
		fqName := prometheus.BuildFQName(exporter.Prefix, replaceName(*metric.Namespace), metric.MetricName)

		tmpMetric, err := prometheus.NewConstMetric(
			prometheus.NewDesc(fqName, fqName, label.Name, nil),
			prometheus.GaugeValue, *datapoint.Average, label.Value...)
		if err != nil {
			logs.Logger.Errorf("[%s] New const metric error: %s, fqName: %s, label: %+v",
				exporter.txnKey, err.Error(), fqName, label)
			continue
		}
		var proMetric prometheus.Metric
		if CloudConf.Global.MetricTimestampExportEnabled {
			proMetric = prometheus.NewMetricWithTimestamp(time.UnixMilli(datapoint.Timestamp), tmpMetric)
		} else {
			proMetric = tmpMetric
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

func transformUnit(md *model.BatchMetricData, label labelInfo) labelInfo {
	if !CloudConf.Global.UnitStandardizationEnabled {
		// 指标单位标准化转换未开启，使用服务上报指标中的单位并退出
		label.Name = append(label.Name, "unit")
		label.Value = append(label.Value, *md.Unit)
		return label
	}

	i18nUnit := getI18NUnit(md)
	// md.MetricName表示特定指标单位转换/补齐操作
	// *表示某类同单位指标的转换操作
	metricNames := []string{md.MetricName, "*"}
	for _, metricName := range metricNames {
		if metricName == "*" && i18nUnit == "" {
			continue
		}
		// 标准转换配置中寻找单位转换配置/单位补齐配置
		standardUnitKey := fmt.Sprintf(keyPatternForUnitStandardization, *md.Namespace, metricName, i18nUnit)
		standardUnitConf, ok := unitTransformConfMap[standardUnitKey]
		if ok {
			label.Name = append(label.Name, "unit", "unit_v2")
			label.Value = append(label.Value, i18nUnit, standardUnitConf.UnitV2)
			return label
		}
	}

	// 最终兜底：服务上报指标数据中的单位
	resultUnit := *md.Unit
	if i18nUnit != "" {
		// 标准化转换配置中未找到，再优先使用I18N配置给出的统一单位
		resultUnit = i18nUnit
	}

	label.Name = append(label.Name, "unit")
	label.Value = append(label.Value, resultUnit)
	return label
}

func getI18NUnit(md *model.BatchMetricData) string {
	unitInI18N, ok := i18nConfigMap[fmt.Sprintf(keyPatternForUnitUnifyI18n, *md.Namespace, md.MetricName)]
	if !ok {
		unitInI18N = ""
	}
	return unitInI18N
}

func isAgentMetric(namespace string) bool {
	return namespace == "AGT.ECS" || namespace == "SERVICE.BMS"
}

func getLabel(metric model.BatchMetricData, info map[string]labelInfo) labelInfo {
	label := getDimLabel(metric)
	if extendLabel, exist := info[GetResourceKeyFromMetricData(metric)]; exist {
		label.Name = append(label.Name, extendLabel.Name...)
		label.Value = append(label.Value, extendLabel.Value...)
		for idx, labelName := range extendLabel.Name {
			if labelName == "epId" && GetEpNameByEpId(extendLabel.Value[idx]) != "" {
				label.Name = append(label.Name, "epName")
				label.Value = append(label.Value, GetEpNameByEpId(extendLabel.Value[idx]))
			}
		}
	}
	return label
}

func getDimLabel(metric model.BatchMetricData) labelInfo {
	var label labelInfo
	for _, dim := range *metric.Dimensions {
		label.Name = append(label.Name, strings.ReplaceAll(dim.Name, "-", "_"))
		label.Value = append(label.Value, getDimValue(metric, dim.Name, dim.Value))
	}
	// deepseek用户要求针对agent指标的磁盘维度拼接evsId字段
	if *metric.Namespace == "AGT.ECS" {
		getEvsInfoForECS(metric, &label)
	}
	if *metric.Namespace == "SERVICE.BMS" {
		getEvsInfoForBMS(metric, &label)
	}
	return label
}

func getDimValue(metricData model.BatchMetricData, dimName, dimValue string) string {
	if !isContainsInStringArr(*metricData.Namespace, []string{"AGT.ECS", "SERVICE.BMS"}) {
		return dimValue
	}

	if !isContainsInStringArr(dimName, []string{"mount_point", "disk", "proc", "gpu", "raid"}) {
		return dimValue
	}

	instanceID := getServerResourceKey(metricData)
	if instanceID == "" {
		return dimValue
	}

	return getAgentOriginValue(instanceID, dimValue)
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
					logs.Logger.Errorf("[%s] Get batch metric data param: %+v", exporter.txnKey, tmpMetrics)
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
	queryDuration := fmt.Sprintf("-%dm", CloudConf.Global.MetricQueryDuration)
	duration, err := time.ParseDuration(queryDuration)
	if err != nil {
		logs.Logger.Errorf("ParseDuration %s error:", queryDuration, err.Error())
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

func getLatestData(data []model.DatapointForBatchMetric) (*model.DatapointForBatchMetric, error) {
	if len(data) == 0 {
		return nil, errors.New("data not found")
	}

	return &data[len(data)-1], nil
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

func getEvsInfoForECS(metric model.BatchMetricData, label *labelInfo) {
	getEvsInfo(metric, label, &ecsInfo)
}

func getEvsInfoForBMS(metric model.BatchMetricData, label *labelInfo) {
	getEvsInfo(metric, label, &bmsInfo)
}

func getEvsInfo(metric model.BatchMetricData, label *labelInfo, serverInfo *serversInfo) {
	instanceID := ""
	diskName := ""

	for _, dim := range *metric.Dimensions {

		if dim.Name == "instance_id" {
			instanceID = dim.Value
		}
		if dim.Name == "disk" {
			diskName = getDimValue(metric, dim.Name, dim.Value)
		}
	}

	if diskName == "" {
		return
	}

	extendInfoMap, ok := serverInfo.ExtendInfo[instanceID]
	if !ok {
		logs.Logger.Warnf("Evs info not found, instanceID is %s", instanceID)
		return
	}

	extendMap, mapOk := extendInfoMap.(map[string]string)
	if !mapOk {
		logs.Logger.Errorf("Convert map failed, instanceID is %s", instanceID)
		return
	}

	evsID, idOk := extendMap[diskName]
	if !idOk {
		label.Name = append(label.Name, "evsId")
		label.Value = append(label.Value, "")
		return
	}
	label.Name = append(label.Name, "evsId")
	label.Value = append(label.Value, evsID)
}
