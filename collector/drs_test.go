package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestDrsGetResourceInfo(t *testing.T) {
	sysConfig := map[string][]string{"instance_id": {"cpu_util"}}
	patches := gomonkey.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	patches.ApplyFuncReturn(listResources, mockRmsResource(), nil)
	defer patches.Reset()

	var drsgetter DRSInfo
	labels, metrics := drsgetter.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 5, len(metrics))
}
