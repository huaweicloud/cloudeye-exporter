package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	cbrmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cbr/v1/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	"github.com/stretchr/testify/assert"
)

func TestGetResourceKeyFromMetricInfo(t *testing.T) {
	metric := model.MetricInfoList{
		Dimensions: []model.MetricsDimension{
			{Name: "instance_id", Value: "0001-0001-000001"},
			{Name: "disk", Value: "vda"},
		},
	}
	assert.Equal(t, "vda.0001-0001-000001", GetResourceKeyFromMetricInfo(metric))
}

func TestGetResourceKeyFromMetricData(t *testing.T) {
	dmsNs := "SYS.DMS"
	agtNs := "AGT.ECS"
	mrsNs := "SYS.MRS"
	dcssNs := "SYS.DCS"
	testCases := []struct {
		name     string
		metric   model.BatchMetricData
		expected string
	}{
		{
			name: "dms-instance",
			metric: model.BatchMetricData{
				Namespace: &dmsNs,
				Dimensions: &[]model.MetricsDimension{
					{"kafka_instance_id", "kafka-0001-0001"},
				},
			},
			expected: "kafka-0001-0001",
		},
		{
			name: "dms-instance-broker",
			metric: model.BatchMetricData{
				Namespace: &dmsNs,
				Dimensions: &[]model.MetricsDimension{
					{"kafka_instance_id", "kafka-0001-0001"},
					{"kafka_broker", "0"},
				},
			},
			expected: "kafka-0001-0001",
		},
		{
			name: "dms-invalid",
			metric: model.BatchMetricData{
				Namespace: &dmsNs,
				Dimensions: &[]model.MetricsDimension{
					{"xxxxx_id", "kafka-0001-0001"},
					{"kafka_broker", "0"},
				},
			},
			expected: "",
		},
		{
			name: "agt-instance",
			metric: model.BatchMetricData{
				Namespace: &agtNs,
				Dimensions: &[]model.MetricsDimension{
					{"instance_id", "0001-0001-000001"},
				},
			},
			expected: "0001-0001-000001",
		},
		{
			name: "agt-instance-disk",
			metric: model.BatchMetricData{
				Namespace: &agtNs,
				Dimensions: &[]model.MetricsDimension{
					{"instance_id", "0001-0001-000001"},
					{"disk", "vda"},
				},
			},
			expected: "0001-0001-000001",
		},
		{
			name: "agt-invalid",
			metric: model.BatchMetricData{
				Namespace: &agtNs,
				Dimensions: &[]model.MetricsDimension{
					{"xxxx_id", "0001-0001-000001"},
					{"disk", "vda"},
				},
			},
			expected: "",
		},
		{
			name: "mrs-cluster",
			metric: model.BatchMetricData{
				Namespace: &mrsNs,
				Dimensions: &[]model.MetricsDimension{
					{"cluster_id", "0001-0001-000001"},
				},
			},
			expected: "0001-0001-000001",
		},
		{
			name: "mrs-cluster-service_name",
			metric: model.BatchMetricData{
				Namespace: &mrsNs,
				Dimensions: &[]model.MetricsDimension{
					{"cluster_id", "0001-0001-000001"},
					{"service_name", "YARN"},
				},
			},
			expected: "0001-0001-000001",
		},
		{
			name: "mrs-invalid",
			metric: model.BatchMetricData{
				Namespace: &mrsNs,
				Dimensions: &[]model.MetricsDimension{
					{"xxxx_id", "0001-0001-000001"},
					{"service_name", "YARN"},
				},
			},
			expected: "",
		},
		{
			name: "dcs-instance",
			metric: model.BatchMetricData{
				Namespace: &dcssNs,
				Dimensions: &[]model.MetricsDimension{
					{"dcs_instance_id", "0001-0001-000001"},
				},
			},
			expected: "0001-0001-000001",
		},
		{
			name: "dcs-instance-node",
			metric: model.BatchMetricData{
				Namespace: &dcssNs,
				Dimensions: &[]model.MetricsDimension{
					{"dcs_instance_id", "0001-0001-000001"},
					{"dcs_cluster_redis_node", "node-00001-0000001"},
				},
			},
			expected: "node-00001-0000001.0001-0001-000001",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, GetResourceKeyFromMetricData(testCase.metric))
		})
	}
}

