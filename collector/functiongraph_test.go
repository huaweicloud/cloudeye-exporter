package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestFunctionGraphGetResourceInfo(t *testing.T) {
	sysConfig := map[string][]string{"package-functionname": {"count"}}
	patches := gomonkey.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	patches.ApplyFuncReturn(listResources, mockRmsResource(), nil)
	defer patches.Reset()

	var fsgetter FunctionGraphInfo
	labels, metrics := fsgetter.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 1, len(metrics))
}
