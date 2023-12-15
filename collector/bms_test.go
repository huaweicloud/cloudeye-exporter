package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestBmsGetResourceInfo(t *testing.T) {
	sysConfig := map[string][]string{"instance_id": {"cpu_utils"}}
	instances := []EcsInstancesInfo{
		{ResourceBaseInfo: ResourceBaseInfo{ID: "0001-0001-000000001", Name: "host01", EpId: "0"}},
	}
	patches := gomonkey.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	patches.ApplyFuncReturn(getAllServerFromRMS, instances, nil)
	patches.ApplyFunc(loadAgentDimensions, func(_ string) { return })
	patches.ApplyFuncReturn(getIPFromEcsInfo, "")
	defer patches.Reset()

	var bmsGetter BMSInfo
	labels, metrics := bmsGetter.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 1, len(metrics))

	var servicesGetter SERVICEBMSInfo
	serviceLabel, _ := servicesGetter.GetResourceInfo()
	assert.Equal(t, 1, len(serviceLabel))
}
