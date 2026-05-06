package collector

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	ddsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dds/v3/model"
)

func TestDdsGetResourceInfo(t *testing.T) {
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	conf.Region = "cn-test-01"

	patches := getPatches()
	defer patches.Reset()

	metricConf = map[string]MetricConf{
		"SYS.DDS": {
			Resource: "rms",
			DimMetricName: map[string][]string{
				"mongodb_instance_id": {"cpu_util"},
			},
		},
	}
	patches.ApplyFuncReturn(getAllDdsInstances, mockDdsInstances(), nil)

	logs.InitLog("")
	var ddsInfo DDSInfo
	labels, metrics := ddsInfo.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 1, len(metrics))
}

func TestDdsGetResourceInfo_withNodes(t *testing.T) {
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	conf.Region = "cn-test-01"

	patches := getPatches()
	defer patches.Reset()

	metricConf = map[string]MetricConf{
		"SYS.DDS": {
			Resource: "rms",
			DimMetricName: map[string][]string{
				"mongodb_instance_id,mongodb_node_id": {"cpu_util"},
				"mongodb_instance_id":                 {"cpu_util"},
			},
		},
	}
	patches.ApplyFuncReturn(getAllDdsInstances, mockDdsInstancesWithNodes(), nil)

	logs.InitLog("./logs.yml")
	var ddsInfoClient DDSInfo
	ddsInfo.LabelInfo = nil
	ddsInfo.TTL = 0
	labels, metrics := ddsInfoClient.GetResourceInfo()
	assert.Equal(t, 3, len(labels))
	assert.Equal(t, 3, len(metrics))
}

func TestDdsGetResourceInfo_getAllDdsInstancesErr(t *testing.T) {
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	conf.Region = "cn-test-01"

	patches := getPatches()
	defer patches.Reset()

	metricConf = map[string]MetricConf{
		"SYS.DDS": {
			Resource: "rms",
			DimMetricName: map[string][]string{
				"mongodb_instance_id": {"cpu_util"},
			},
		},
	}
	patches.ApplyFuncReturn(getAllDdsInstances, nil, errors.New("list dds instances error"))

	logs.InitLog("../logs.yml")
	var ddsInfoClient DDSInfo
	ddsInfo.LabelInfo = nil
	ddsInfo.FilterMetrics = nil
	ddsInfo.TTL = 0
	labels, metrics := ddsInfoClient.GetResourceInfo()
	assert.Nil(t, labels)
	assert.Nil(t, metrics)
}

func TestDdsGetResourceInfo_dimConfigNotExists(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	metricConf = map[string]MetricConf{
		"SYS.DDS": {
			Resource:      "rms",
			DimMetricName: map[string][]string{},
		},
	}

	logs.InitLog("../logs.yml")
	var ddsInfoClient DDSInfo
	ddsInfo.LabelInfo = nil
	ddsInfo.FilterMetrics = nil
	ddsInfo.TTL = 0
	labels, metrics := ddsInfoClient.GetResourceInfo()
	assert.Nil(t, labels)
	assert.Nil(t, metrics)
}

func TestGetAllDdsInstances(t *testing.T) {
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	conf.Region = "cn-test-01"
	conf.AuthMode = "aksk"
	logs.InitLog("../logs.yml")
	patches := getPatches()
	defer patches.Reset()

	outputs := []gomonkey.OutputCell{
		{
			Values: gomonkey.Params{mockListInstancesResponse(), nil},
		},
		{
			Values: gomonkey.Params{&ddsmodel.ListInstancesResponse{Instances: &[]ddsmodel.QueryInstanceResponse{}}, nil},
		},
	}
	patches.ApplyMethodSeq(getDDSClient(), "ListInstances", outputs)

	instances, err := getAllDdsInstances()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(instances))
	assert.Equal(t, "dds-instance-001", instances[0].ID)
	assert.Equal(t, "dds-instance-001", instances[0].Name)
	assert.Equal(t, "Sharding", instances[0].Mode)
	assert.Equal(t, "wiredTiger", instances[0].Engine)
	assert.Equal(t, "DDS-Community", instances[0].DatastoreType)
	assert.Equal(t, "3.4", instances[0].DatastoreVersion)
}

