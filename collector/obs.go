package collector

import (
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	iam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var obsServerInfo newServersInfo

type OBSInfo struct{}

func (getter OBSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	var filteredMetrics []model.MetricInfoList
	obsServerInfo.Lock()
	defer obsServerInfo.Unlock()
	if obsServerInfo.LabelInfo == nil || time.Now().Unix() > obsServerInfo.TTL {
		allMetrics, err := listAllMetrics("SYS.OBS")
		if err != nil {
			logs.Logger.Errorf("[%s] Get all metrics of SYS.OBS failed, error is: %s", err.Error())
			return obsServerInfo.LabelInfo, obsServerInfo.GetFilteredMetrics()
		}
		for _, metricInfo := range allMetrics {
			if IsMetricInfoInWhiteList(metricInfo) {
				filteredMetrics = append(filteredMetrics, metricInfo)
			}
		}
		metricDimMap := map[string]cesmodel.MetricsDimension{}
		for _, metric := range filteredMetrics {
			for _, dimension := range metric.Dimensions {
				metricDimMap[dimension.Value] = dimension
			}
		}

		services, err := getAllServerFromRMS("obs", "buckets")
		if err != nil {
			logs.Logger.Error("Get all obs server from RMS failed, error is:", err.Error())
			return obsServerInfo.LabelInfo, obsServerInfo.GetFilteredMetrics()
		}
		for _, instance := range services {
			if _, ok := metricDimMap[instance.Name]; ok {
				info := labelInfo{
					Name:  []string{"epId"},
					Value: []string{instance.EpId},
				}
				keys, values := getTags(instance.Tags)
				info.Name = append(info.Name, keys...)
				info.Value = append(info.Value, values...)
				resourceInfos[metricDimMap[instance.Name].Value] = info
			}
		}

		userInfoMap := getUserInfoFromIAM()
		if userInfoMap != nil {
			for _, metric := range filteredMetrics {
				tenantId := getTenantId(metric)
				if tenantId == "" {
					continue
				}
				if userName, ok := userInfoMap[tenantId]; ok {
					info := labelInfo{
						Name:  []string{"username"},
						Value: []string{userName},
					}
					resourceInfos[GetResourceKeyFromMetricInfo(metric)] = info
				}
			}
		}
		obsServerInfo.LabelInfo = resourceInfos
		obsServerInfo.FilterMetrics = mergeMetricsWithCache(filteredMetrics, obsServerInfo.FilterMetrics)
		obsServerInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return obsServerInfo.LabelInfo, obsServerInfo.GetFilteredMetrics()
}

func getTenantId(metric cesmodel.MetricInfoList) string {
	for _, dim := range metric.Dimensions {
		if dim.Name == "tenant_id" {
			return dim.Value
		}
	}
	return ""
}

func getUserInfoFromIAM() map[string]string {
	userInfoMap := make(map[string]string)
	request := &iam.KeystoneListAuthDomainsRequest{}
	response, err := GetIAMClient().KeystoneListAuthDomains(request)
	if err != nil {
		logs.Logger.Errorf("Get user info from IAM failed, error is : %s", err.Error())
		return nil
	}
	if len(*response.Domains) == 0 {
		logs.Logger.Info("Get user info from IAM is empty")
		return nil
	}
	for _, domain := range *response.Domains {
		userInfoMap[domain.Id] = domain.Name
	}
	return userInfoMap
}
