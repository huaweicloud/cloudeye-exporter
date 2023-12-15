package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestEvsGetResourceInfo(t *testing.T) {
	sysConfig := map[string][]string{"disk_name": {"disk_device_read_bytes_rate"}}
	patches := gomonkey.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	patches.ApplyFuncReturn(getResourceFromRMS, true)
	volumes := mockRmsResource()
	volumes[0].Properties = map[string]interface{}{
		"attachments": []Attachment{
			{Device: "vad", ServerId: "0001-0001-00000001"},
		},
	}
	patches.ApplyFuncReturn(listResources, volumes, nil)
	defer patches.Reset()

	var evsgetter EVSInfo
	labels, metrics := evsgetter.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 1, len(metrics))
}
