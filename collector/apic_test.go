package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/apig/v2/model"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	"github.com/stretchr/testify/assert"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

func TestShowDetailsOfInstanceV2(t *testing.T) {
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	conf.AuthMode = "aksk"
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
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	var (
		id       = "0001-0001-000001"
		name     = "instance01"
		eip      = "*.*.*.*"
		epId     = "0"
		apiId    = "api0000001"
		emptyStr = ""
	)

	metricConf = make(map[string]MetricConf)
	metricConf["SYS.APIC"] = MetricConf{
		Resource: "service",
		DimMetricName: map[string][]string{
			"instance_id":                               {"requests"},
			"instance_id,api_id":                        {"req_count"},
			"instance_id,node_ip":                       {"node_qps"},
			"instance_id,app_id":                        {"req_count"},
			"instance_id,api_group_id":                  {"req_count"},
			"instance_id,app_id,api_id":                 {"req_count"},
			"instance_id,plugin_type,plugin_id,node_ip": {"req_count"},
		},
	}

	instances := []model.RespInstanceBase{{Id: &id, InstanceName: &name, EipAddress: &eip, EnterpriseProjectId: &epId}}
	apis := []model.ApiInfoPerPage{{Id: &apiId, Name: "api1", GroupName: &emptyStr, GroupId: "group1"}}
	instance := model.ShowDetailsOfInstanceV2Response{
		NodeIps: &model.NodeIps{
			Livedata: &[]string{"*.*.*.2"},
			Shubao:   &[]string{"*.*.*.3"},
		},
	}
	bindNum := int32(10)
	apps := []model.AppInfoWithBindNum{{Name: getStringPointer("test-app-name"), BindNum: &bindNum, Id: getStringPointer("app-1")}}
	groups := []model.ApiGroupInfo{{Id: "group-1", Name: "test-name", Status: model.GetApiGroupInfoStatusEnum().E_1, Version: getStringPointer("1.1"), SlDomain: "0.0.0.0", Remark: getStringPointer("aaaaa")}}
	plugins := []model.PluginInfo{{PluginId: "000000", PluginName: "test-plugin", PluginType: model.GetPluginInfoPluginTypeEnum().CORS}}

	patches := getPatches()
	patches.ApplyFuncReturn(getAllAPICInstances, instances, nil)
	patches.ApplyFuncReturn(getApisOfInstances, apis, nil)
	patches.ApplyFuncReturn(showDetailsOfInstanceV2, &instance, nil)
	patches.ApplyFuncReturn(getApiGroupsOfInstances, groups, nil)
	patches.ApplyFuncReturn(getAppsOfInstances, apps, nil)
	patches.ApplyFuncReturn(getPlugins, plugins, nil)
	patches.ApplyFuncReturn(listAllMetrics, []cesmodel.MetricInfoList{
		{
			Namespace: "SYS.APIC",
			Dimensions: []cesmodel.MetricsDimension{
				{Name: "instance_id", Value: "0001-0001-000001"},
			},
			MetricName: "requests",
		},
		{
			Namespace: "SYS.APIC",
			Dimensions: []cesmodel.MetricsDimension{
				{Name: "instance_id", Value: "0001-0001-000001"},
				{Name: "api_id", Value: "api1"},
			},
			MetricName: "req_count",
		},
		{
			Namespace: "SYS.APIC",
			Dimensions: []cesmodel.MetricsDimension{
				{Name: "instance_id", Value: "0001-0001-000001"},
				{Name: "node_ip", Value: "*.*.*.2"},
			},
			MetricName: "node_qps",
		},
		{
			Namespace: "SYS.APIC",
			Dimensions: []cesmodel.MetricsDimension{
				{Name: "instance_id", Value: "0001-0001-000001"},
				{Name: "node_ip", Value: "*.*.*.3"},
			},
			MetricName: "node_qps",
		},
		{
			Namespace: "SYS.APIC",
			Dimensions: []cesmodel.MetricsDimension{
				{Name: "instance_id", Value: "0001-0001-000001"},
				{Name: "app_id", Value: "app-1"},
			},
			MetricName: "req_count",
		},
		{
			Namespace: "SYS.APIC",
			Dimensions: []cesmodel.MetricsDimension{
				{Name: "instance_id", Value: "0001-0001-000001"},
				{Name: "api_group_id", Value: "group-1"},
			},
			MetricName: "req_count",
		},
		{
			Namespace: "SYS.APIC",
			Dimensions: []cesmodel.MetricsDimension{
				{Name: "instance_id", Value: "0001-0001-000001"},
				{Name: "plugin_type", Value: "cors"},
				{Name: "plugin_id", Value: "000000"},
				{Name: "node_ip", Value: "*.*.*.2"},
			},
			MetricName: "req_count",
		},
	}, nil)
	defer patches.Reset()
	logs.InitLog("")
	var getter = APICInfo{}
	label, metrics := getter.GetResourceInfo()
	assert.Equal(t, 8, len(label))
	assert.Equal(t, 3, len(metrics))
}

