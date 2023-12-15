package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	"github.com/stretchr/testify/assert"
)

func TestListAllResources(t *testing.T) {
	exporter := GetMonitoringCollector([]string{"SYS.ECS"})
	lables, metrics := exporter.listAllResources("TEST.ECS")
	assert.Equal(t, 0, len(lables))
	assert.Equal(t, 0, len(metrics))

	ecsInfo := ECSInfo{}
	serviceMap["SYS.ECS"] = ecsInfo
	lables = map[string]labelInfo{
		"0001-0001-000000001": {
			Name:  []string{"name"},
			Value: []string{"host01"},
		},
	}
	metrics = []model.MetricInfoList{
		{
			Dimensions: []model.MetricsDimension{
				{Name: "instance_id", Value: "0001-0001-000000001"},
			},
			MetricName: "cpu_utils",
			Namespace:  "SYS.ECS",
		},
	}
	patches := gomonkey.ApplyMethodReturn(ecsInfo, "GetResourceInfo", lables, metrics)
	l, m := exporter.listAllResources("SYS.ECS")
	assert.Equal(t, 1, len(l))
	assert.Equal(t, 1, len(m))
	patches.Reset()
}