func TestGetEndpoint(t *testing.T) {
	host = "iam.cn-test-9.myhuaweicloud.com"
	assert.Equal(t, "https://ecs.cn-test-9.myhuaweicloud.com/v2", getEndpoint("ecs", "v2"))
}

func TestGetTags(t *testing.T) {
	tags := map[string]string{
		"key_a": "value_a",
	}
	keys, values := getTags(tags)
	assert.Equal(t, "key_a", keys[0])
	assert.Equal(t, "value_a", values[0])

	tags = map[string]string{
		"1111": "11111",
	}
	keys, values = getTags(tags)
	assert.Equal(t, 0, len(keys))
	assert.Equal(t, 0, len(values))
}

func TestFmtTags(t *testing.T) {
	tag1 := []byte("test-error")
	tags := fmtTags(tag1)
	assert.Equal(t, 0, len(tags))

	valueA := "value_a"
	tagInfo := []cbrmodel.Tag{
		{Key: "key_a", Value: &valueA},
	}
	tags = fmtTags(tagInfo)
	assert.Equal(t, "value_a", tags["key_a"])
}

func TestGetDimsNameKey(t *testing.T) {
	dims := []model.MetricsDimension{
		{Name: "instance_id", Value: "0001-001-0000001"},
		{Name: "disk", Value: "vda"},
	}
	assert.Equal(t, "instance_id,disk", getDimsNameKey(dims))
}

func TestGetDimsValueKey(t *testing.T) {
	dims := []model.MetricsDimension{
		{Name: "instance_id", Value: "0001-001-0000001"},
		{Name: "disk", Value: "vda"},
	}
	assert.Equal(t, "0001-001-0000001,vda", getDimsValueKey(dims))
}

func TestBuildSingleDimensionMetrics(t *testing.T) {
	metricNames := []string{"cpu_util", "mem_util", "disk_util_inband"}
	metrics := buildSingleDimensionMetrics(metricNames, "SYS.ECS", "instance_id", "0001-0001-0000001")
	assert.Equal(t, len(metricNames), len(metrics))
}

func TestBuildDimensionMetrics(t *testing.T) {
	metricNames := []string{"cpu_util", "mem_util", "disk_util_inband"}
	dims := []model.MetricsDimension{
		{Name: "instance_id", Value: "0001-0001-0000001"},
		{Name: "disk", Value: "vda"},
	}
	metrics := buildDimensionMetrics(metricNames, "SYS.ECS", dims)
	assert.Equal(t, len(metricNames), len(metrics))
}

func TestGetDefaultString(t *testing.T) {
	assert.Equal(t, "", getDefaultString(nil))
	value := "test"
	assert.Equal(t, "test", getDefaultString(&value))
}

func TestFmtResourceProperties(t *testing.T) {
	propertiesmap := map[string]interface{}{
		"id": "000000001",
	}

	var properties struct {
		ID string `json:"id"`
	}
	err := fmtResourceProperties(propertiesmap, &properties)
	assert.Equal(t, true, err == nil)
	assert.Equal(t, "000000001", properties.ID)
}

func TestGetResourcesBaseInfoFromRMS(t *testing.T) {
	patches := gomonkey.ApplyFuncReturn(listResources, mockRmsResource(), nil)
	resource, err := getResourcesBaseInfoFromRMS("ecs", "cloudserver")
	assert.Equal(t, true, err == nil)
	assert.Equal(t, 1, len(resource))
	patches.Reset()
}

func TestGenDefaultReqDefWithOffsetAndLimit(t *testing.T) {
	type Response struct {
		HttpStatusCode int `json:"-"`
	}
	path := "/v1.0/{project_id}/streaming/jobs"
	requestDef := genDefaultReqDefWithOffsetAndLimit(path, new(Response))
	assert.Equal(t, path, requestDef.Path)
}

func TestGetHcClient(t *testing.T) {
	client := getHcClient("ces.test.huawei.com")
	assert.Equal(t, true, client != nil)
}
