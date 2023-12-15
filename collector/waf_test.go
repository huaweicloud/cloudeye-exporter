package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestWafGetResourceInfo(t *testing.T) {
	sysConfig := map[string][]string{"waf_instance_id": {"attacks"}}
	patches := gomonkey.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	patches.ApplyFuncReturn(listResources, mockRmsResource(), nil)
	defer patches.Reset()

	var wafgetter WAFInfo
	labels, metrics := wafgetter.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 1, len(metrics))
}
