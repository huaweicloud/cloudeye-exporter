package collector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/huaweicloud/golangsdk/openstack/ces/v1/metricdata"
	"github.com/huaweicloud/golangsdk/openstack/ces/v1/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

var defaultLabelsToResource = map[string]string{
	"lbaas_listener_id":         "listener",
	"lb_instance_id":            "lb",
	"direct_connect_id":         "direct",
	"history_direct_connect_id": "history",
	"virtual_interface_id":      "virtual",
	"bandwidth_id":              "bandwidth",
	"publicip_id":               "eip",
	"rabbitmq_instance_id":      "instance",
	"kafka_instance_id":         "instance",
}

var privateResourceFlag = map[string]string{
	"kafka_broker":              "broker",
	"kafka_topics":              "topics",
	"kafka_partitions":          "partitions",
	"kafka_groups":              "groups",
	"rabbitmq_node":             "rabbitmq_node",
	"rds_instance_id":           "instance",
	"postgresql_instance_id":    "instance",
	"rds_instance_sqlserver_id": "instance",
}

type BaseHuaweiCloudExporter struct {
	From            string
	To              string
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

func GetMonitoringCollector(configpath string, namespaces []string) (*BaseHuaweiCloudExporter, error) {
	globalConfig, err := NewCloudConfigFromFile(configpath)
	if err != nil {
		logs.Logger.Fatalln("NewCloudConfigFromFile error: ", err.Error())
		os.Exit(1)
	}

	client, err := InitConfig(globalConfig)
	if err != nil {
		return nil, err
	}

	exporter := &BaseHuaweiCloudExporter{
		Namespaces:      namespaces,
		Prefix:          globalConfig.Global.Prefix,
		MaxRoutines:     globalConfig.Global.MaxRoutines,
		ClientConfig:    client,
		ScrapeBatchSize: globalConfig.Global.ScrapeBatchSize,
	}
	return exporter, nil
}

func GetMetricPrefixName(prefix string, namespace string) string {
	return fmt.Sprintf("%s_%s", prefix, replaceName(namespace))
}

// Describe simply sends the two Descs in the struct to the channel.
func (exporter *BaseHuaweiCloudExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

func (exporter *BaseHuaweiCloudExporter) listMetrics(namespace string) ([]metrics.Metric, map[string][]string) {
	allResourcesInfo, metrics := exporter.getAllResource(namespace)
	logs.Logger.Debugf("[%s] Resource number of %s: %d", exporter.txnKey, namespace, len(allResourcesInfo))

	if len(*metrics) > 0 {
		return *metrics, allResourcesInfo
	}
	logs.Logger.Debugf("[%s] Start to getAllMetric from CES", exporter.txnKey)
	allMetrics, err := getAllMetric(exporter.ClientConfig, namespace)
	if err != nil {
		logs.Logger.Errorf("[%s] Get all metrics error: %s", exporter.txnKey, err.Error())
		return nil, nil
	}
	logs.Logger.Debugf("[%s] End to getAllMetric, Total number of of metrics: %d", exporter.txnKey, len(*allMetrics))
	return *allMetrics, allResourcesInfo
}

type LabelInfo struct {
	Labels          []string
	Values          []string
	PreResourceName string
}

func (exporter *BaseHuaweiCloudExporter) getLabelInfo(allResourcesInfo map[string][]string, metric metricdata.MetricData) *LabelInfo {
	labels, values, preResourceName, privateFlag := getOriginalLabelInfo(&metric.Dimensions)

	if isResourceExist(&metric.Dimensions, &allResourcesInfo) {
		labels = exporter.getExtensionLabels(labels, preResourceName, metric.Namespace, privateFlag)
		values = exporter.getExtensionLabelValues(values, &allResourcesInfo, getOriginalID(&metric.Dimensions))
	}

	if len(labels) != len(values) {
		logs.Logger.Errorf("[%s] Inconsistent label and value: expected %d label %#v, but values got %d in %#v", exporter.txnKey,
			len(labels), labels, len(values), values)
		return nil
	}
	return &LabelInfo{
		Labels:          labels,
		Values:          values,
		PreResourceName: preResourceName,
	}
}

func (exporter *BaseHuaweiCloudExporter) setProData(ctx context.Context, ch chan<- prometheus.Metric,
	dataList []metricdata.MetricData, allResourcesInfo map[string][]string) {
	for _, metric := range dataList {
		exporter.debugMetricInfo(metric)
		data, err := getLatestData(metric.Datapoints)
		if err != nil {
			logs.Logger.Warnf("[%s] Get data point error: %s, metric_name: %s, dimension: %+v", exporter.txnKey, err.Error(), metric.MetricName, metric.Dimensions)
			continue
		}

		labelInfo := exporter.getLabelInfo(allResourcesInfo, metric)
		if labelInfo == nil {
			continue
		}

		fqName := prometheus.BuildFQName(GetMetricPrefixName(exporter.Prefix, metric.Namespace), labelInfo.PreResourceName, metric.MetricName)
		proMetric := prometheus.MustNewConstMetric(
			prometheus.NewDesc(fqName, fqName, labelInfo.Labels, nil),
			prometheus.GaugeValue, data, labelInfo.Values...)
		if err := sendMetricData(ctx, ch, proMetric); err != nil {
			logs.Logger.Errorf("[%s] Context has canceled, no need to send metric data, metric name: %s", exporter.txnKey, fqName)
		}
	}
}

func (exporter *BaseHuaweiCloudExporter) collectMetricByNamespace(ctx context.Context, ch chan<- prometheus.Metric, namespace string) {
	defer func() {
		if err := recover(); err != nil {
			logs.Logger.Fatalln(err)
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
	tmpMetrics := make([]metricdata.Metric, 0, exporter.ScrapeBatchSize)

	for _, metric := range allMetrics {
		count++
		tmpMetrics = append(tmpMetrics, transMetric(metric))
		if (len(tmpMetrics) == exporter.ScrapeBatchSize) || (count == len(allMetrics)) {
			workChan <- struct{}{}
			wg.Add(1)
			go func(tmpMetrics []metricdata.Metric) {
				defer func() {
					<-workChan
					wg.Done()
				}()
				logs.Logger.Debugf("[%s] Start to getBatchMetricData, metric count: %d", exporter.txnKey, len(tmpMetrics))
				dataList, err := getBatchMetricData(exporter.ClientConfig, &tmpMetrics, exporter.From, exporter.To)
				if err != nil {
					return
				}
				exporter.setProData(ctx, ch, *dataList, allResourcesInfo)
			}(tmpMetrics)
			tmpMetrics = make([]metricdata.Metric, 0, exporter.ScrapeBatchSize)
		}
	}

	wg.Wait()
	logs.Logger.Debugf("[%s] End to scrape all metric data", exporter.txnKey)
}

func (exporter *BaseHuaweiCloudExporter) Collect(ch chan<- prometheus.Metric) {
	duration, err := time.ParseDuration("-10m")
	if err != nil {
		logs.Logger.Errorln("ParseDuration -10m error:", err.Error())
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	now := time.Now()
	exporter.From = strconv.FormatInt(now.Add(duration).UnixNano()/1e6, 10)
	exporter.To = strconv.FormatInt(now.UnixNano()/1e6, 10)
	exporter.txnKey = fmt.Sprintf("%s-%s-%s", strings.Join(exporter.Namespaces, "-"), exporter.From, exporter.To)

	logs.Logger.Debugf("[%s] Start to collect data", exporter.txnKey)
	var wg sync.WaitGroup
	for _, namespace := range exporter.Namespaces {
		wg.Add(1)
		go func(ctx context.Context, ch chan<- prometheus.Metric, namespace string) {
			defer wg.Done()
			exporter.collectMetricByNamespace(ctx, ch, namespace)
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

func (exporter *BaseHuaweiCloudExporter) debugMetricInfo(md metricdata.MetricData) {
	dataJson, err := json.Marshal(md)
	if err != nil {
		logs.Logger.Errorf("[%s] Marshal metricData error: %s", exporter.txnKey, err.Error())
		return
	}
	logs.Logger.Debugf("[%s] Get data points of metric are: %s", exporter.txnKey, string(dataJson))
}

func isResourceExist(dims *[]metricdata.Dimension, allResourceInfo *map[string][]string) bool {
	if _, ok := (*allResourceInfo)[getOriginalID(dims)]; ok {
		return true
	}

	return false
}

func getLatestData(data []metricdata.Data) (float64, error) {
	if len(data) == 0 {
		return 0, errors.New("data not found")
	}

	return data[len(data)-1].Average, nil
}

func getOriginalID(dimensions *[]metricdata.Dimension) string {
	id := ""

	if len(*dimensions) == 1 {
		id = (*dimensions)[0].Value
	} else if len(*dimensions) == 2 {
		id = (*dimensions)[1].Value
	}

	return id
}

func getOriginalLabelInfo(dims *[]metricdata.Dimension) ([]string, []string, string, string) {
	labels := []string{}
	dimensionValues := []string{}
	preResourceName := ""
	privateFlag := ""
	for _, dimension := range *dims {
		if val, ok := defaultLabelsToResource[dimension.Name]; ok {
			preResourceName = val
		}

		if val, ok := privateResourceFlag[dimension.Name]; ok {
			privateFlag = val
		}

		dimensionValues = append(dimensionValues, dimension.Value)
		if strings.ContainsAny(dimension.Name, "-") {
			labels = append(labels, strings.Replace(dimension.Name, "-", "_", -1))
			continue
		}
		labels = append(labels, dimension.Name)
	}

	return labels, dimensionValues, preResourceName, privateFlag
}

func (exporter *BaseHuaweiCloudExporter) getExtensionLabels(
	lables []string, preResourceName string, namespace string, privateFlag string) []string {

	namespace = replaceName(namespace)
	if preResourceName != "" {
		namespace = namespace + "_" + preResourceName
	}

	if privateFlag != "" {
		namespace = namespace + "_" + privateFlag
	}

	newlabels := append(lables, defaultExtensionLabels[namespace]...)

	return newlabels
}

func (exporter *BaseHuaweiCloudExporter) getExtensionLabelValues(
	dimensionValues []string,
	allResourceInfo *map[string][]string,
	originalID string) []string {

	for lb := range *allResourceInfo {
		if lb == originalID {
			dimensionValues = append(dimensionValues, (*allResourceInfo)[lb]...)
			return dimensionValues
		}
	}

	return dimensionValues
}
