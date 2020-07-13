package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/huaweicloud/golangsdk/openstack/ces/v1/metricdata"
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
	"kafka_broker":     "broker",
	"kafka_topics":     "topics",
	"kafka_partitions": "partitions",
	"kafka_groups":     "groups",
	"rabbitmq_node":    "rabbitmq_node",
}

type BaseHuaweiCloudExporter struct {
	From         string
	To           string
	Namespaces   []string
	Prefix       string
	Metrics      map[string]*prometheus.Desc
	MetricLen    int
	ClientConfig *Config
	Region       string
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

	exporter := &BaseHuaweiCloudExporter{
		Namespaces: namespaces,
		Prefix:     globalConfig.Global.Prefix,
	}

	exporter.ClientConfig = initClient(globalConfig)

	return exporter, nil
}

func GetMetricPrefixName(prefix string, namespace string) string {
	return fmt.Sprintf("%s_%s", prefix, replaceName(namespace))
}

// Describe simply sends the two Descs in the struct to the channel.
func (exporter *BaseHuaweiCloudExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

func (exporter *BaseHuaweiCloudExporter) collectMetricByNamespace(ctx context.Context, ch chan<- prometheus.Metric, namespace string) {
	defer func() {
		if err := recover(); err != nil {
			logs.Logger.Fatalln(err)
		}
	}()

	allMetrics, err := getAllMetric(exporter.ClientConfig, namespace)
	if err != nil {
		logs.Logger.Errorln("Get all metrics error: ", err.Error())
		return
	}

	if len(*allMetrics) == 0 {
		logs.Logger.Warnf("The metric resources of service(%s) are not found.", namespace)
		return
	}
	exporter.MetricLen = len(*allMetrics)

	allResoucesInfo := exporter.getAllResource(namespace)
	metricTimestamp := 0
	count := 0
	end := len(*allMetrics)
	tmpMetrics := []metricdata.Metric{}
	for _, metric := range *allMetrics {
		count++

		tmpMetrics = append(tmpMetrics, getDataMetric(metric))
		if (0 == count%10) || (count == end) {
			mds, err := getBatchMetricData(exporter.ClientConfig, &tmpMetrics, exporter.From, exporter.To)
			tmpMetrics = []metricdata.Metric{}
			if err != nil {
				continue
			}

			for _, md := range *mds {
				exporter.debugMetricInfo(md)
				datapoint, err := getDatapoint(md.Datapoints)
				if err != nil {
					logs.Logger.Warnf("Get data point error: %s, the metric is: %s", err.Error(), md.MetricName)
					continue
				}

				labels, values, preResourceName, privateFlag := getOriginalLabelInfo(&md.Dimensions)
				if isResouceExist(&md.Dimensions, &allResoucesInfo) {
					labels = exporter.getExtensionLabels(labels, preResourceName, namespace, privateFlag)
					values = exporter.getExtensionLabelValues(values, &allResoucesInfo, getOriginalID(&md.Dimensions))
				}

				if len(labels) != len(values) {
					logs.Logger.Errorf("Inconsistent label and value: expected %d label %#v, but values got %d in %#v",
						len(labels), labels, len(values), values)
					continue
				}

				newMetricName := prometheus.BuildFQName(GetMetricPrefixName(exporter.Prefix, namespace), preResourceName, md.MetricName)
				if err := sendMetricData(ctx, ch, prometheus.MustNewConstMetric(
					prometheus.NewDesc(newMetricName, newMetricName, labels, nil),
					prometheus.GaugeValue, datapoint, values...)); err != nil {
					logs.Logger.Errorf("Context has canceled, no need to send metric data, metric name: %s", newMetricName)
				}
			}

			metricTimestamp, err = getMetricDataTimestamp((*mds)[len(*mds)-1].Datapoints)
			if err != nil {
				logs.Logger.Warnln("Get metric data timestamp error: ", err.Error())
			}
		}
	}

	if metricTimestamp == 0 {
		return
	}

	to64, parseErr := strconv.ParseFloat(exporter.To, 64)
	if parseErr != nil {
		logs.Logger.Error("Parse exporter.To error: ", parseErr.Error())
	}
	stamp64 := float64(metricTimestamp)

	sub_duration := (to64 - stamp64) / 1000
	if err := sendMetricData(ctx, ch, prometheus.MustNewConstMetric(
		prometheus.NewDesc(GetMetricPrefixName(exporter.Prefix, namespace)+"_duration_seconds",
			namespace, nil, nil), prometheus.GaugeValue, sub_duration)); err != nil {
		logs.Logger.Errorf("Context has canceled, no need to send metric data, metric name: %s", GetMetricPrefixName(exporter.Prefix, namespace)+"_duration_seconds")
	}
}

func (exporter *BaseHuaweiCloudExporter) Collect(ch chan<- prometheus.Metric) {
	periodm, err := time.ParseDuration("-5m")
	if err != nil {
		logs.Logger.Errorln("ParseDuration -5m error:", err.Error())
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	now := time.Now()
	from := strconv.FormatInt(int64(now.Add(periodm).UnixNano()/1e6), 10)
	to := strconv.FormatInt(int64(now.UnixNano()/1e6), 10)
	exporter.From = from
	exporter.To = to

	finishChan := make(chan bool)
	serviceTotal := 0
	for _, namespace := range exporter.Namespaces {
		serviceTotal++
		go func(ch chan<- prometheus.Metric, namespace string) {
			exporter.collectMetricByNamespace(ctx, ch, namespace)
			finishChan <- true
		}(ch, namespace)
	}
	for i := 0; i < serviceTotal; i++ {
		select {
		case _, ok := <-finishChan:
			if ok {
				continue
			}
		case <-time.After(30 * time.Second):
			logs.Logger.Errorf("Error collecting metrics: Timeout making calls, waited for 30s without response")
			continue
		}
	}
}

func sendMetricData(ctx context.Context, ch chan<- prometheus.Metric, metric prometheus.Metric) error {
	// Check whether the Context has canceled
	select {
	case _, ok := <-ctx.Done():
		if ok {
			return ctx.Err()
		}
	default: // continue
	}
	// If no, send the metric
	ch <- metric
	return nil
}

func (exporter *BaseHuaweiCloudExporter) debugMetricInfo(md metricdata.MetricData) {
	logs.Logger.Debugln("Get datapoints of metric begin... (from):", exporter.From)
	dataJson, err := json.MarshalIndent(md.Datapoints, "", " ")
	if err != nil {
		logs.Logger.Debugln("MarshalIndent Datapoints error: ", err.Error())
	}
	metricJson, err := json.MarshalIndent(md.Dimensions, "", " ")
	if err != nil {
		logs.Logger.Debugln("MarshalIndent Dimensions error: ", err.Error())
	}
	logs.Logger.Debugln("The datapoints of metric are:" + string(dataJson))
	logs.Logger.Debugln("The metric value is:", string(metricJson))
	logs.Logger.Debugln("Get datapoints of metric end. (to):", exporter.To)
}

func isResouceExist(dims *[]metricdata.Dimension, allResouceInfo *map[string][]string) bool {
	if _, ok := (*allResouceInfo)[getOriginalID(dims)]; ok {
		return true
	}

	return false
}

func getDatapoint(datapoints []metricdata.Data) (float64, error) {
	var datapoint float64
	if len(datapoints) > 0 {
		datapoint = (datapoints)[len(datapoints)-1].Average
	} else {
		return 0, fmt.Errorf("The data point of metric are not found")
	}

	return datapoint, nil
}

func getMetricDataTimestamp(datapoints []metricdata.Data) (int, error) {
	var metricTimestamp int
	if len(datapoints) > 0 {
		metricTimestamp = (datapoints)[len(datapoints)-1].Timestamp
	} else {
		return 0, fmt.Errorf("The data point of metric are not found")
	}

	return metricTimestamp, nil
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
	allResouceInfo *map[string][]string,
	originalID string) []string {

	for lb := range *allResouceInfo {
		if lb == originalID {
			dimensionValues = append(dimensionValues, (*allResouceInfo)[lb]...)
			return dimensionValues
		}
	}

	return dimensionValues
}

func initClient(global_config *CloudConfig) *Config {
	c, err := InitConfig(global_config)
	if err != nil {
		logs.Logger.Fatalln("Init config error: ", err.Error())
	}

	return c
}
