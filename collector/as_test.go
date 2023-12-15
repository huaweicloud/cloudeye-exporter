package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestAsGetResourceInfo(t *testing.T) {
	sysConfig := map[string][]string{"AutoScalingGroup": {"req_count"}}
	groups := []ResourceBaseInfo{
		{ID: "0001-0001-000000001", Name: "group01", EpId: "0"},
	}
	patches := gomonkey.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	patches.ApplyFuncReturn(getResourcesBaseInfoFromRMS, groups, nil)
	defer patches.Reset()

	var asGetter ASInfo
	labels, metrics := asGetter.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 1, len(metrics))
}
