package collector

import (
	"time"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

var bmsInfo serversInfo

type BMSInfo struct{}

func (getter BMSInfo) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]cesmodel.MetricInfoList, 0)
	bmsInfo.Lock()
	defer bmsInfo.Unlock()
	if bmsInfo.LabelInfo == nil || time.Now().Unix() > bmsInfo.TTL {
		services, err := getAllServerFromRMS("bms", "servers")
		if err != nil {
			logs.Logger.Error("Get all bms server error:", err.Error())
			return bmsInfo.LabelInfo, bmsInfo.FilterMetrics
		}
		sysConfigMap := getMetricConfigMap("SYS.BMS")
		if metricNames, ok := sysConfigMap["instance_id"]; ok {
			for _, instance := range services {
				loadAgentDimensions(instance.ID)
				metrics := buildSingleDimensionMetrics(metricNames, "SYS.BMS", "instance_id", instance.ID)
				filterMetrics = append(filterMetrics, metrics...)
				info := labelInfo{
					Name:  []string{"name", "epId", "ip"},
					Value: []string{instance.Name, instance.EpId, instance.IP},
				}
				keys, values := getTags(instance.Tags)
				info.Name = append(info.Name, keys...)
				info.Value = append(info.Value, values...)
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			}
		}
		bmsInfo.LabelInfo = resourceInfos
		bmsInfo.FilterMetrics = filterMetrics
		bmsInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return bmsInfo.LabelInfo, bmsInfo.FilterMetrics
}

type SERVICEBMSInfo struct{}

func (getter SERVICEBMSInfo) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	bmsInfo.Lock()
	defer bmsInfo.Unlock()
	return bmsInfo.LabelInfo, nil
}
