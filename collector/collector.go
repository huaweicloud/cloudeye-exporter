package collector

import (
	"fmt"
	"time"
	"strconv"
	"strings"
	"encoding/json"

	"github.com/prometheus/common/log"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/openstack/ces/v1/metrics"
)


var defaultLabelsToResource = map[string]string{
  "lbaas_listener_id": "listener",
  "lb_instance_id": "lb",
  "direct_connect_id": "direct",
  "history_direct_connect_id": "history",
  "virtual_interface_id": "virtual",
}


type BaseHuaweiCloudExporter struct {
	Debug   bool
	Name    string
	Prefix  string
	Metrics map[string]*prometheus.Desc
	AllMetric *[]metrics.Metric
	Client *golangsdk.ServiceClient
}


func (exporter *BaseHuaweiCloudExporter) GetName() string {
	return fmt.Sprintf("%s_%s", exporter.Prefix, exporter.Name)
}


func (exporter *BaseHuaweiCloudExporter) AddMetric(preMetricName string, name string, labels []string, constLabels prometheus.Labels) {
	if exporter.Metrics == nil {
		exporter.Metrics = map[string]*prometheus.Desc{}
	}

	if constLabels == nil {
		constLabels = prometheus.Labels{}
	}

	newMetricName := prometheus.BuildFQName(exporter.GetName(), preMetricName, name)
	if _, ok := exporter.Metrics[newMetricName]; !ok {
		log.Infof("Adding metric: %s to exporter: %s", newMetricName, exporter.Name)
		exporter.Metrics[newMetricName] = prometheus.NewDesc(newMetricName, name, labels, constLabels)
	}
}


// Describe simply sends the two Descs in the struct to the channel.
func (exporter *BaseHuaweiCloudExporter) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range exporter.Metrics {
		ch <- metric
	}
}


func (exporter *BaseHuaweiCloudExporter) Collect(ch chan<- prometheus.Metric) {
	periodm, _ := time.ParseDuration("-1m")

	now := time.Now()
	from := strconv.FormatInt(int64(now.Add(periodm).UnixNano() / 1e6),10)
	to := strconv.FormatInt(int64(now.UnixNano() / 1e6),10)

	for _, metric := range *exporter.AllMetric {
		dimensionValues := []string{}
		preResourceName := ""

		for _, dimension := range metric.Dimensions {
			if val, ok := defaultLabelsToResource[dimension.Name]; ok {
				preResourceName = val
			}
			dimensionValues = append(dimensionValues, dimension.Value)
		}

		datapoints, err:= getMetricData(exporter.Client, &metric, &metric.Dimensions, from, to)
		if err != nil {
			continue
		}

		if exporter.Debug == true {
			fmt.Println("Get datapoints of metric begin...")
			dataJson, _ := json.MarshalIndent(*datapoints, "", " ")
			metricJson, _ := json.MarshalIndent(metric, "", " ")
			fmt.Println("The datapoints of metric are:" + string(dataJson))
			fmt.Println("The metric value is:", string(metricJson))
			fmt.Println("Get datapoints of metric end.")
		}

		var datapoint float64
		if len(*datapoints) == 0 {
			datapoint = 0
		} else if len(*datapoints) != 1 {
			datapoint = (*datapoints)[len(*datapoints) - 1].Average
		} else {
			datapoint = (*datapoints)[0].Average
		}

		newMetricName := prometheus.BuildFQName(exporter.GetName(), preResourceName, metric.MetricName)
		ch <- prometheus.MustNewConstMetric(exporter.Metrics[newMetricName],
			prometheus.GaugeValue, datapoint, dimensionValues...)
	}
}


func NewBaseHuaweiCloudExporter(prefix string, namespace string) *BaseHuaweiCloudExporter {
	name := strings.Replace(namespace, ".", "_", -1)
	name = strings.ToLower(name)
	return &BaseHuaweiCloudExporter{
		Name: name,
		Prefix: prefix,
	}
}


func GetMonitoringCollector(client *golangsdk.ServiceClient,
						     prefix string, namespace string, debug bool) (*BaseHuaweiCloudExporter, error) {
	exporter := NewBaseHuaweiCloudExporter(prefix, namespace)

	allMetrics, err := getAllMetric(client, namespace)
	if err != nil {
		return exporter, err
	}

	if len(*allMetrics) == 0 {
		return exporter, fmt.Errorf("The metric resources are not found.")
	}

	exporter.AllMetric = allMetrics
	exporter.Client = client
	exporter.Debug = debug

	if debug == true {
		metricsJson, _ := json.MarshalIndent(*allMetrics, "", " ")
		fmt.Println("The all metris are:", string(metricsJson))
	}

	for _, metric := range *allMetrics {
		labels := []string{}
		preMetricName := ""
		for _, dim := range metric.Dimensions {
			if val, ok := defaultLabelsToResource[dim.Name]; ok {
				preMetricName = val
			}
			labels = append(labels, dim.Name)
		}

		exporter.AddMetric(preMetricName, metric.MetricName, labels, nil)
	}

	return exporter, nil
}
