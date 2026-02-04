package collector

import (
	"sort"
	"testing"
	"time"

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

	endpointConfig = map[string]string{
		"ecs": "https://ecs.cn-test-9.myhuaweicloud.com/v2",
	}
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
		{Key: "key_a", Value: valueA},
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

func TestGetResourceInfoExpirationTime(t *testing.T) {
	CloudConf.Global.ResourceSyncIntervalMinutes = 180
	expirationTime := GetResourceInfoExpirationTime()
	assert.Equal(t, 180*time.Minute, expirationTime)

	CloudConf.Global.ResourceSyncIntervalMinutes = 5
	expirationTime = GetResourceInfoExpirationTime()
	assert.Equal(t, 10*time.Minute, expirationTime)
}

func TestContainsInArray(t *testing.T) {
	testArray := []string{"a", "c", "b"}
	sort.Strings(testArray)
	assert.True(t, ContainsInArray(testArray, "b"))
	assert.False(t, ContainsInArray(testArray, "g"))
}

func TestDimNameEquals(t *testing.T) {
	dimName := "task_id,operator,city"
	sortedDimName := "city,operator,task_id"
	wrongDimName := "city,task_id"
	assert.True(t, DimNameEquals(dimName, sortedDimName))
	assert.True(t, DimNameEquals(dimName, dimName))
	assert.False(t, DimNameEquals(dimName, wrongDimName))
	dimName = "task_id,"
	sortedDimName = "city"
	assert.False(t, DimNameEquals(dimName, sortedDimName))
}

func TestStrSliceContains(t *testing.T) {
	testCases := []struct {
		name   string
		str    string
		slice  []string
		expect func(t *testing.T, result bool)
	}{
		{
			"return_true",
			"aaa",
			[]string{"aaa", "bbb"},
			func(t *testing.T, result bool) {
				assert.True(t, result)
			},
		},
		{
			"return_false",
			"aaa",
			[]string{"bbb", "ccc"},
			func(t *testing.T, result bool) {
				assert.False(t, result)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := strSliceContains(testCase.slice, testCase.str)
			testCase.expect(t, result)
		})
	}
}

func TestGetServerResourceKeyFromMetricInfo(t *testing.T) {
	testCases := []struct {
		name           string
		metricInfoList model.MetricInfoList
		expect         func(t *testing.T, resourceKey string)
	}{
		{
			"no_instance_id_dim_name",
			model.MetricInfoList{
				Namespace:  namespace,
				MetricName: "cpu_usage",
				Dimensions: []model.MetricsDimension{
					{
						Name:  "instance_id1",
						Value: "9234ad9f-87a5-49a9-b6ec-11ecde1d943b",
					},
				},
			},
			func(t *testing.T, resourceKey string) {
				assert.Equal(t, "", resourceKey)
			},
		},
		{
			"has_instance_id_dim_name",
			model.MetricInfoList{
				Namespace:  namespace,
				MetricName: "cpu_usage",
				Dimensions: []model.MetricsDimension{
					{
						Name:  "instance_id",
						Value: "9234ad9f-87a5-49a9-b6ec-11ecde1d943b",
					},
				},
			},
			func(t *testing.T, resourceKey string) {
				assert.NotEqual(t, "", resourceKey)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resourceKey := getServerResourceKeyFromMetricInfo(testCase.metricInfoList)
			testCase.expect(t, resourceKey)
		})
	}
}

func TestCleanMergedMetrics(t *testing.T) {
	// 创建测试用的指标数据
	now := time.Now().Unix()

	// 创建一些测试指标
	metric1 := model.MetricInfoList{
		Namespace:  "SYS.ECS",
		MetricName: "cpu_util",
		Dimensions: []model.MetricsDimension{
			{Name: "instance_id", Value: "instance-1"},
		},
	}

	metric2 := model.MetricInfoList{
		Namespace:  "SYS.ECS",
		MetricName: "mem_util",
		Dimensions: []model.MetricsDimension{
			{Name: "instance_id", Value: "instance-2"},
		},
	}

	metric3 := model.MetricInfoList{
		Namespace:  "SYS.ECS",
		MetricName: "disk_util",
		Dimensions: []model.MetricsDimension{
			{Name: "instance_id", Value: "instance-3"},
		},
	}

	// 构建测试用的mergedMetricMap
	mergedMetricMap := map[string]MetricInfoListWithTTL{
		"instance-1.cpu_util": {
			TTL:            now + 3600, // 1小时后过期
			MetricInfoList: metric1,
		},
		"instance-2.mem_util": {
			TTL:            now - 3600, // 1小时前过期
			MetricInfoList: metric2,
		},
		"instance-3.disk_util": {
			TTL:            now + 7200, // 2小时后过期
			MetricInfoList: metric3,
		},
	}

	// 调用被测试函数
	result := cleanMergedMetrics(mergedMetricMap)

	// 验证结果：应该只包含未过期的指标
	assert.Equal(t, 2, len(result))

	// 验证返回的指标是否正确
	resultMap := make(map[string]bool)
	for _, metric := range result {
		key := getMetricKeyFromMetricInfo(metric.MetricInfoList)
		resultMap[key] = true
	}

	// 应该包含未过期的指标
	assert.True(t, resultMap["instance-1.cpu_util"])
	assert.True(t, resultMap["instance-3.disk_util"])

	// 不应该包含已过期的指标
	assert.False(t, resultMap["instance-2.mem_util"])
}
