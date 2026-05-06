package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	ddmmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ddm/v1/model"
)

func TestDdmsGetResourceInfo(t *testing.T) {
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	conf.AuthMode = "aksk"
	conf.Region = "cn-test-01"

	patches := gomonkey.NewPatches()
	metricConf = map[string]MetricConf{
		"SYS.DDMS": {
			Resource: "rms",
			DimMetricName: map[string][]string{
				"instance_id,node_id": {"cpu_util"},
			},
		},
	}
	patches.ApplyFuncReturn(getAllDdmsInstances, mockDdmsInstances(), nil)
	defer patches.Reset()

	var ddmsInfo DDMSInfo
	labels, metrics := ddmsInfo.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 1, len(metrics))
}

func TestGetAllDdmsInstances(t *testing.T) {
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	conf.AuthMode = "aksk"
	conf.Region = "cn-test-01"

	var (
		instanceId   = "ddms-instance-001"
		instanceName = "test-ddms-instance"
		accessIp     = "192.168.1.100"
		accessPort   = "3306"
		epId         = "0"
	)

	instanceResp := ddmmodel.ListInstancesResponse{
		HttpStatusCode: 200,
		Instances: &[]ddmmodel.ShowInstanceBeanResponse{
			{
				Id:                  instanceId,
				Name:                instanceName,
				AccessIp:            accessIp,
				AccessPort:          accessPort,
				EnterpriseProjectId: epId,
			},
		},
	}

	emptyResp := ddmmodel.ListInstancesResponse{
		HttpStatusCode: 200,
		Instances:      &[]ddmmodel.ShowInstanceBeanResponse{},
	}

	ddmsClient := getDDMSClient()
	patches := gomonkey.ApplyMethodFunc(ddmsClient, "ListInstances", func(req *ddmmodel.ListInstancesRequest) (*ddmmodel.ListInstancesResponse, error) {
		if *req.Offset == 0 {
			return &instanceResp, nil
		}
		return &emptyResp, nil
	})
	patches.ApplyFuncReturn(getDdmsInstanceNodes, mockDdmsNodes(), nil)
	defer patches.Reset()

	instances, err := getAllDdmsInstances()
	assert.Equal(t, true, err == nil)
	assert.Equal(t, 1, len(instances))
	assert.Equal(t, instanceId, instances[0].Instance.Id)
	assert.Equal(t, instanceName, instances[0].Instance.Name)
}

func TestGetDdmsInstanceNodes(t *testing.T) {
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	conf.AuthMode = "aksk"
	conf.Region = "cn-test-01"

	var (
		instanceId = "ddms-instance-001"
		nodeId1    = "node-001"
		nodeName1  = "test-node-1"
		privateIp1 = "192.168.1.101"
		nodeId2    = "node-002"
		nodeName2  = "test-node-2"
		privateIp2 = "192.168.1.102"
	)

	listResp := ddmmodel.ListNodesResponse{
		HttpStatusCode: 200,
		Nodes: &[]ddmmodel.NodeList{
			{NodeId: &nodeId1},
			{NodeId: &nodeId2},
		},
	}

	emptyListResp := ddmmodel.ListNodesResponse{
		HttpStatusCode: 200,
		Nodes:          &[]ddmmodel.NodeList{},
	}

	showResp1 := ddmmodel.ShowNodeResponse{
		NodeId:    &nodeId1,
		Name:      &nodeName1,
		PrivateIp: &privateIp1,
	}

	showResp2 := ddmmodel.ShowNodeResponse{
		NodeId:    &nodeId2,
		Name:      &nodeName2,
		PrivateIp: &privateIp2,
	}

	ddmsClient := getDDMSClient()
	patches := gomonkey.ApplyMethodFunc(ddmsClient, "ListNodes", func(req *ddmmodel.ListNodesRequest) (*ddmmodel.ListNodesResponse, error) {
		if *req.Offset == 0 {
			return &listResp, nil
		}
		return &emptyListResp, nil
	})
	patches.ApplyMethodReturn(ddmsClient, "ShowNode", &showResp1, nil)
	patches.ApplyMethodReturn(ddmsClient, "ShowNode", &showResp2, nil)
	defer patches.Reset()

	nodes, err := getDdmsInstanceNodes(instanceId)
	assert.Equal(t, true, err == nil)
	assert.Equal(t, 2, len(nodes))
}

func TestGetDdmsInstanceNodesError(t *testing.T) {
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	conf.AuthMode = "aksk"
	conf.Region = "cn-test-01"

	ddmsClient := getDDMSClient()
	patches := gomonkey.ApplyMethodFunc(ddmsClient, "ListNodes", func(req *ddmmodel.ListNodesRequest) (*ddmmodel.ListNodesResponse, error) {
		return nil, assert.AnError
	})
	defer patches.Reset()

	nodes, err := getDdmsInstanceNodes("ddms-instance-001")
	assert.Equal(t, true, err != nil)
	assert.Equal(t, 0, len(nodes))
}

func TestGetAllDdmsInstancesError(t *testing.T) {
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	conf.AuthMode = "aksk"
	conf.Region = "cn-test-01"
	logs.InitLog("../logs.yml")

	ddmsClient := getDDMSClient()
	patches := gomonkey.ApplyMethodFunc(ddmsClient, "ListInstances", func(req *ddmmodel.ListInstancesRequest) (*ddmmodel.ListInstancesResponse, error) {
		return nil, assert.AnError
	})
	defer patches.Reset()

	instances, err := getAllDdmsInstances()
	assert.Equal(t, true, err != nil)
	assert.Equal(t, 0, len(instances))
}

func mockDdmsInstances() []DdmsInstanceInfo {
	var (
		instanceId   = "ddms-instance-001"
		instanceName = "test-ddms-instance"
		accessIp     = "192.168.1.100"
		accessPort   = "3306"
		epId         = "0"
		nodeId       = "node-001"
		nodeName     = "test-node"
		privateIp    = "192.168.1.101"
	)

	instance := ddmmodel.ShowInstanceBeanResponse{
		Id:                  instanceId,
		Name:                instanceName,
		AccessIp:            accessIp,
		AccessPort:          accessPort,
		EnterpriseProjectId: epId,
	}

	node := ddmmodel.ShowNodeResponse{
		NodeId:    &nodeId,
		Name:      &nodeName,
		PrivateIp: &privateIp,
	}

	return []DdmsInstanceInfo{
		{Instance: instance, Nodes: []ddmmodel.ShowNodeResponse{node}},
	}
}

func mockDdmsNodes() []ddmmodel.ShowNodeResponse {
	var (
		nodeId1    = "node-001"
		nodeName1  = "test-node-1"
		privateIp1 = "192.168.1.101"
	)

	return []ddmmodel.ShowNodeResponse{
		{
			NodeId:    &nodeId1,
			Name:      &nodeName1,
			PrivateIp: &privateIp1,
		},
	}
}
