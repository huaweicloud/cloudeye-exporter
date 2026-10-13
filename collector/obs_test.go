package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

func TestOBSInfo_GetResourceInfo(t *testing.T) {
	obsServerInfo = newServersInfo{}
	metricConf = map[string]MetricConf{
		"SYS.OBS": {
			DimMetricName: map[string][]string{
				"bucket_name": {"download_traffic"},
			},
		},
	}
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	patches.ApplyFuncReturn(listAllMetrics, []model.MetricInfoList{
		{
			Dimensions: []model.MetricsDimension{
				{
					Name:  "bucket_name",
					Value: "test_name",
				},
			},
			Namespace:  "SYS.OBS",
			MetricName: "download_traffic",
		},
	}, nil)
	patches.ApplyFuncReturn(getUserInfoFromIAM, map[string]string{})
	patches.ApplyFuncReturn(getAllServerFromRMS, []EcsInstancesInfo{}, nil)
	var obsInfo OBSInfo
	CloudConf.Global.MetricInfoExpirationDays = 2
	CloudConf.Global.MetricInfoCleanThreshold = 50000
	labels, metrics := obsInfo.GetResourceInfo()
	assert.NotEmpty(t, metrics)
	assert.Empty(t, labels)
}

func TestOBSInfo_GetResourceInfo_BucketNameWithDot(t *testing.T) {
	obsServerInfo = newServersInfo{}
	metricConf = map[string]MetricConf{
		"SYS.OBS": {
			DimMetricName: map[string][]string{
				"bucket_name": {"download_traffic"},
			},
		},
	}
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	// CES侧bucket_name维度值将点转义为下划线: test.help -> test_help
	patches.ApplyFuncReturn(listAllMetrics, []model.MetricInfoList{
		{
			Dimensions: []model.MetricsDimension{
				{
					Name:  "bucket_name",
					Value: "test_help",
				},
			},
			Namespace:  "SYS.OBS",
			MetricName: "download_traffic",
		},
	}, nil)
	patches.ApplyFuncReturn(getUserInfoFromIAM, map[string]string{})
	// RMS返回原始桶名: test.help
	patches.ApplyFuncReturn(getAllServerFromRMS, []EcsInstancesInfo{
		{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   "id-1",
				Name: "test.help",
				EpId: "ep-1",
			},
		},
	}, nil)
	var obsInfo OBSInfo
	CloudConf.Global.MetricInfoExpirationDays = 2
	CloudConf.Global.MetricInfoCleanThreshold = 50000
	labels, metrics := obsInfo.GetResourceInfo()
	assert.NotEmpty(t, metrics)
	// 桶名匹配成功后,resourceInfos应包含该桶的epId扩展标签
	assert.NotEmpty(t, labels)
	info, ok := labels["test_help"]
	assert.True(t, ok)
	assert.Equal(t, "epId", info.Name[0])
	assert.Equal(t, "ep-1", info.Value[0])
}
