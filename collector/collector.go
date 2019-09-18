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
	"fmt"
	"time"
	"strconv"
	"strings"
	"encoding/json"

	"github.com/prometheus/common/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/huaweicloud/golangsdk/openstack/ces/v1/metrics"
)


var defaultLabelsToResource = map[string]string{
  "lbaas_listener_id": "listener",
  "lb_instance_id": "lb",
  "direct_connect_id": "direct",
  "history_direct_connect_id": "history",
  "virtual_interface_id": "virtual",
  "bandwidth_id": "bandwidth",
  "publicip_id": "eip",
}

var privateResourceFlag = map[string]string {
	"kafka_broker": "broker",
	"kafka_topics": "topics",
	"kafka_partitions": "partitions",
	"kafka_groups": "groups",
	"rabbitmq_node": "rabbitmq_node",
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


func replaceName(name string) (string) {
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


func (exporter *BaseHuaweiCloudExporter) collectMetricByNamespace(ch chan<- prometheus.Metric, namespace string)  {
	allMetrics, err := getAllMetric(exporter.ClientConfig, namespace)
	if err != nil {
		log.Fatal(err)
	}

	if len(*allMetrics) == 0 {
		log.Warnf("The metric resources of service(%s) are not found.", namespace)
	}

	if exporter.Debug == true {
		metricsJson, _ := json.MarshalIndent(*allMetrics, "", " ")
		fmt.Println("The all metrics are:", string(metricsJson))
	}

	allResoucesInfo := exporter.getAllResource(namespace)

	var metricTimestamp int
	for _, metric := range *allMetrics {
		dimensionValues := []string{}
		labels := []string{}
		preResourceName := ""
		privateFlag := ""

		for _, dimension := range metric.Dimensions {
			if val, ok := defaultLabelsToResource[dimension.Name]; ok {
				preResourceName = val
			}

			if val, ok := privateResourceFlag[dimension.Name]; ok {
				privateFlag = val
			}

			dimensionValues = append(dimensionValues, dimension.Value)
			labels = append(labels, dimension.Name)
		}

		datapoints, err:= getMetricData(exporter.ClientConfig, &metric, &metric.Dimensions, exporter.From, exporter.To)
		if err != nil {
			continue
		}

		if exporter.Debug == true {
			fmt.Println("Get datapoints of metric begin... (from):", exporter.From)
			dataJson, _ := json.MarshalIndent(*datapoints, "", " ")
			metricJson, _ := json.MarshalIndent(metric, "", " ")
			fmt.Println("The datapoints of metric are:" + string(dataJson))
			fmt.Println("The metric value is:", string(metricJson))
			fmt.Println("Get datapoints of metric end. (to):", exporter.To)
		}

		var datapoint float64
		if len(*datapoints) > 0 {
			datapoint = (*datapoints)[len(*datapoints) - 1].Average
			metricTimestamp = (*datapoints)[len(*datapoints) - 1].Timestamp
		} else {
			fmt.Println("The data point of metric are not found, the metric is:", metric.MetricName)
			metricJson, _ := json.MarshalIndent(metric, "", " ")
			fmt.Println("The metric value is:", string(metricJson))
			continue
		}

		labels = exporter.setExtensionLabels(labels, preResourceName, namespace, privateFlag)
		dimensionValues = exporter.setExtensionLabelValues(dimensionValues, &allResoucesInfo, getOriginalID(&metric.Dimensions))

		newMetricName := prometheus.BuildFQName(GetMetricPrefixName(exporter.Prefix, namespace), preResourceName, metric.MetricName)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(newMetricName, newMetricName, labels, nil),
			prometheus.GaugeValue, datapoint, dimensionValues...)
	}

	to64, _ := strconv.ParseFloat(exporter.To, 64)
	stamp64 := float64(metricTimestamp)

	sub_duration := (to64 - stamp64) / 1000
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(GetMetricPrefixName(exporter.Prefix, namespace) + "_duration_seconds",
			namespace, nil, nil), prometheus.GaugeValue, sub_duration)
}


func (exporter *BaseHuaweiCloudExporter) Collect(ch chan<- prometheus.Metric) {
	periodm, _ := time.ParseDuration("-5m")

	now := time.Now()
	from := strconv.FormatInt(int64(now.Add(periodm).UnixNano() / 1e6),10)
	to := strconv.FormatInt(int64(now.UnixNano() / 1e6),10)
	exporter.From = from
	exporter.To = to

	for _, namespace := range exporter.Namespaces {
		exporter.collectMetricByNamespace(ch, namespace)
	}
}

func getOriginalID(dimensions *[]metrics.Dimension) (string) {
	id := ""

	if len(*dimensions) == 1 {
		id = (*dimensions)[0].Value
	} else if len(*dimensions) == 2 {
		id = (*dimensions)[1].Value
	}

	return id
}


func (exporter *BaseHuaweiCloudExporter) setExtensionLabels(
	lables []string, preResourceName string, namespace string, privateFlag string) ([]string) {

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


func (exporter *BaseHuaweiCloudExporter) setExtensionLabelValues(
	dimensionValues []string,
	allResouceInfo *map[string][]string,
	originalID string) ([]string) {

	for lb := range *allResouceInfo{
		if lb == originalID {
			dimensionValues = append(dimensionValues, (*allResouceInfo)[lb]...)
			return dimensionValues
		}
	}

	return dimensionValues
}


func initClient(global_config *CloudConfig) (*Config)  {
	c, err := InitConfig(global_config)
	if err != nil {
		log.Fatal(err)
	}

	return c
}
