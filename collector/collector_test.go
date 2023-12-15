package collector

import (
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