func TestGetAllAPICInstances(t *testing.T) {
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
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
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
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

func TestBuildAppsInfo(t *testing.T) {
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	metricConf = make(map[string]MetricConf)
	metricConf["SYS.APIC"] = MetricConf{
		Resource: "service",
		DimMetricName: map[string][]string{
			"instance_id,app_id":        {"requests"},
			"instance_id,app_id,api_id": {"requests"},
		},
	}
	patches := getPatches()
	logs.InitLog("")

	apicClient := getAPICSClient()
	testAppId := "test-app-id"
	testAppName := "test-app_name"
	var testAppBindNum int32 = 1
	appOutputs := []gomonkey.OutputCell{
		{
			Values: gomonkey.Params{&model.ListAppsV2Response{
				Total: 1,
				Size:  1,
				Apps: &[]model.AppInfoWithBindNum{
					{
						Id:      &testAppId,
						Name:    &testAppName,
						Creator: &model.AppInfoWithBindNumCreator{},
						BindNum: &testAppBindNum,
					},
				},
			}, nil},
		},
		{
			Values: gomonkey.Params{&model.ListAppsV2Response{
				Total: 1,
				Size:  0,
				Apps:  &[]model.AppInfoWithBindNum{},
			}, nil},
		},
	}
	patches.ApplyMethodSeq(apicClient, "ListAppsV2", appOutputs)

	testApiId := "test-api-id"
	testApiName := "test-api_name"
	testApiGroupName := "test-api_group_name"
	apiOutputs := []gomonkey.OutputCell{
		{
			Values: gomonkey.Params{&model.ListApisBindedToAppV2Response{
				Total: 1,
				Size:  1,
				Auths: &[]model.ApiAuthInfo{
					{
						Id:        &testApiId,
						ApiName:   &testApiName,
						ApiId:     &testApiId,
						GroupName: &testApiGroupName,
					},
				},
			}, nil},
		},
		{
			Values: gomonkey.Params{&model.ListApisBindedToAppV2Response{
				Total: 1,
				Size:  0,
				Auths: &[]model.ApiAuthInfo{},
			}, nil},
		},
	}
	patches.ApplyMethodSeq(apicClient, "ListApisBindedToAppV2", apiOutputs)
	resourceInfos := map[string]labelInfo{}
	instanceInfo := labelInfo{
		Name:  []string{"instanceName", "eipAddress", "epId"},
		Value: []string{"test-instance-name", "test-eip", "test-ep-id"},
	}
	buildAppsInfo("test-instance-id", resourceInfos, instanceInfo)
	assert.Equal(t, 2, len(resourceInfos))
}

func TestBuildApiGroupsInfo(t *testing.T) {
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	metricConf = make(map[string]MetricConf)
	metricConf["SYS.APIC"] = MetricConf{
		Resource: "service",
		DimMetricName: map[string][]string{
			"instance_id,api_group_id": {"req_count"},
		},
	}
	testVersion := "test-version"
	testRemark := "test-remark"
	apiGroupOutputs := []gomonkey.OutputCell{
		{
			Values: gomonkey.Params{&model.ListApiGroupsV2Response{
				Total: 1,
				Size:  1,
				Groups: &[]model.ApiGroupInfo{
					{
						Id:       "test-group-id",
						Name:     "test-group-name",
						Status:   model.ApiGroupInfoStatus{},
						SlDomain: "test-domain",
						Remark:   &testRemark,
						Version:  &testVersion,
					},
				},
			}, nil},
		},
		{
			Values: gomonkey.Params{&model.ListApiGroupsV2Response{
				Total:  1,
				Size:   0,
				Groups: &[]model.ApiGroupInfo{},
			}, nil},
		},
	}
	patches := getPatches()
	logs.InitLog("")
	apicClient := getAPICSClient()
	patches.ApplyMethodSeq(apicClient, "ListApiGroupsV2", apiGroupOutputs)
	resourceInfos := map[string]labelInfo{}
	instanceInfo := labelInfo{
		Name:  []string{"instanceName", "eipAddress", "epId"},
		Value: []string{"test-instance-name", "test-eip", "test-ep-id"},
	}
	buildApiGroupsInfo("test-instance-id", resourceInfos, instanceInfo)
	assert.Equal(t, 1, len(resourceInfos))
}

func getStringPointer(str string) *string {
	return &str
}