func TestGetAllDdsInstances_Error(t *testing.T) {
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	conf.Region = "cn-test-01"

	patches := getPatches()
	defer patches.Reset()
	conf.AuthMode = "aksk"
	logs.InitLog("../logs.yml")
	patches.ApplyFuncReturn(getDDSClient().ListInstances, nil, errors.New("list instances error"))

	instances, err := getAllDdsInstances()
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(instances))
}

func TestFmtDdsInstance(t *testing.T) {
	instance := mockQueryInstanceResponse()
	ddsInfos := fmtDdsInstance(instance)

	assert.Equal(t, "dds-instance-001", ddsInfos.ID)
	assert.Equal(t, "dds-instance-001", ddsInfos.Name)
	assert.Equal(t, "Sharding", ddsInfos.Mode)
	assert.Equal(t, "wiredTiger", ddsInfos.Engine)
	assert.Equal(t, "DDS-Community", ddsInfos.DatastoreType)
	assert.Equal(t, "3.4", ddsInfos.DatastoreVersion)
	assert.Equal(t, "0", ddsInfos.EpId)
	assert.Equal(t, "tagValue1", ddsInfos.Tags["tagKey1"])
	assert.Equal(t, 2, len(ddsInfos.Nodes))
}

func mockDdsInstances() []DdsInstanceInfo {
	return []DdsInstanceInfo{
		{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   "dds-instance-001",
				Name: "dds-instance-001",
				Tags: map[string]string{"tagKey1": "tagValue1"},
				EpId: "0",
			},
			Mode:             "Sharding",
			Engine:           "wiredTiger",
			DatastoreType:    "DDS-Community",
			DatastoreVersion: "3.4",
			Nodes:            []ddsmodel.NodeItem{},
		},
	}
}

func mockDdsInstancesWithNodes() []DdsInstanceInfo {
	node1 := ddsmodel.NodeItem{
		Id:        "node-001",
		Name:      "dds-instance-001-node-001",
		Role:      "master",
		PrivateIp: "192.168.1.1",
		PublicIp:  "",
	}
	node2 := ddsmodel.NodeItem{
		Id:        "node-002",
		Name:      "dds-instance-001-node-002",
		Role:      "Secondary",
		PrivateIp: "192.168.1.2",
		PublicIp:  "",
	}
	return []DdsInstanceInfo{
		{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   "dds-instance-001",
				Name: "dds-instance-001",
				Tags: map[string]string{"tagKey1": "tagValue1"},
				EpId: "0",
			},
			Mode:             "Sharding",
			Engine:           "wiredTiger",
			DatastoreType:    "DDS-Community",
			DatastoreVersion: "3.4",
			Nodes:            []ddsmodel.NodeItem{node1, node2},
		},
	}
}

func mockQueryInstanceResponse() ddsmodel.QueryInstanceResponse {
	tag1 := ddsmodel.TagResponse{
		Key:   "tagKey1",
		Value: "tagValue1",
	}
	datastore := ddsmodel.DatastoreItem{
		Type:    "DDS-Community",
		Version: "3.4",
	}
	node1 := ddsmodel.NodeItem{
		Id:        "node-001",
		Name:      "dds-instance-001-node-001",
		Role:      "master",
		PrivateIp: "192.168.1.1",
		PublicIp:  "",
	}
	node2 := ddsmodel.NodeItem{
		Id:        "node-002",
		Name:      "dds-instance-001-node-002",
		Role:      "Secondary",
		PrivateIp: "192.168.1.2",
		PublicIp:  "",
	}
	group := ddsmodel.GroupResponseItem{
		Type:  "shard",
		Id:    "group-001",
		Name:  "shardGroup1",
		Nodes: []ddsmodel.NodeItem{node1, node2},
	}
	return ddsmodel.QueryInstanceResponse{
		Id:                  "dds-instance-001",
		Name:                "dds-instance-001",
		Mode:                "Sharding",
		Engine:              "wiredTiger",
		EnterpriseProjectId: "0",
		Datastore:           &datastore,
		Tags:                []ddsmodel.TagResponse{tag1},
		Groups:              []ddsmodel.GroupResponseItem{group},
	}
}

func mockListInstancesResponse() *ddsmodel.ListInstancesResponse {
	instances := []ddsmodel.QueryInstanceResponse{mockQueryInstanceResponse()}
	totalCount := int32(1)
	return &ddsmodel.ListInstancesResponse{
		Instances:  &instances,
		TotalCount: &totalCount,
	}
}
