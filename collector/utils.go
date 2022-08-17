package collector

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

const TTL = time.Hour * 3

var tagRegexp *regexp.Regexp

func init() {
	var err error
	tagRegexp, err = regexp.Compile("^[A-Za-z_]+$")
	if err != nil {
		logs.Logger.Error("init tag regexp error: %s", err.Error())
	}
}

type serversInfo struct {
	TTL           int64
	LabelInfo     map[string]labelInfo
	FilterMetrics []model.MetricInfoList
	sync.Mutex
}

type labelInfo struct {
	Name  []string
	Value []string
}

type RmsInfo struct {
	Id   string
	Name string
	EpId string
	Tags map[string]string
}

func GetResourceKeyFromMetricInfo(metric model.MetricInfoList) string {
	sort.Slice(metric.Dimensions, func(i, j int) bool {
		return metric.Dimensions[i].Name < metric.Dimensions[j].Name
	})
	dimValuesList := make([]string, 0, len(metric.Dimensions))
	for _, dim := range metric.Dimensions {
		dimValuesList = append(dimValuesList, dim.Value)
	}
	return strings.Join(dimValuesList, ".")
}

func GetResourceKeyFromMetricData(metric model.BatchMetricData) string {
	sort.Slice(*metric.Dimensions, func(i, j int) bool {
		return (*metric.Dimensions)[i].Name < (*metric.Dimensions)[j].Name
	})
	dimValuesList := make([]string, 0, len(*metric.Dimensions))
	for _, dim := range *metric.Dimensions {
		dimValuesList = append(dimValuesList, dim.Value)
	}
	return strings.Join(dimValuesList, ".")
}

func getEndpoint(server, version string) string {
	return fmt.Sprintf("https://%s/%s", strings.Replace(host, "iam", server, 1), version)
}

// 标签只允许大写字母，小写字母和下划线，过滤tags中有效的tag
func getTags(tags map[string]string) ([]string, []string) {
	var keys, values []string
	for key, value := range tags {
		valid := tagRegexp.MatchString(key)
		if !valid {
			continue
		}
		keys = append(keys, key)
		values = append(values, value)
	}
	return keys, values
}

func fmtTags(tagInfo interface{}) map[string]string {
	bytes, err := json.Marshal(tagInfo)
	if err != nil {
		return nil
	}
	var tags []struct {
		Key   string `json:"key,omitempty"`
		Value string `json:"value,omitempty"`
	}
	err = json.Unmarshal(bytes, &tags)
	if err != nil {
		return nil
	}
	tagMap := make(map[string]string)
	for _, tag := range tags {
		tagMap[tag.Key] = tag.Value
	}
	return tagMap
}

type ResourceBaseInfo struct {
	ID   string
	Name string
	EpId string
	Tags map[string]string
}

func getDimsNameKey(dims []model.MetricsDimension) string {
	dimsNamesList := make([]string, 0, len(dims))
	for _, dim := range dims {
		dimsNamesList = append(dimsNamesList, dim.Name)
	}
	return strings.Join(dimsNamesList, ",")
}

func buildSingleDimensionMetrics(metricNames []string, namespace, dimName, dimValue string) []model.MetricInfoList {
	filterMetrics := make([]model.MetricInfoList, len(metricNames))
	for index := range metricNames {
		filterMetrics[index] = model.MetricInfoList{
			Namespace:  namespace,
			MetricName: metricNames[index],
			Dimensions: []model.MetricsDimension{
				{
					Name:  dimName,
					Value: dimValue,
				},
			},
		}
	}
	return filterMetrics
}

func buildDimensionMetrics(metricNames []string, namespace string, dimensions []model.MetricsDimension) []model.MetricInfoList {
	filterMetrics := make([]model.MetricInfoList, len(metricNames))
	for index := range metricNames {
		filterMetrics[index] = model.MetricInfoList{
			Namespace:  namespace,
			MetricName: metricNames[index],
			Dimensions: dimensions,
		}
	}
	return filterMetrics
}
