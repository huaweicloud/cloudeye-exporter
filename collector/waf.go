package collector

import (
	"net/http"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	waf "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/waf/v1"
	wafModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/waf/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var wafInfo serversInfo

type WAFInfo struct{}

func (getter WAFInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	wafInfo.Lock()
	defer wafInfo.Unlock()
	if wafInfo.LabelInfo == nil || time.Now().Unix() > wafInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.WAF")
		wafInstances, err := getAllWafInstancesFromRMS()
		if err != nil {
			logs.Logger.Errorf("Failed to get all waf instances, error: %s", err.Error())
			return wafInfo.LabelInfo, wafInfo.FilterMetrics
		}

		for _, instance := range wafInstances {
			if metricNames, ok := sysConfigMap["waf_instance_id"]; ok {
				metrics := buildSingleDimensionMetrics(metricNames, "SYS.WAF", "waf_instance_id", instance.ID)
				filterMetrics = append(filterMetrics, metrics...)
				info := labelInfo{
					Name:  []string{"name", "epId"},
					Value: []string{instance.Name, instance.EpId},
				}
				keys, values := getTags(instance.Tags)
				info.Name = append(info.Name, keys...)
				info.Value = append(info.Value, values...)
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			}
		}

		premiumWafInstances, err := getAllPremiumWafInstances()
		if err != nil {
			return wafInfo.LabelInfo, wafInfo.FilterMetrics
		}
		for _, instance := range premiumWafInstances {
			if metricNames, ok := sysConfigMap["instance_id"]; ok {
				metrics := buildSingleDimensionMetrics(metricNames, "SYS.WAF", "instance_id", *instance.Id)
				filterMetrics = append(filterMetrics, metrics...)
				info := labelInfo{
					Name:  []string{"name"},
					Value: []string{*instance.InstanceName},
				}
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			}
		}

		wafInfo.LabelInfo = resourceInfos
		wafInfo.FilterMetrics = filterMetrics
		wafInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return wafInfo.LabelInfo, wafInfo.FilterMetrics
}

func getAllWafInstancesFromRMS() ([]ResourceBaseInfo, error) {
	return getResourcesBaseInfoFromRMS("waf", "instance")
}

func getWAFClient() *waf.WafClient {
	return waf.NewWafClient(waf.WafClientBuilder().WithCredential(
		authCredentialMap[conf.AuthMode](RegionServiceType)).
		WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
		WithEndpoint(getEndpoint("waf", "v1")).Build())
}

func getAllPremiumWafInstances() ([]wafModel.ListInstance, error) {
	allInstances := make([]wafModel.ListInstance, 0)
	epIds := getEpIdRequestPart()
	for _, epId := range epIds {
		instancesByEpId, err := getAllPremiumWafInstancesByEpId(epId)
		if err != nil {
			return nil, err
		}
		allInstances = append(allInstances, instancesByEpId...)
	}
	return allInstances, nil
}

func getAllPremiumWafInstancesByEpId(epId string) ([]wafModel.ListInstance, error) {
	result := make([]wafModel.ListInstance, 0)
	pageSize := int32(100)
	page := int32(1)
	req := &wafModel.ListInstanceRequest{
		Page:                &page,
		Pagesize:            &pageSize,
		EnterpriseProjectId: &epId,
	}
	for {
		resp, err := getWAFClient().ListInstance(req)
		if err != nil {
			logs.Logger.Errorf("Get all premiumWafInstances err, err is : %s", err.Error())
			return nil, err
		}
		if resp.HttpStatusCode != http.StatusOK {
			logs.Logger.Errorf("Get all premiumWafInstances HttpStatusCode is %d", resp.HttpStatusCode)
			break
		}
		if len(*resp.Items) == 0 {
			break
		}
		instanceInfo := *resp.Items
		result = append(result, instanceInfo...)
		*req.Page += 1
	}
	return result, nil
}

func (getter WAFInfo) resetResourceInfo() {
	wafInfo.LabelInfo = nil
}
