package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func TestRdsGetResourceInfo(t *testing.T) {
	sysConfig := map[string][]string{"rds_cluster_id": {"rds001_cpu_util"}}
	patches := gomonkey.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	volumes := mockRmsResource()
	volumes[0].Properties = map[string]interface{}{
		"cpu":        "2",
		"mem":        "16",
		"engineName": "mysql",
	}
	patches.ApplyFuncReturn(listResources, volumes, nil)
	defer patches.Reset()

	var rdsInfo RDSInfo
	labels, metrics := rdsInfo.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 1, len(metrics))
}
