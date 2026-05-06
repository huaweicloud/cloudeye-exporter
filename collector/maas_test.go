package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	"github.com/stretchr/testify/assert"
)

func TestMaaSInfo_GetResourceInfo(t *testing.T) {
	// 设置测试配置
	CloudConf.Global.ResourceSyncIntervalMinutes = 10
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	conf.ProjectID = "test_project_id"

	// 创建测试数据
	testApiId := "test_api_id"
	testApiName := "test_api_name"
	testKeyId := "test_key_id"
	testKeyName := "test_key_name"
	testMetricName := "test_metric"
	testNamespace := "SYS.MaaS"

	// 模拟的API响应数据
	maasResp := &ListMaasInstanceResp{
		Instances: []ListMaasInstanceRespItem{
			{
				MaasApiIdDimInfo: ListMaasInstanceRespDimInfo{
					Name: testApiName,
					Id:   testApiId,
				},
				MaasKeyIdDimInfo: ListMaasInstanceRespDimInfo{
					Name: testKeyName,
					Id:   testKeyId,
				},
			},
		},
		Total: 1,
	}

	// 模拟的指标数据
	metricInfo := model.MetricInfoList{
		Namespace:  testNamespace,
		MetricName: testMetricName,
		Dimensions: []model.MetricsDimension{
			{
				Name:  "maas_api_id",
				Value: testApiId,
			},
			{
				Name:  "maas_key_id",
				Value: testKeyId,
			},
		},
	}

	// 创建测试补丁
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock getHcClient函数
	outputs := []gomonkey.OutputCell{
		{
			Values: gomonkey.Params{
				maasResp, nil,
			},
		},
		{
			Values: gomonkey.Params{
				&ListMaasInstanceResp{Instances: []ListMaasInstanceRespItem{}}, nil,
			},
		},
		{
			Values: gomonkey.Params{
				maasResp, nil,
			},
		},
		{
			Values: gomonkey.Params{
				&ListMaasInstanceResp{Instances: []ListMaasInstanceRespItem{}}, nil,
			},
		},
	}
	hcClient := &core.HcHttpClient{}
	patches.ApplyMethodSeq(hcClient, "Sync", outputs)
	patches.ApplyFuncReturn(getHcClient, hcClient)

	// Mock getEndpoint函数
	patches.ApplyFuncReturn(getEndpoint, "https://maas.test.com/v1.0")

	// Mock listAllMetrics函数
	patches.ApplyFuncReturn(listAllMetrics, []model.MetricInfoList{metricInfo}, nil)

	// Mock IsMetricInfoInWhiteList函数
	patches.ApplyFuncReturn(IsMetricInfoInWhiteList, true)

	// 创建MaaS信息实例
	maasInfo := MaaSInfo{}

	// 调用GetResourceInfo方法
	resourceInfo, filterMetrics := maasInfo.GetResourceInfo()

	// 验证结果
	assert.NotNil(t, resourceInfo)
	assert.NotNil(t, filterMetrics)
	assert.Len(t, filterMetrics, 1)
	assert.Equal(t, testMetricName, filterMetrics[0].MetricName)

	// 验证资源信息
	expectedKey := testApiId + "." + testKeyId
	assert.Contains(t, resourceInfo, expectedKey)
	assert.Equal(t, []string{"maas_api_name", "maas_key_name"}, resourceInfo[expectedKey].Name)
	assert.Equal(t, []string{testApiName, testKeyName}, resourceInfo[expectedKey].Value)
}

func TestMaaSInfo_GetResourceInfoWithEmptyResponse(t *testing.T) {
	// 设置测试配置
	CloudConf.Global.ResourceSyncIntervalMinutes = 10
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	conf.ProjectID = "test_project_id"

	// 模拟的API响应数据（空响应）
	maasResp := &ListMaasInstanceResp{
		Instances: []ListMaasInstanceRespItem{},
		Total:     0,
	}

	// 创建测试补丁
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock getHcClient函数
	outputs := []gomonkey.OutputCell{
		{
			Values: gomonkey.Params{
				maasResp, nil,
			},
		},
		{
			Values: gomonkey.Params{
				&ListMaasInstanceResp{Instances: []ListMaasInstanceRespItem{}}, nil,
			},
		},
		{
			Values: gomonkey.Params{
				maasResp, nil,
			},
		},
		{
			Values: gomonkey.Params{
				&ListMaasInstanceResp{Instances: []ListMaasInstanceRespItem{}}, nil,
			},
		},
	}
	hcClient := &core.HcHttpClient{}
	patches.ApplyMethodSeq(hcClient, "Sync", outputs)
	patches.ApplyFuncReturn(getHcClient, hcClient)

	// Mock getEndpoint函数
	patches.ApplyFuncReturn(getEndpoint, "https://maas.test.com/v1.0")

	// Mock listAllMetrics函数
	patches.ApplyFuncReturn(listAllMetrics, []model.MetricInfoList{}, nil)

	// Mock IsMetricInfoInWhiteList函数
	patches.ApplyFuncReturn(IsMetricInfoInWhiteList, true)

	// 创建MaaS信息实例
	maasInfo := MaaSInfo{}

	// 调用GetResourceInfo方法
	resourceInfo, filterMetrics := maasInfo.GetResourceInfo()

	// 验证结果
	assert.NotNil(t, resourceInfo)
	assert.NotNil(t, filterMetrics)
	assert.Len(t, filterMetrics, 1)
}

