package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestNatGetResourceInfo(t *testing.T) {
	sysConfig := map[string][]string{"nat_gateway_id": {"snat_connection"}}
	patches := gomonkey.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	patches.ApplyFuncReturn(listResources, mockRmsResource(), nil)
	defer patches.Reset()

	var netgetter NATInfo
	labels, metrics := netgetter.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 1, len(metrics))
}
