package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/apig/v2/model"
	"github.com/stretchr/testify/assert"
)

func TestShowDetailsOfInstanceV2(t *testing.T) {
	apigClient := getAPICSClient()
	id := "0001-0001-0000001"
	name := "instance01"
	instanceInfo := model.ShowDetailsOfInstanceV2Response{
		Id:           &id,
		InstanceName: &name,
	}
	patches := gomonkey.ApplyMethodReturn(apigClient, "ShowDetailsOfInstanceV2", &instanceInfo, nil)
	defer patches.Reset()
	resp, err := showDetailsOfInstanceV2(id)
	assert.Equal(t, true, err == nil)
	assert.Equal(t, name, *resp.InstanceName)
}

func TestGetResourceInfo(t *testing.T) {
	var (
		id       = "0001-0001-000001"
		name     = "instance01"
		eip      = "*.*.*.*"
		epId     = "0"
		apiId    = "api0000001"
		emptyStr = ""
	)

	sysConfig := map[string][]string{
		"instance_id":         {"requests"},
		"instance_id,api_id":  {"req_count"},
		"instance_id,node_ip": {"node_qps"},
	}

	instances := []model.RespInstanceBase{{Id: &id, InstanceName: &name, EipAddress: &eip, EnterpriseProjectId: &epId}}
	apis := []model.ApiInfoPerPage{{Id: &apiId, Name: "api1", GroupName: &emptyStr, GroupId: "group1"}}
	instance := model.ShowDetailsOfInstanceV2Response{
		NodeIps: &model.NodeIps{
			Livedata: &[]string{"*.*.*.2"},
			Shubao:   &[]string{"*.*.*.3"},
		},
	}

	patches := gomonkey.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	patches.ApplyFuncReturn(getAllAPICInstances, instances, nil)
	patches.ApplyFuncReturn(getApisOfInstances, apis, nil)
	patches.ApplyFuncReturn(showDetailsOfInstanceV2, &instance, nil)
	defer patches.Reset()
	var getter = APICInfo{}
	label, metrics := getter.GetResourceInfo()
	assert.Equal(t, 4, len(label))
	assert.Equal(t, 4, len(metrics))
}

func TestGetAllAPICInstances(t *testing.T) {
	var (
		id   = "0001-0001-000001"
		name = "instance01"
	)
	respPage1 := model.ListInstancesV2Response{
		HttpStatusCode: 200,
		Instances: &[]model.RespInstanceBase{
			{Id: &id, InstanceName: &name},
		},
	}
	respPage2 := model.ListInstancesV2Response{
		HttpStatusCode: 200,
		Instances:      &[]model.RespInstanceBase{},
	}
	apicClient := getAPICSClient()
	patches := gomonkey.ApplyMethodFunc(apicClient, "ListInstancesV2", func(req *model.ListInstancesV2Request) (*model.ListInstancesV2Response, error) {
		if *req.Offset == 0 {
			return &respPage1, nil
		}
		return &respPage2, nil
	})
	defer patches.Reset()
	instances, err := getAllAPICInstances()
	assert.Equal(t, true, err == nil)
	assert.Equal(t, 1, len(instances))
}

func TestGetApisOfInstances(t *testing.T) {
	var (
		id   = "0001-0001-000001"
		name = "api01"
	)
	respPage1 := model.ListApisV2Response{
		HttpStatusCode: 200,
		Apis: &[]model.ApiInfoPerPage{
			{Id: &id, Name: name},
		},
	}
	respPage2 := model.ListApisV2Response{
		HttpStatusCode: 200,
		Apis:           &[]model.ApiInfoPerPage{},
	}
	apicClient := getAPICSClient()
	patches := gomonkey.ApplyMethodFunc(apicClient, "ListApisV2", func(req *model.ListApisV2Request) (*model.ListApisV2Response, error) {
		if *req.Offset == 0 {
			return &respPage1, nil
		}
		return &respPage2, nil
	})
	defer patches.Reset()
	apis, err := getApisOfInstances("00001")
	assert.Equal(t, true, err == nil)
	assert.Equal(t, 1, len(apis))
}
