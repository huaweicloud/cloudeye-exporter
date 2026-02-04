package collector

import (
	"fmt"
	"strings"
	"time"

	apig "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/apig/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/apig/v2/model"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var apicInfo serversInfo

type APICInfo struct{}

func (getter APICInfo) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]cesmodel.MetricInfoList, 0)
	apicInfo.Lock()
	defer apicInfo.Unlock()
	if apicInfo.LabelInfo == nil || time.Now().Unix() > apicInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.APIC")
		metricNames, ok := sysConfigMap["instance_id"]
		if !ok {
			logs.Logger.Warnf("Metric config is empty of SYS.APIC, dim_metric_name is instance_id")
			return apicInfo.LabelInfo, apicInfo.FilterMetrics
		}
		instances, err := getAllAPICInstances()
		if err != nil {
			logs.Logger.Errorf("Get all apic instances: %s", err.Error())
			return apicInfo.LabelInfo, apicInfo.FilterMetrics
		}
		for _, instance := range instances {
			metrics := buildSingleDimensionMetrics(metricNames, "SYS.APIC", "instance_id", *instance.Id)
			info := labelInfo{
				Name:  []string{"instanceName", "eipAddress", "epId"},
				Value: []string{getDefaultString(instance.InstanceName), getDefaultString(instance.EipAddress), getDefaultString(instance.EnterpriseProjectId)},
			}
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			buildApisInfo(*instance.Id, resourceInfos, info)
			nodeIps := buildNodeInfo(*instance.Id, resourceInfos, info)
			buildAppsInfo(*instance.Id, resourceInfos, info)
			buildApiGroupsInfo(*instance.Id, resourceInfos, info)
			buildPluginNodeInfo(*instance.Id, resourceInfos, info, nodeIps)
		}

		allMetrics, err := listAllMetrics("SYS.APIC")
		if err != nil {
			logs.Logger.Errorf("Get all apic metrics: %s", err.Error())
			return apicInfo.LabelInfo, apicInfo.FilterMetrics
		}

		for _, metricInfo := range allMetrics {
			resourceKey := GetResourceKeyFromMetricInfo(metricInfo)
			if resourceKey == "" {
				continue
			}
			if _, ok := resourceInfos[resourceKey]; !ok {
				continue
			}

			if IsMetricInfoInWhiteList(metricInfo) {
				filterMetrics = append(filterMetrics, metricInfo)
			}
		}

		apicInfo.LabelInfo = resourceInfos
		apicInfo.FilterMetrics = filterMetrics
		apicInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return apicInfo.LabelInfo, apicInfo.FilterMetrics
}

func buildPluginNodeInfo(instanceId string, resourceInfos map[string]labelInfo, instanceInfo labelInfo, nodeIps []string) {
	sysConfigMap := getMetricConfigMap("SYS.APIC")
	apiMetricNames, ok := sysConfigMap["instance_id,plugin_type,plugin_id,node_ip"]
	if !ok {
		logs.Logger.Warnf("Metric config is empty of SYS.APIC, dim_metric_name is instance_id,plugin_type,plugin_id,node_ip")
		return
	}
	plugins, err := getPlugins(instanceId)
	if err != nil {
		logs.Logger.Errorf("Get all apis of apic plugins: %s", err.Error())
		return
	}
	for _, plugin := range plugins {
		for _, nodeIp := range nodeIps {
			metrics := buildDimensionMetrics(apiMetricNames, "SYS.APIC",
				[]cesmodel.MetricsDimension{{Name: "instance_id", Value: instanceId}, {Name: "plugin_type", Value: plugin.PluginType.Value()},
					{Name: "plugin_id", Value: plugin.PluginId}, {Name: "node_ip", Value: nodeIp}})
			info := labelInfo{
				Name:  []string{"pluginName", "nodeIP"},
				Value: []string{plugin.PluginName, strings.ReplaceAll(nodeIp, "_", ".")},
			}
			info.Name = append(info.Name, instanceInfo.Name...)
			info.Value = append(info.Value, instanceInfo.Value...)
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
		}
	}
}

