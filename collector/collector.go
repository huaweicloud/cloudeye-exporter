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
	txnKey       string
	MaxRoutines  int
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
		Namespaces:  namespaces,
		Prefix:      globalConfig.Global.Prefix,
		MaxRoutines: globalConfig.Global.MaxRoutines,
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
	logs.Logger.Debugf("[%s] Start getAllResource, namespace:%s", exporter.txnKey, namespace)
	allResourcesInfo, filterMetrics := exporter.getAllResource(namespace)
	logs.Logger.Debugf("[%s] End getAllResource, Total number of of resource:%d", exporter.txnKey, len(allResourcesInfo))
	allMetrics := filterMetrics
	var err error
	if len(*allMetrics) == 0 {
		logs.Logger.Debugf("[%s] Start getAllMetric", exporter.txnKey)
		allMetrics, err = getAllMetric(exporter.ClientConfig, namespace)
		if err != nil {
			logs.Logger.Errorln("[%s] Get all metrics error: %s", exporter.txnKey, err.Error())
			return
		}
		logs.Logger.Debugf("[%s] End getAllMetric, Total number of of metrics:%d", exporter.txnKey, len(*allMetrics))

		if len(*allMetrics) == 0 {
			logs.Logger.Warnf("[%s] The metric resources of service(%s) are not found.", exporter.txnKey, namespace)
			return
		}
	}

	exporter.MetricLen = len(*allMetrics)
	count := 0
	end := len(*allMetrics)
	tmpMetrics := []metricdata.Metric{}

	logs.Logger.Debugf("[%s] Start set data", exporter.txnKey)
	workChan := make(chan bool, exporter.MaxRoutines)
	finishChan := make(chan bool)
	defer close(workChan)
	defer close(finishChan)
	taskCount := 0
	for _, metric := range *allMetrics {
		count++
		tmpMetrics = append(tmpMetrics, getDataMetric(metric))
		if (0 == count%10) || (count == end) {
			taskCount++
			workChan <- true
			go func(tmpMetrics []metricdata.Metric) {
				defer func() {
					<-workChan
					finishChan <- true
				}()
				mds, err := getBatchMetricData(exporter.ClientConfig, &tmpMetrics, exporter.From, exporter.To)
				if err != nil {
					return
				}

				for _, md := range *mds {
					exporter.debugMetricInfo(md)
					datapoint, err := getDatapoint(md.Datapoints)
					if err != nil {
						logs.Logger.Warnf("[%s] Get data point error: %s, the metric is: %s", exporter.txnKey, err.Error(), md.MetricName)
						continue
					}

					labels, values, preResourceName, privateFlag := getOriginalLabelInfo(&md.Dimensions)
					if isResouceExist(&md.Dimensions, &allResourcesInfo) {
						labels = exporter.getExtensionLabels(labels, preResourceName, namespace, privateFlag)
						values = exporter.getExtensionLabelValues(values, &allResourcesInfo, getOriginalID(&md.Dimensions))
					}

					if len(labels) != len(values) {
						logs.Logger.Errorf("[%s] Inconsistent label and value: expected %d label %#v, but values got %d in %#v", exporter.txnKey,
							len(labels), labels, len(values), values)
						continue
					}

					newMetricName := prometheus.BuildFQName(GetMetricPrefixName(exporter.Prefix, namespace), preResourceName, md.MetricName)
					if err := sendMetricData(ctx, ch, prometheus.MustNewConstMetric(
						prometheus.NewDesc(newMetricName, newMetricName, labels, nil),
						prometheus.GaugeValue, datapoint, values...)); err != nil {
						logs.Logger.Errorf("[%s] Context has canceled, no need to send metric data, metric name: %s", exporter.txnKey, newMetricName)
					}
				}
			}(tmpMetrics)
			tmpMetrics = []metricdata.Metric{}
		}
	}
	for i := taskCount; i > 0; i-- {
		<-finishChan
	}
	logs.Logger.Debugf("[%s] Finished of set data", exporter.txnKey)
}

func (exporter *BaseHuaweiCloudExporter) Collect(ch chan<- prometheus.Metric) {
	periodm, err := time.ParseDuration("-10m")
	if err != nil {
		logs.Logger.Errorln("ParseDuration -10m error:", err.Error())
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	now := time.Now()
	from := strconv.FormatInt(int64(now.Add(periodm).UnixNano()/1e6), 10)
	to := strconv.FormatInt(int64(now.UnixNano()/1e6), 10)
	exporter.From = from
	exporter.To = to
	exporter.txnKey = fmt.Sprintf("%s-%s-%s", strings.Join(exporter.Namespaces, "-"), exporter.From, exporter.To)

	finishChan := make(chan bool)
	serviceTotal := 0
	logs.Logger.Debugf("[%s] Start Collect to data", exporter.txnKey)
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
			logs.Logger.Errorf("[%s] Error collecting metrics: Timeout making calls, waited for 30s without response", exporter.txnKey)
			continue
		}
	}
	logs.Logger.Debugf("[%s] End Collect to data", exporter.txnKey)
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
