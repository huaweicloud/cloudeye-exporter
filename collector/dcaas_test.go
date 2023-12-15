package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestDcaassGetResourceInfo(t *testing.T) {
	sysConfig := map[string][]string{
		"virtual_gateway_id":   {"network_incoming_bits_rate"},
		"virtual_interface_id": {"network_incoming_bits_rate"},
		"direct_connect_id":    {"network_incoming_bits_rate"}}
	patches := gomonkey.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	patches.ApplyFuncReturn(listResources, mockRmsResource(), nil)
	defer patches.Reset()

	var dcaasgetter DCAASInfo
	labels, metrics := dcaasgetter.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 3, len(metrics))
}