func buildApisInfo(instanceId string, resourceInfos map[string]labelInfo, instanceInfo labelInfo) {
	sysConfigMap := getMetricConfigMap("SYS.APIC")
	apiMetricNames, ok := sysConfigMap["instance_id,api_id"]
	if !ok {
		logs.Logger.Warnf("Metric config is empty of SYS.APIC, dim_metric_name is instance_id,api_id")
		return
	}
	apis, err := getApisOfInstances(instanceId)
	if err != nil {
		logs.Logger.Errorf("Get all apis of apic instances: %s", err.Error())
		return
	}
	for _, api := range apis {
		metrics := buildDimensionMetrics(apiMetricNames, "SYS.APIC",
			[]cesmodel.MetricsDimension{{Name: "instance_id", Value: instanceId}, {Name: "api_id", Value: *api.Id}})
		appInfo := labelInfo{
			Name:  []string{"apiName", "groupName", "groupId", "reqMethod", "reqUri"},
			Value: []string{api.Name, getDefaultString(api.GroupName), api.GroupId, api.ReqMethod.Value(), api.ReqUri},
		}
		appInfo.Name = append(appInfo.Name, instanceInfo.Name...)
		appInfo.Value = append(appInfo.Value, instanceInfo.Value...)
		resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = appInfo
	}
}

func buildAppsInfo(instanceId string, resourceInfos map[string]labelInfo, instanceInfo labelInfo) {
	sysConfigMap := getMetricConfigMap("SYS.APIC")
	apiMetricNames, ok := sysConfigMap["instance_id,app_id"]
	if !ok {
		logs.Logger.Warnf("Metric config is empty of SYS.APIC, dim_metric_name is instance_id,api_id")
		return
	}
	apps, err := getAppsOfInstances(instanceId)
	if err != nil {
		logs.Logger.Errorf("Get all apis of apic instances: %s", err.Error())
		return
	}
	for _, app := range apps {
		metrics := buildDimensionMetrics(apiMetricNames, "SYS.APIC",
			[]cesmodel.MetricsDimension{{Name: "instance_id", Value: instanceId}, {Name: "app_id", Value: *app.Id}})
		var creator string
		if app.Creator != nil {
			creator = app.Creator.Value()
		}
		var status string
		if app.Status != nil {
			status = fmt.Sprintf("%d", app.Status.Value())
		}
		var appType string
		if app.AppType != nil {
			appType = app.AppType.Value()
		}
		var apiCount string
		if app.BindNum != nil {
			apiCount = fmt.Sprintf("%d", *app.BindNum)
		}
		appInfo := labelInfo{
			Name:  []string{"appName", "creator", "status", "appType", "apiCount"},
			Value: []string{*app.Name, creator, status, appType, apiCount},
		}
		appInfo.Name = append(appInfo.Name, instanceInfo.Name...)
		appInfo.Value = append(appInfo.Value, instanceInfo.Value...)
		resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = appInfo
		buildApiBindedToAppInfo(instanceId, *app.Id, *app.Name, resourceInfos, instanceInfo)
	}
}

func buildApiBindedToAppInfo(instanceId, appId, appName string, resourceInfos map[string]labelInfo, instanceInfo labelInfo) {
	sysConfigMap := getMetricConfigMap("SYS.APIC")
	metricNames, ok := sysConfigMap["instance_id,app_id,api_id"]
	if !ok {
		logs.Logger.Warnf("Metric config is empty of SYS.APIC, dim_metric_name is instance_id,api_id")
		return
	}
	apis, err := getApisBindToApp(instanceId, appId)
	if err != nil {
		logs.Logger.Errorf("Get all apis binded to app instances: %s", err.Error())
		return
	}
	for _, api := range apis {
		metrics := buildDimensionMetrics(metricNames, "SYS.APIC",
			[]cesmodel.MetricsDimension{{Name: "instance_id", Value: instanceId}, {Name: "app_id", Value: appId}, {Name: "api_id", Value: getDefaultString(api.ApiId)}})
		apiInfo := labelInfo{
			Name:  []string{"appName", "apiName", "groupName"},
			Value: []string{appName, getDefaultString(api.ApiName), getDefaultString(api.GroupName)},
		}
		apiInfo.Name = append(apiInfo.Name, instanceInfo.Name...)
		apiInfo.Value = append(apiInfo.Value, instanceInfo.Value...)
		resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = apiInfo
	}
}

