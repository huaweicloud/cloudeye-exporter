// Copyright 2019 HuaweiCloud.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/huaweicloud/golangsdk/openstack/ces/v1/metricdata"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

var defaultLabelsToResource = map[string]string{
	"lbaas_listener_id":         "listener",
	"lb_instance_id":            "lb",
	"direct_connect_id":         "direct",
	"history_direct_connect_id": "history",
	"virtual_interface_id":      "virtual",
	"bandwidth_id":              "bandwidth",
	"publicip_id":               "eip",
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
	Debug        bool
	Namespaces   []string
	Prefix       string
	Metrics      map[string]*prometheus.Desc
	ClientConfig *Config
	Region       string
}

func replaceName(name string) string {
	newName := strings.Replace(name, ".", "_", -1)
	newName = strings.ToLower(newName)

	return newName
}

func GetMonitoringCollector(configpath string, namespaces []string, debug bool) (*BaseHuaweiCloudExporter, error) {
	global_config, err := NewCloudConfigFromFile(configpath)
	if err != nil {
		log.Fatal(err)
	}

	exporter := &BaseHuaweiCloudExporter{
		Namespaces: namespaces,
		Prefix:     global_config.Global.Prefix,
		Debug:      debug,
	}

	exporter.ClientConfig = initClient(global_config)

	return exporter, nil
}

func GetMetricPrefixName(prefix string, namespace string) string {
	return fmt.Sprintf("%s_%s", prefix, replaceName(namespace))
}

// Describe simply sends the two Descs in the struct to the channel.
func (exporter *BaseHuaweiCloudExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

func (exporter *BaseHuaweiCloudExporter) collectMetricByNamespace(ch chan<- prometheus.Metric, namespace string) {
	allMetrics, err := getAllMetric(exporter.ClientConfig, namespace)
	if err != nil {
		log.Fatal(err)
	}

	if len(*allMetrics) == 0 {
		log.Warnf("The metric resources of service(%s) are not found.", namespace)
		return
	}

	if exporter.Debug == true {
		metricsJSON, _ := json.MarshalIndent(*allMetrics, "", " ")
		log.Debug("The all metrics are:", string(metricsJSON))
	}

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
					log.Warnf("%s, the metric is: %s", err, md.MetricName)
					continue
				}

				labels, values, preResourceName, privateFlag := getOriginalLabelInfo(&md.Dimensions)
				if isResouceExist(&md.Dimensions, &allResoucesInfo) == true {
					labels = exporter.getExtensionLabels(labels, preResourceName, namespace, privateFlag)
					values = exporter.getExtensionLabelValues(values, &allResoucesInfo, getOriginalID(&md.Dimensions))
				}

				newMetricName := prometheus.BuildFQName(GetMetricPrefixName(exporter.Prefix, namespace), preResourceName, md.MetricName)
				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(newMetricName, newMetricName, labels, nil),
					prometheus.GaugeValue, datapoint, values...)
			}

			metricTimestamp, _ = getMetricDataTimestamp((*mds)[len(*mds)-1].Datapoints)
		}
	}

	if metricTimestamp == 0 {
		return
	}

	to64, _ := strconv.ParseFloat(exporter.To, 64)
	stamp64 := float64(metricTimestamp)

	subDuration := (to64 - stamp64) / 1000
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(GetMetricPrefixName(exporter.Prefix, namespace)+"_duration_seconds",
			namespace, nil, nil), prometheus.GaugeValue, subDuration)
}

func (exporter *BaseHuaweiCloudExporter) Collect(ch chan<- prometheus.Metric) {
	periodm, _ := time.ParseDuration("-5m")

	now := time.Now()
	from := strconv.FormatInt(int64(now.Add(periodm).UnixNano()/1e6), 10)
	to := strconv.FormatInt(int64(now.UnixNano()/1e6), 10)
	exporter.From = from
	exporter.To = to

	for _, namespace := range exporter.Namespaces {
		exporter.collectMetricByNamespace(ch, namespace)
	}
}

func (exporter *BaseHuaweiCloudExporter) debugMetricInfo(md metricdata.MetricData) {
	if exporter.Debug == true {
		log.Debug("Get datapoints of metric begin... (from):", exporter.From)
		dataJSON, _ := json.MarshalIndent(md.Datapoints, "", " ")
		metricJSON, _ := json.MarshalIndent(md.Dimensions, "", " ")
		log.Debug("The datapoints of metric are:" + string(dataJSON))
		log.Debug("The metric value is:", string(metricJSON))
		log.Debug("Get datapoints of metric end. (to):", exporter.To)
	}
}

func isResouceExist(dims *[]metricdata.Dimension, allResouceInfo *map[string][]string) bool {
	for _, dimension := range *dims {
		if _, ok := (*allResouceInfo)[dimension.Value]; ok {
			return true
		}
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
		log.Fatal(err)
	}

	return c
}