func TestMaaSInfo_GetResourceInfoWithApiError(t *testing.T) {
	// 设置测试配置
	CloudConf.Global.ResourceSyncIntervalMinutes = 10
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	conf.ProjectID = "test_project_id"

	// 创建测试补丁
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock getHcClient函数
	patches.ApplyFuncReturn(getHcClient, &core.HcHttpClient{})

	// Mock getEndpoint函数
	patches.ApplyFuncReturn(getEndpoint, "https://maas.test.com/v1.0")

	// Mock listAllMetrics函数，返回错误
	patches.ApplyFuncReturn(listAllMetrics, []model.MetricInfoList{}, nil)

	// Mock IsMetricInfoInWhiteList函数
	patches.ApplyFuncReturn(IsMetricInfoInWhiteList, true)

	// 创建MaaS信息实例
	maasInfo := MaaSInfo{}

	// 调用GetResourceInfo方法
	resourceInfo, filterMetrics := maasInfo.GetResourceInfo()

	// 验证结果
	assert.NotNil(t, resourceInfo)
	assert.NotNil(t, filterMetrics)
}

func TestMaaSInfo_GetResourceInfoWithCache(t *testing.T) {
	// 设置测试配置
	CloudConf.Global.ResourceSyncIntervalMinutes = 10
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	conf.ProjectID = "test_project_id"

	// 创建测试数据
	testApiId := "test_api_id"
	testApiName := "test_api_name"
	testKeyId := "test_key_id"
	testKeyName := "test_key_name"
	testMetricName := "test_metric"
	testNamespace := "SYS.MaaS"

	// 模拟的API响应数据
	maasResp := &ListMaasInstanceResp{
		Instances: []ListMaasInstanceRespItem{
			{
				MaasApiIdDimInfo: ListMaasInstanceRespDimInfo{
					Name: testApiName,
					Id:   testApiId,
				},
				MaasKeyIdDimInfo: ListMaasInstanceRespDimInfo{
					Name: testKeyName,
					Id:   testKeyId,
				},
			},
		},
		Total: 1,
	}

	// 模拟的指标数据
	metricInfo := model.MetricInfoList{
		Namespace:  testNamespace,
		MetricName: testMetricName,
		Dimensions: []model.MetricsDimension{
			{
				Name:  "maas_api_id",
				Value: testApiId,
			},
			{
				Name:  "maas_key_id",
				Value: testKeyId,
			},
		},
	}

	// 创建测试补丁
	patches := gomonkey.NewPatches()
	defer patches.Reset()

	// Mock getHcClient函数
	outputs := []gomonkey.OutputCell{
		{
			Values: gomonkey.Params{
				maasResp, nil,
			},
		},
		{
			Values: gomonkey.Params{
				&ListMaasInstanceResp{Instances: []ListMaasInstanceRespItem{}}, nil,
			},
		},
		{
			Values: gomonkey.Params{
				maasResp, nil,
			},
		},
		{
			Values: gomonkey.Params{
				&ListMaasInstanceResp{Instances: []ListMaasInstanceRespItem{}}, nil,
			},
		},
	}
	hcClient := &core.HcHttpClient{}
	patches.ApplyMethodSeq(hcClient, "Sync", outputs)
	patches.ApplyFuncReturn(getHcClient, hcClient)

	// Mock getEndpoint函数
	patches.ApplyFuncReturn(getEndpoint, "https://maas.test.com/v1.0")

	// Mock listAllMetrics函数
	patches.ApplyFuncReturn(listAllMetrics, []model.MetricInfoList{metricInfo}, nil)

	// Mock IsMetricInfoInWhiteList函数
	patches.ApplyFuncReturn(IsMetricInfoInWhiteList, true)

	// 创建MaaS信息实例
	maasInfo := MaaSInfo{}

	// 第一次调用，应该会执行实际逻辑
	resourceInfo1, filterMetrics1 := maasInfo.GetResourceInfo()

	// 第二次调用，应该使用缓存
	resourceInfo2, filterMetrics2 := maasInfo.GetResourceInfo()

	// 验证结果
	assert.NotNil(t, resourceInfo1)
	assert.NotNil(t, filterMetrics1)
	assert.NotNil(t, resourceInfo2)
	assert.NotNil(t, filterMetrics2)
	assert.Equal(t, resourceInfo1, resourceInfo2)
	assert.Equal(t, filterMetrics1, filterMetrics2)
}