func buildApiGroupsInfo(instanceId string, resourceInfos map[string]labelInfo, instanceInfo labelInfo) {
	sysConfigMap := getMetricConfigMap("SYS.APIC")
	apiMetricNames, ok := sysConfigMap["instance_id,api_group_id"]
	if !ok {
		logs.Logger.Warnf("Metric config is empty of SYS.APIC, dim_metric_name is instance_id,api_id")
		return
	}
	apiGroups, err := getApiGroupsOfInstances(instanceId)
	if err != nil {
		logs.Logger.Errorf("Get all apis of apic instances: %s", err.Error())
		return
	}
	for _, group := range apiGroups {
		metrics := buildDimensionMetrics(apiMetricNames, "SYS.APIC",
			[]cesmodel.MetricsDimension{{Name: "instance_id", Value: instanceId}, {Name: "api_group_id", Value: group.Id}})
		groupInfo := labelInfo{
			Name:  []string{"groupName", "status", "version", "domain", "remark"},
			Value: []string{group.Name, fmt.Sprintf("%d", group.Status.Value()), getDefaultString(group.Version), group.SlDomain, getDefaultString(group.Remark)},
		}
		groupInfo.Name = append(groupInfo.Name, instanceInfo.Name...)
		groupInfo.Value = append(groupInfo.Value, instanceInfo.Value...)
		resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = groupInfo
	}
}

func buildNodeInfo(instanceId string, resourceInfos map[string]labelInfo, instanceInfo labelInfo) []string {
	sysConfigMap := getMetricConfigMap("SYS.APIC")
	apiMetricNames, ok := sysConfigMap["instance_id,node_ip"]
	var resultNodeIps []string
	if !ok {
		logs.Logger.Warnf("Metric config is empty of SYS.APIC, dim_metric_name is instance_id,node_ip")
		return resultNodeIps
	}
	instance, err := showDetailsOfInstanceV2(instanceId)
	if err != nil {
		logs.Logger.Errorf("Get all apis of apic instances: %s", err.Error())
		return resultNodeIps
	}
	nodeIps := make([]string, len(*instance.NodeIps.Livedata)+len(*instance.NodeIps.Shubao))
	nodeIps = append(nodeIps, *instance.NodeIps.Livedata...)
	nodeIps = append(nodeIps, *instance.NodeIps.Shubao...)

	for _, nodeIP := range *instance.NodeIps.Livedata {
		metrics := buildDimensionMetrics(apiMetricNames, "SYS.APIC",
			[]cesmodel.MetricsDimension{{Name: "instance_id", Value: instanceId}, {Name: "node_ip", Value: strings.ReplaceAll(nodeIP, ".", "_")}})
		resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = getNodeInfo(nodeIP, "livedata", instanceInfo)
		resultNodeIps = append(resultNodeIps, strings.ReplaceAll(nodeIP, ".", "_"))
	}

	for _, nodeIP := range *instance.NodeIps.Shubao {
		metrics := buildDimensionMetrics(apiMetricNames, "SYS.APIC",
			[]cesmodel.MetricsDimension{{Name: "instance_id", Value: instanceId}, {Name: "node_ip", Value: strings.ReplaceAll(nodeIP, ".", "_")}})
		resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = getNodeInfo(nodeIP, "shubao", instanceInfo)
		resultNodeIps = append(resultNodeIps, strings.ReplaceAll(nodeIP, ".", "_"))
	}
	return resultNodeIps
}

func getNodeInfo(nodeIP, nodeType string, instanceInfo labelInfo) labelInfo {
	appInfo := labelInfo{
		Name:  []string{"nodeIP", "nodeType"},
		Value: []string{nodeIP, nodeType},
	}
	appInfo.Name = append(appInfo.Name, instanceInfo.Name...)
	appInfo.Value = append(appInfo.Value, instanceInfo.Value...)
	return appInfo
}

func getAllAPICInstances() ([]model.RespInstanceBase, error) {
	limit := int32(500)
	offset := int64(0)
	var instances []model.RespInstanceBase
	for {
		request := &model.ListInstancesV2Request{Limit: &limit, Offset: &offset}
		response, err := getAPICSClient().ListInstancesV2(request)
		if err != nil {
			logs.Logger.Errorf("Failed to get all apic instances, error: %s", err.Error())
			return nil, err
		}
		if len(*response.Instances) == 0 {
			break
		}
		instances = append(instances, *response.Instances...)
		*request.Offset += int64(limit)
	}

	return instances, nil
}

