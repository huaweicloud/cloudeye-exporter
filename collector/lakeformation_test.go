package collector

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	lakeformationmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/lakeformation/v1/model"
)

func TestLakeFormationGetResourceInfoViaRMS(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	logs.InitLog("")
	lakeformationInfo = serversInfo{}

	// Set metricConf with Resource: "rms" so getResourceFromRMS returns true naturally
	metricConf = map[string]MetricConf{
		"SYS.LakeFormation": {
			Resource: "rms",
			DimMetricName: map[string][]string{
				"instance_id": {"lk_cpu_util", "lk_mem_util"},
			},
		},
	}

	// Only mock external dependencies, not getResourceFromRMS/getMetricConfigMap
	patches.ApplyFuncReturn(getAllInstanceFromRMS, []ResourceBaseInfo{
		{ID: "lk-inst-001", Name: "lk-instance-1", EpId: "ep-001", Tags: map[string]string{"env": "prod"}},
		{ID: "lk-inst-002", Name: "lk-instance-2", EpId: "ep-002", Tags: map[string]string{"env": "test"}},
	}, nil)
	patches.ApplyFuncReturn(GetResourceInfoExpirationTime, 5*time.Minute)

	var getter LakeFormationInfo
	labels, metrics := getter.GetResourceInfo()

	assert.Equal(t, 2, len(labels))
	assert.Equal(t, 4, len(metrics)) // 2 instances × 2 metrics
	assert.Greater(t, lakeformationInfo.TTL, time.Now().Unix()-1)
}

func TestLakeFormationGetResourceInfoViaAPI(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	logs.InitLog("")
	lakeformationInfo = serversInfo{}

	// No Resource field → getResourceFromRMS returns false → API path
	metricConf = map[string]MetricConf{
		"SYS.LakeFormation": {
			DimMetricName: map[string][]string{
				"instance_id": {"lk_cpu_util"},
			},
		},
	}

	patches.ApplyFuncReturn(getEpIdRequestPart, []string{"all_granted_eps"})

	instanceId := "lk-api-001"
	instanceName := "api-instance-1"
	epId := "ep-api"
	patches.ApplyFunc(getLKInstanceByEpId, func(epIdParam string) ([]ResourceBaseInfo, error) {
		return []ResourceBaseInfo{
			{ID: instanceId, Name: instanceName, EpId: epId},
		}, nil
	})

	patches.ApplyFuncReturn(GetResourceInfoExpirationTime, 5*time.Minute)

	var getter LakeFormationInfo
	labels, metrics := getter.GetResourceInfo()

	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 1, len(metrics))
}

func TestLakeFormationGetResourceInfoCache(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	logs.InitLog("")
	cachedLabels := map[string]labelInfo{
		"cached-lk-id": {Name: []string{"name", "epId"}, Value: []string{"cached-lk", "ep-0"}},
	}
	cachedMetrics := []model.MetricInfoList{
		{Namespace: "SYS.LakeFormation", MetricName: "lk_cpu_util"},
	}
	lakeformationInfo = serversInfo{
		TTL:           time.Now().Add(10 * time.Minute).Unix(),
		LabelInfo:     cachedLabels,
		FilterMetrics: cachedMetrics,
	}

	patches.ApplyFunc(getAllInstanceFromRMS, func() ([]ResourceBaseInfo, error) {
		t.Fatal("getAllInstanceFromRMS should not be called when cache is valid")
		return nil, nil
	})
	patches.ApplyFunc(getAllLKInstance, func() ([]ResourceBaseInfo, error) {
		t.Fatal("getAllLKInstance should not be called when cache is valid")
		return nil, nil
	})

	var getter LakeFormationInfo
	labels, metrics := getter.GetResourceInfo()

	assert.Equal(t, 1, len(labels))
	assert.Equal(t, []string{"cached-lk", "ep-0"}, labels["cached-lk-id"].Value)
	assert.Equal(t, 1, len(metrics))

	lakeformationInfo = serversInfo{}
}

func TestLakeFormationGetResourceInfoInstanceError(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	logs.InitLog("")
	lakeformationInfo = serversInfo{}

	// No Resource → API path
	metricConf = map[string]MetricConf{
		"SYS.LakeFormation": {
			DimMetricName: map[string][]string{
				"instance_id": {"lk_cpu_util"},
			},
		},
	}

	patches.ApplyFuncReturn(getAllLKInstance, nil, fmt.Errorf("api connection error"))
	patches.ApplyFuncReturn(GetResourceInfoExpirationTime, 5*time.Minute)

	var getter LakeFormationInfo
	labels, metrics := getter.GetResourceInfo()

	assert.Nil(t, labels)
	assert.Nil(t, metrics)
}

