package collector

import (
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var ddosInfo serversInfo

type DDOSInfo struct{}

func (getter DDOSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	ddosInfo.Lock()
	defer ddosInfo.Unlock()
	if ddosInfo.LabelInfo == nil || time.Now().Unix() > ddosInfo.TTL {
		instances, err := getAllDDosInstancesFromRMS()
		if err != nil {
			logs.Logger.Errorf("Get All DDos Instances error: %s", err.Error())
			return ddosInfo.LabelInfo, ddosInfo.FilterMetrics
		}
		sysConfigMap := getMetricConfigMap("SYS.DDOS")
		for _, instance := range instances {
			if metricNames, ok := sysConfigMap["instance_id"]; ok {
				metrics := buildSingleDimensionMetrics(metricNames, "SYS.DDOS", "instance_id", instance.ID)
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
		ddosInfo.LabelInfo = resourceInfos
		ddosInfo.FilterMetrics = filterMetrics
		ddosInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return ddosInfo.LabelInfo, ddosInfo.FilterMetrics
}

func getAllDDosInstancesFromRMS() ([]ResourceBaseInfo, error) {
	return getResourcesBaseInfoFromRMS("aad", "instances")
}