func getApisOfInstances(instanceID string) ([]model.ApiInfoPerPage, error) {
	limit := int32(500)
	offset := int64(0)
	var apis []model.ApiInfoPerPage
	for {
		request := &model.ListApisV2Request{InstanceId: instanceID, Limit: &limit, Offset: &offset}
		response, err := getAPICSClient().ListApisV2(request)
		if err != nil {
			logs.Logger.Errorf("Failed to get all apis of apic instances, error: %s", err.Error())
			return nil, err
		}
		if len(*response.Apis) == 0 {
			break
		}
		apis = append(apis, *response.Apis...)
		*request.Offset += int64(limit)
	}
	return apis, nil
}

func getApiGroupsOfInstances(instanceID string) ([]model.ApiGroupInfo, error) {
	limit := int32(500)
	offset := int64(0)
	var apiGroups []model.ApiGroupInfo
	for {
		request := &model.ListApiGroupsV2Request{InstanceId: instanceID, Limit: &limit, Offset: &offset}
		response, err := getAPICSClient().ListApiGroupsV2(request)
		if err != nil {
			logs.Logger.Errorf("Failed to get all apis of apic instances, error: %s", err.Error())
			return nil, err
		}
		if len(*response.Groups) == 0 {
			break
		}
		apiGroups = append(apiGroups, *response.Groups...)
		*request.Offset += int64(limit)
	}
	return apiGroups, nil
}

func getAppsOfInstances(instanceID string) ([]model.AppInfoWithBindNum, error) {
	limit := int32(500)
	offset := int64(0)
	var apps []model.AppInfoWithBindNum
	for {
		request := &model.ListAppsV2Request{InstanceId: instanceID, Limit: &limit, Offset: &offset}
		response, err := getAPICSClient().ListAppsV2(request)
		if err != nil {
			logs.Logger.Errorf("Failed to get all apis of apic instances, error: %s", err.Error())
			return nil, err
		}
		if len(*response.Apps) == 0 {
			break
		}
		apps = append(apps, *response.Apps...)
		*request.Offset += int64(limit)
	}
	return apps, nil
}

func getApisBindToApp(instanceId, appId string) ([]model.ApiAuthInfo, error) {
	limit := int32(500)
	offset := int64(0)
	var apis []model.ApiAuthInfo
	for {
		request := &model.ListApisBindedToAppV2Request{InstanceId: instanceId, AppId: appId, Limit: &limit, Offset: &offset}
		response, err := getAPICSClient().ListApisBindedToAppV2(request)
		if err != nil {
			logs.Logger.Errorf("Failed to get api binded to app, error: %s", err.Error())
			return nil, err
		}
		if len(*response.Auths) == 0 {
			break
		}
		apis = append(apis, *response.Auths...)
		*request.Offset += int64(limit)
	}
	return apis, nil
}

func getPlugins(instanceId string) ([]model.PluginInfo, error) {
	limit := int32(500)
	offset := int64(0)
	var plugins []model.PluginInfo
	for {
		request := &model.ListPluginsRequest{InstanceId: instanceId, Limit: &limit, Offset: &offset}
		response, err := getAPICSClient().ListPlugins(request)
		if err != nil {
			logs.Logger.Errorf("Failed to get api binded to app, error: %s", err.Error())
			return nil, err
		}
		if len(*response.Plugins) == 0 {
			break
		}
		plugins = append(plugins, *response.Plugins...)
		*request.Offset += int64(limit)
	}
	return plugins, nil
}

func showDetailsOfInstanceV2(instanceID string) (*model.ShowDetailsOfInstanceV2Response, error) {
	request := &model.ShowDetailsOfInstanceV2Request{InstanceId: instanceID}
	response, err := getAPICSClient().ShowDetailsOfInstanceV2(request)
	if err != nil {
		logs.Logger.Errorf("Failed to get instance info[%s], error: %s", instanceID, err.Error())
		return nil, err
	}
	return response, nil
}

func getAPICSClient() *apig.ApigClient {
	return apig.NewApigClient(apig.ApigClientBuilder().WithCredential(
		authCredentialMap[conf.AuthMode](RegionServiceType)).
		WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
		WithEndpoint(getEndpoint("apig", "v2")).Build())
}
