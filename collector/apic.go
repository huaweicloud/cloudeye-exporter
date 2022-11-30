package collector

import (
	"time"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	apig "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/apig/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/apig/v2/model"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
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
			filterMetrics = append(filterMetrics, metrics...)
			info := labelInfo{
				Name:  []string{"instanceName", "eipAddress", "epId"},
				Value: []string{getDefaultString(instance.InstanceName), getDefaultString(instance.EipAddress), getDefaultString(instance.EnterpriseProjectId)},
			}
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			buildApisInfo(*instance.Id, resourceInfos, &filterMetrics, info)
		}
		apicInfo.LabelInfo = resourceInfos
		apicInfo.FilterMetrics = filterMetrics
		apicInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return apicInfo.LabelInfo, apicInfo.FilterMetrics
}

func buildApisInfo(instanceId string, resourceInfos map[string]labelInfo, filterMetrics *[]cesmodel.MetricInfoList, instanceInfo labelInfo) {
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
		*filterMetrics = append(*filterMetrics, metrics...)
		appInfo := labelInfo{
			Name:  []string{"appName", "groupName", "groupId", "reqMethod", "reqUri"},
			Value: []string{api.Name, getDefaultString(api.GroupName), getDefaultString(api.GroupName), api.ReqMethod.Value(), api.ReqUri},
		}
		appInfo.Name = append(appInfo.Name, instanceInfo.Name...)
		appInfo.Value = append(appInfo.Value, instanceInfo.Value...)
		resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = appInfo
	}
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

func getAPICSClient() *apig.ApigClient {
	return apig.NewApigClient(apig.ApigClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(true)).
		WithEndpoint(getEndpoint("apig", "v2")).Build())
}
