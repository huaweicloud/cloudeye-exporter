package collector

import (
	"time"

	aad "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/aad/v1"
	aadmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/aad/v1/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var ddosInfo serversInfo

type DDOSInfo struct{}

func getDDOSClient() *aad.AadClient {
	return aad.NewAadClient(aad.AadClientBuilder().WithCredential(
		authCredentialMap[conf.AuthMode](GlobalServiceType)).
		WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
		WithEndpoint(getEndpoint("ddos", "v1")).Build())
}

func (getter DDOSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	ddosInfo.Lock()
	defer ddosInfo.Unlock()
	if ddosInfo.LabelInfo == nil || time.Now().Unix() > ddosInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.DDOS")
		if sysConfigMap == nil {
			return ddosInfo.LabelInfo, ddosInfo.FilterMetrics
		}

		// 查询aad实例并封装资源+指标查询条件
		instances, err := getResourcesBaseInfoFromRMS("aad", "instances")
		if err != nil {
			logs.Logger.Errorf("Get All DDos Instances error: %s", err.Error())
			return ddosInfo.LabelInfo, ddosInfo.FilterMetrics
		}
		var instanceMetrics []model.MetricInfoList
		instanceMetrics, resourceInfos = buildAadInstanceMetrics(sysConfigMap, instances, resourceInfos)
		filterMetrics = append(filterMetrics, instanceMetrics...)

		// 查询cnad防护包资源+指标
		var packageMetrics []model.MetricInfoList
		packageMetrics, resourceInfos = getCnadPackageResources(sysConfigMap, resourceInfos)
		filterMetrics = append(filterMetrics, packageMetrics...)

		// 获取防护包IP子维度相关指标+资源]
		var packageIpMetrics []model.MetricInfoList
		packageIpMetrics, resourceInfos = getCnadProtectedIpResources(sysConfigMap, resourceInfos)
		filterMetrics = append(filterMetrics, packageIpMetrics...)

		ddosInfo.LabelInfo = resourceInfos
		ddosInfo.FilterMetrics = filterMetrics
		ddosInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return ddosInfo.LabelInfo, ddosInfo.FilterMetrics
}

func buildAadInstanceMetrics(sysConfigMap map[string][]string, resources []ResourceBaseInfo,
	resourceInfos map[string]labelInfo) ([]model.MetricInfoList, map[string]labelInfo) {
	var filterMetrics []model.MetricInfoList
	for _, resource := range resources {
		metricNames, ok := sysConfigMap["instance_id"]
		if !ok {
			continue
		}
		metrics := buildSingleDimensionMetrics(metricNames, "SYS.DDOS", "instance_id", resource.ID)
		filterMetrics = append(filterMetrics, metrics...)
		info := labelInfo{
			Name:  []string{"name", "epId"},
			Value: []string{resource.Name, resource.EpId},
		}
		keys, values := getTags(resource.Tags)
		info.Name = append(info.Name, keys...)
		info.Value = append(info.Value, values...)
		resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
	}
	return filterMetrics, resourceInfos
}

func getCnadPackageResources(sysConfigMap map[string][]string, resourceInfos map[string]labelInfo) ([]model.MetricInfoList, map[string]labelInfo) {
	var filterMetrics []model.MetricInfoList
	request := &aadmodel.ListPackageRequest{}
	packageResponse, err := getDDOSClient().ListPackage(request)
	if err != nil {
		logs.Logger.Errorf("Get cnad package resource error: %s", err.Error())
		return filterMetrics, resourceInfos
	}
	if packageResponse == nil || packageResponse.HttpStatusCode != 200 {
		logs.Logger.Errorf("Get cnad package resource return http code: %d", packageResponse.HttpStatusCode)
		return filterMetrics, resourceInfos
	}

	metricNames, ok := sysConfigMap["package"]
	if !ok {
		return filterMetrics, resourceInfos
	}
	for _, packageItem := range *packageResponse.Items {
		packageId := packageItem.PackageId
		packageName := packageItem.PackageName
		metrics := buildSingleDimensionMetrics(metricNames, "SYS.DDOS", "package", packageId)
		info := labelInfo{
			Name:  []string{"name", "region"},
			Value: []string{packageName, packageItem.RegionId},
		}
		filterMetrics = append(filterMetrics, metrics...)
		resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
	}
	return filterMetrics, resourceInfos
}

func getCnadProtectedIpResources(sysConfigMap map[string][]string, resourceInfos map[string]labelInfo) ([]model.
	MetricInfoList, map[string]labelInfo) {
	var filterMetrics []model.MetricInfoList
	request := &aadmodel.ListProtectedIpRequest{}
	protectIpResponse, err := getDDOSClient().ListProtectedIp(request)
	if err != nil {
		logs.Logger.Errorf("Get cnad protected ip resource error: %s", err.Error())
		return filterMetrics, resourceInfos
	}

	if protectIpResponse == nil || protectIpResponse.HttpStatusCode != 200 {
		logs.Logger.Errorf("Get cnad protected ip resource return http code: %d", protectIpResponse.HttpStatusCode)
		return filterMetrics, resourceInfos
	}

	metricNames, ok := sysConfigMap["package,ip"]
	if !ok {
		return filterMetrics, resourceInfos
	}

	for _, protectedIpItem := range *protectIpResponse.Items {
		packageId := protectedIpItem.PackageId
		packageName := protectedIpItem.PackageName
		protectedIpId := protectedIpItem.Id
		protectedIp := protectedIpItem.Ip
		metrics := buildDimensionMetrics(metricNames, "SYS.DDOS", []model.MetricsDimension{
			{Name: "package", Value: packageId},
			{Name: "ip", Value: protectedIpId},
		})

		// 将父维度资源中的防护包名+企业项目设置到资源标签中
		info := labelInfo{
			Name:  []string{"package_name", "ip_address", "region"},
			Value: []string{packageName, protectedIp, protectedIpItem.Region},
		}
		filterMetrics = append(filterMetrics, metrics...)
		resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
	}
	return filterMetrics, resourceInfos
}
