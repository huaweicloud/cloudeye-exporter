package collector

import (
	"time"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
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
			return nil, nil
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

		wafInfo.LabelInfo = resourceInfos
		wafInfo.FilterMetrics = filterMetrics
		wafInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return wafInfo.LabelInfo, wafInfo.FilterMetrics
}

func getAllWafInstancesFromRMS() ([]ResourceBaseInfo, error) {
	resp, err := listResources("waf", "instance")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of waf.instance, error: %s", err.Error())
		return nil, err
	}
	wafInstances := make([]ResourceBaseInfo, 0, len(resp))
	for _, resource := range resp {
		wafInstances = append(wafInstances, ResourceBaseInfo{
			ID:   *resource.Id,
			Name: *resource.Name,
			EpId: *resource.EpId,
			Tags: resource.Tags,
		})
	}
	return wafInstances, nil
}
