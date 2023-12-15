package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	"github.com/stretchr/testify/assert"
)

func TestGaussdbGetResourceInfo(t *testing.T) {
	sysConfig := map[string][]string{"gaussdb_mysql_instance_id,gaussdb_mysql_node_id": {"gaussdb_mysql114_innodb_bufpool_read_ahead"}}
	patches := gomonkey.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	nodes := mockRmsResource()
	nodes[0].Properties = map[string]interface{}{
		"dimensions": []model.MetricsDimension{
			{Name: "gaussdb_mysql_instance_id", Value: "0001-0001-0000001"},
			{Name: "gaussdb_mysql_node_id", Value: "node-0001-0001-0000001"},
		},
	}
	patches.ApplyFuncReturn(listResources, mockRmsResource(), nil)
	defer patches.Reset()

	var gaussdbgetter GAUSSDBInfo
	labels, metrics := gaussdbgetter.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 1, len(metrics))
}
