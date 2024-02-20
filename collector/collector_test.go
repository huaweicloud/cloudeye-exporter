package collector

import (
	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/cloudeye-exporter/logs"
	"go.uber.org/zap/zapcore"
	"sync"
	"testing"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	"github.com/stretchr/testify/assert"
)

func TestReplaceName(t *testing.T) {
	assert.Equal(t, "sys_ecs", replaceName("SYS.ECS"))
}

func TestGetLabel(t *testing.T) {
	ns := "SYS.ECS"
	avg := 23.5
	unit := "%"
	metric := model.BatchMetricData{
		Namespace:  &ns,
		MetricName: "cpu_utils",
		Dimensions: &[]model.MetricsDimension{
			{Name: "instance_id", Value: "0001-0001-0000001"},
		},
		Datapoints: []model.DatapointForBatchMetric{
			{Average: &avg, Timestamp: time.Now().Unix() * 1000},
		},
		Unit: &unit,
	}
	info := map[string]labelInfo{
		"0001-0001-0000001": {Name: []string{"name"}, Value: []string{"host01"}},
	}
	label := getLabel(metric, info)
	assert.Equal(t, 3, len(label.Name))
	assert.Equal(t, 3, len(label.Value))
}

func TestGetDimValue(t *testing.T) {
	assert.Equal(t, "0001-0001-0000001", getDimValue("SYS.ECS", "instance_id", "0001-0001-0000001"))
	assert.Equal(t, "0001-0001-0000001", getDimValue("AGT.ECS", "instance_id", "0001-0001-0000001"))
	assert.Equal(t, "000000000000000", getDimValue("AGT.ECS", "disk", "000000000000000"))
	agentDimensions.Store("000000000000000", "vda")
	assert.Equal(t, "vda", getDimValue("AGT.ECS", "disk", "000000000000000"))
}

func TestTransMetric(t *testing.T) {
	metricInfoList := model.MetricInfoList{
		Namespace: "SYS.ECS",
	}
	assert.Equal(t, "SYS.ECS", transMetric(metricInfoList).Namespace)
}

func TestGetLatestData(t *testing.T) {
	value, err := getLatestData([]model.DatapointForBatchMetric{})
	assert.Equal(t, true, value == 0)
	assert.Equal(t, "data not found", err.Error())
	avg1, avg2 := 23.5, 63.52
	data := []model.DatapointForBatchMetric{
		{Average: &avg1, Timestamp: time.Now().Unix()*1000 - 1000*60},
		{Average: &avg2, Timestamp: time.Now().Unix() * 1000},
	}
	value, err = getLatestData(data)
	assert.Equal(t, true, err == nil)
	assert.Equal(t, avg2, value)
}

func TestIsMetricLabelConflict(t *testing.T) {
	var label = labelInfo{
		Name:  []string{"disk", "instance_id", "unit"},
		Value: []string{"0eb113c0-0111-4f51-9d44-f2157fcc138c", "2e84018fc8b4484b94e89aae212fe615", "%"},
	}
	proMap := PrometheusMetricMap{
		RWMutex:   sync.RWMutex{},
		MetricMap: make(map[string]bool),
	}
	proMap.MetricMap["huaweicloud_agt_ecs_cpu_usage{instance_id=9234ad9f-87a5-49a9-b6ec-11ecde1d943b,mount_point=2e84018fc8b4484b94e89aae212fe615,unit=%}"] = true
	result1 := isMetricLabelConflict("huaweicloud_agt_ecs_cpu_usage", label, &proMap)
	assert.False(t, result1)

	proMap.MetricMap["huaweicloud_agt_ecs_cpu_usage{instance_id=9234ad9f-87a5-49a9-b6ec-11ecde1d943b,disk=2e84018fc8b4484b94e89aae212fe615,unit=%}"] = true
	result2 := isMetricLabelConflict("huaweicloud_agt_ecs_cpu_usage", label, &proMap)
	assert.True(t, result2)
}

func TestIsAgentMetric(t *testing.T) {
	resultECS := isAgentMetric("AGT.ECS")
	assert.True(t, resultECS)

	resultBMS := isAgentMetric("SERVICE.BMS")
	assert.True(t, resultBMS)
}

func TestSetProData1(t *testing.T) {
	proMap := PrometheusMetricMap{
		RWMutex:   sync.RWMutex{},
		MetricMap: make(map[string]bool),
	}
	proMap.MetricMap["agt_ecs_cpu_usage{instance_id=9234ad9f-87a5-49a9-b6ec-11ecde1d943b,mount_point=2e84018fc8b4484b94e89aae212fe615,unit=%}"] = true

	unit := "%"
	namespace := "AGT.ECS"
	avg := 0.33
	data := model.BatchMetricData{
		Unit: &unit,
		Datapoints: []model.DatapointForBatchMetric{
			{
				Average: &avg,
			},
		},
		Namespace:  &namespace,
		MetricName: "cpu_usage",
		Dimensions: &[]model.MetricsDimension{
			{
				Name:  "instance_id",
				Value: "9234ad9f-87a5-49a9-b6ec-11ecde1d943b",
			},
			{
				Name:  "mount_point",
				Value: "2e84018fc8b4484b94e89aae212fe615",
			},
		},
	}
	metricDataArray := []model.BatchMetricData{data}
	var exporter BaseHuaweiCloudExporter
	resourceInfo := make(map[string]labelInfo, 0)
	var label = labelInfo{
		Name:  []string{"mount_point", "instance_id", "unit"},
		Value: []string{"2e84018fc8b4484b94e89aae212fe615", "0eb113c0-0111-4f51-9d44-f2157fcc138c", "%"},
	}
	resourceInfo["instance_id"] = label
	confLoader := logs.ConfLoader{}
	patches := gomonkey.ApplyMethodFunc(confLoader, "LoadFile", func(fPath string, cfg interface{}) error {
		cfgTmp, ok := cfg.(*map[string][]logs.Config)
		assert.True(t, ok)
		cfgPointer := make(map[string][]logs.Config)
		cfgPointer["business"] = []logs.Config{
			{
				Level: zapcore.InfoLevel,
			},
		}
		*cfgTmp = cfgPointer
		return nil
	})
	defer patches.Reset()
	logs.InitLog()
	exporter.Collect(nil)
	exporter.setProData(nil, nil, metricDataArray, resourceInfo, &proMap)
	assert.Equal(t, 3, len(label.Name))
}