func TestLakeFormationGetResourceInfoEmptyInstances(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	logs.InitLog("")
	lakeformationInfo = serversInfo{}

	metricConf = map[string]MetricConf{
		"SYS.LakeFormation": {
			DimMetricName: map[string][]string{
				"instance_id": {"lk_cpu_util"},
			},
		},
	}

	patches.ApplyFuncReturn(getAllLKInstance, []ResourceBaseInfo{}, nil)
	patches.ApplyFuncReturn(GetResourceInfoExpirationTime, 5*time.Minute)

	var getter LakeFormationInfo
	labels, metrics := getter.GetResourceInfo()

	assert.Equal(t, 0, len(labels))
	assert.Equal(t, 0, len(metrics))
}

func TestLakeFormationGetResourceInfoCacheExpired(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	logs.InitLog("")
	lakeformationInfo = serversInfo{
		TTL:       time.Now().Add(-1 * time.Minute).Unix(),
		LabelInfo: map[string]labelInfo{"old-id": {Name: []string{"name"}, Value: []string{"old-lk"}}},
	}

	// Use RMS path to avoid getLakeFormationClient() panic
	metricConf = map[string]MetricConf{
		"SYS.LakeFormation": {
			Resource: "rms",
			DimMetricName: map[string][]string{
				"instance_id": {"lk_cpu_util"},
			},
		},
	}

	patches.ApplyFuncReturn(getAllInstanceFromRMS, []ResourceBaseInfo{
		{ID: "new-lk-id", Name: "new-lk", EpId: "ep-new"},
	}, nil)
	patches.ApplyFuncReturn(GetResourceInfoExpirationTime, 5*time.Minute)

	var getter LakeFormationInfo
	labels, metrics := getter.GetResourceInfo()

	assert.Equal(t, 1, len(labels))
	assert.Equal(t, []string{"new-lk", "ep-new"}, labels["new-lk-id"].Value)
	assert.Equal(t, 1, len(metrics))

	lakeformationInfo = serversInfo{}
}

func TestGetAllLKInstance(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	patches.ApplyFuncReturn(getEpIdRequestPart, []string{"ep-1", "ep-2"})
	patches.ApplyFunc(getLKInstanceByEpId, func(epId string) ([]ResourceBaseInfo, error) {
		if epId == "ep-1" {
			return []ResourceBaseInfo{
				{ID: "inst-1", Name: "name-1", EpId: "ep-1"},
			}, nil
		}
		return []ResourceBaseInfo{
			{ID: "inst-2", Name: "name-2", EpId: "ep-2"},
		}, nil
	})

	results, err := getAllLKInstance()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(results))
	assert.Equal(t, "inst-1", results[0].ID)
	assert.Equal(t, "inst-2", results[1].ID)
}

func TestGetAllLKInstanceError(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	patches.ApplyFuncReturn(getEpIdRequestPart, []string{"ep-1"})
	patches.ApplyFuncReturn(getLKInstanceByEpId, nil, fmt.Errorf("connection timeout"))

	results, err := getAllLKInstance()
	assert.NotNil(t, err)
	assert.Nil(t, results)
}

func TestGetLKInstanceByEpId(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	conf.Region = "cn-test-01"
	conf.AuthMode = "aksk"
	endpointConfig = map[string]string{
		"lakeformation": "https://lakeformation.cn-test-01.myhuaweicloud.com",
	}

	instId := "lk-001"
	instName := "test-instance"
	instEpId := "ep-001"

	respPage1 := lakeformationmodel.ListLakeFormationInstancesResponse{
		HttpStatusCode: 200,
		Instances: &[]lakeformationmodel.LakeFormationInstance{
			{InstanceId: &instId, Name: &instName, EnterpriseProjectId: &instEpId},
		},
	}
	respPage2 := lakeformationmodel.ListLakeFormationInstancesResponse{
		HttpStatusCode: 200,
		Instances:      &[]lakeformationmodel.LakeFormationInstance{},
	}

	lfClient := getLakeFormationClient()
	patches.ApplyMethodFunc(lfClient, "ListLakeFormationInstances",
		func(req *lakeformationmodel.ListLakeFormationInstancesRequest) (*lakeformationmodel.ListLakeFormationInstancesResponse, error) {
			if req.Offset == 0 {
				return &respPage1, nil
			}
			return &respPage2, nil
		})

	results, err := getLKInstanceByEpId("ep-001")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "lk-001", results[0].ID)
	assert.Equal(t, "test-instance", results[0].Name)
	assert.Equal(t, "ep-001", results[0].EpId)
}
