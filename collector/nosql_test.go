package collector

import (
	"errors"
	"testing"

	nosqlmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/gaussdbfornosql/v3/model"
	"github.com/stretchr/testify/assert"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

func TestNoSQLInfo_GetResourceInfo(t *testing.T) {
	InitCloudConf("../clouds.yml")
	conf.AuthMode = "aksk"
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_secret_key"
	instance := []nosqlmodel.ListInstancesResult{
		{
			Id:   "1111",
			Name: "test_xxx",
			Datastore: &nosqlmodel.ListInstancesDatastoreResult{
				Type: "redis",
			},
			Groups: []nosqlmodel.ListInstancesGroupResult{
				{
					Nodes: []nosqlmodel.ListInstancesNodeResult{
						{
							Id:   "test_node111",
							Name: "test_node1",
						},
						{
							Id:   "test_node222",
							Name: "test_node2",
						},
					},
				},
			},
		},
	}
	sysConfig := map[string][]string{
		"redis_cluster_id":               {"redis667_cluster_qps"},
		"redis_cluster_id,redis_node_id": {"gemini001_cpu_usage"},
		"mongodb_cluster_id":             {"mongodb001_command_ps"},
	}

	var dbInstance []DynamoTableInfo

	response := &DynamoTablesResp{
		TotalCount:     1,
		Tables:         dbInstance,
		HttpStatusCode: 200,
	}
	patches := getPatches()
	patches = patches.ApplyFuncReturn(getAllNoSQLInstances, instance, nil)
	patches = patches.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	patches = patches.ApplyFuncReturn(getDynamoTables, response, nil)
	defer patches.Reset()
	logs.InitLog("")
	noSQLInfo1 := NoSQLInfo{}
	_, filteredMetricInfos := noSQLInfo1.GetResourceInfo()
	assert.NotNil(t, filteredMetricInfos)
}

func TestNoSQLInfo_GetDynamoDbInstances_ResponseInvalid(t *testing.T) {
	conf.AuthMode = "aksk"
	InitCloudConf("../clouds.yml")
	var dbInstance []DynamoTableInfo

	response := DynamoTablesResp{
		TotalCount:     1,
		Tables:         dbInstance,
		HttpStatusCode: 200,
	}
	patches := getPatches()
	patches = patches.ApplyFuncReturn(getDynamoTables, response, nil)
	defer patches.Reset()
	logs.InitLog("")
	_, err := getDynamoDbInstances()
	assert.NotNil(t, err)
}

func TestNoSQLInfo_GetDynamoDbInstances_RequestError(t *testing.T) {
	conf.AuthMode = "aksk"
	InitCloudConf("../clouds.yml")
	var dbInstance []DynamoTableInfo

	response := &DynamoTablesResp{
		TotalCount:     1,
		Tables:         dbInstance,
		HttpStatusCode: 200,
	}
	patches := getPatches()
	patches = patches.ApplyFuncReturn(getDynamoTables, response, errors.New("resp type is not DynamoTablesResp"))
	defer patches.Reset()
	logs.InitLog("")
	_, err := getDynamoDbInstances()
	assert.NotNil(t, err)
}
