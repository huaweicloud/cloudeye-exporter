package collector

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

func TestCFWInfo_GetResourceInfo_configIsNil(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()
	logs.InitLog()
	cfwInfoTest := CFWInfo{}
	labelInfos, filterMetrics := cfwInfoTest.GetResourceInfo()
	assert.Nil(t, labelInfos)
	assert.Nil(t, filterMetrics)
}

func TestCFWInfo_GetResourceInfo_dimConfigIsNotExists(t *testing.T) {
	metricConfigMap := map[string][]string{}
	patches := getPatches()
	defer patches.Reset()
	patches.ApplyFuncReturn(getMetricConfigMap, metricConfigMap)

	logs.InitLog()
	cfwInfoTest := CFWInfo{}
	labelInfos, filterMetrics := cfwInfoTest.GetResourceInfo()
	assert.Nil(t, labelInfos)
	assert.Nil(t, filterMetrics)
}

func TestCFWInfo_GetResourceInfo_dimConfigIsEmpty(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	metricConfigMap := map[string][]string{
		"fw_instance_id": nil,
	}
	patches.ApplyFuncReturn(getMetricConfigMap, metricConfigMap)
	logs.InitLog()
	cfwInfoTest := CFWInfo{}
	labelInfos, filterMetrics := cfwInfoTest.GetResourceInfo()
	assert.Nil(t, labelInfos)
	assert.Nil(t, filterMetrics)
}

func TestCFWInfo_GetResourceInfo_getResourcesFromRMSFailed(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	metricConfigMap := map[string][]string{
		"fw_instance_id": {"metric1", "metric2"},
	}
	patches.ApplyFuncReturn(getMetricConfigMap, metricConfigMap)
	patches.ApplyFuncReturn(listResources, nil, errors.New("test err"))

	logs.InitLog()
	cfwInfoTest := CFWInfo{}
	labelInfos, filterMetrics := cfwInfoTest.GetResourceInfo()
	assert.Nil(t, labelInfos)
	assert.Nil(t, filterMetrics)
}

func TestCFWInfo_GetResourceInfo_success(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	metricConfigMap := map[string][]string{
		"fw_instance_id": {"metric1", "metric2"},
	}
	patches.ApplyFuncReturn(getMetricConfigMap, metricConfigMap)
	patches.ApplyFuncReturn(listResources, resourceEntityInit(), nil)
	logs.InitLog()
	cfwInfoTest := CFWInfo{}
	// 两个指标，两个资源
	labelInfos, filterMetrics := cfwInfoTest.GetResourceInfo()
	assert.NotNil(t, labelInfos)
	assert.Equal(t, 2, len(labelInfos))
	assert.NotNil(t, filterMetrics)
	assert.Equal(t, 4, len(filterMetrics))
}

func getPatches() *gomonkey.Patches {
	confLoader := &logs.ConfLoader{}
	patches := gomonkey.ApplyMethodFunc(*confLoader, "LoadFile", func(fPath string, cfg interface{}) error {
		cfgTmp, _ := cfg.(*map[string][]logs.Config)
		cfgPointer := make(map[string][]logs.Config)
		cfgPointer["business"] = []logs.Config{
			{
				Level: zapcore.InfoLevel,
			},
		}
		*cfgTmp = cfgPointer
		return nil
	})
	return patches
}

func resourceEntityInit() []model.ResourceEntity {
	id1 := "xxxx2"
	name1 := "test"
	epId1 := "1"
	epName1 := "测试企业1"
	checksum1 := "xxxb"
	create1 := "2023-12-23T07:58:14.000Z"
	update1 := "2023-12-23T07:58:14.000Z"
	provisioningState1 := "Succeeded"

	id2 := "xxxx3"
	name2 := "test2"
	epId2 := "2"
	epName2 := "测试企业2"
	checksum2 := "xxxc"
	create2 := "2023-12-23T07:58:14.000Z"
	update2 := "2023-12-25T07:58:14.000Z"
	provisioningState2 := "Succeeded"

	provider := "cfw"
	typestr := "cfw_instance"
	regionId := "cn-north-7"
	projectId := "xxxx0"
	projectName := "cn-north-7"
	tags := map[string]string{}
	properties := map[string]interface{}{
		"domain_id":             "xxxx001",
		"engine_type":           0,
		"enterprise_project_id": "1",
		"service_type":          0,
		"project_id":            "xxxx002",
		"fw_instance_name":      "test",
		"name":                  "1703318294687",
		"policy_count":          0,
		"fw_instance_id":        "xxxx003",
		"status":                "status",
	}

	response := []model.ResourceEntity{
		{
			Id:                &id1,
			Name:              &name1,
			Provider:          &provider,
			Type:              &typestr,
			RegionId:          &regionId,
			ProjectId:         &projectId,
			ProjectName:       &projectName,
			EpId:              &epId1,
			EpName:            &epName1,
			Checksum:          &checksum1,
			Created:           &create1,
			Updated:           &update1,
			ProvisioningState: &provisioningState1,
			Tags:              tags,
			Properties:        properties,
		},
		{
			Id:                &id2,
			Name:              &name2,
			Provider:          &provider,
			Type:              &typestr,
			RegionId:          &regionId,
			ProjectId:         &projectId,
			ProjectName:       &projectName,
			EpId:              &epId2,
			EpName:            &epName2,
			Checksum:          &checksum2,
			Created:           &create2,
			Updated:           &update2,
			ProvisioningState: &provisioningState2,
			Tags:              tags,
			Properties:        properties,
		},
	}
	return response
}
