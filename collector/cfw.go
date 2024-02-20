package collector

import (
	"time"

	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var cfwServerInfo serversInfo

type CFWInfo struct{}

func (cfw CFWInfo) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]cesmodel.MetricInfoList, 0)
	cfwServerInfo.Lock()
	defer cfwServerInfo.Unlock()
	if cfwServerInfo.LabelInfo == nil || time.Now().Unix() > cfwServerInfo.TTL {
		cfwConfigMap := getMetricConfigMap("SYS.CFW")
		if cfwConfigMap == nil {
			logs.Logger.Warn("Metric config is nil.")
			return cfwServerInfo.LabelInfo, cfwServerInfo.FilterMetrics
		}

		if _, ok := cfwConfigMap["fw_instance_id"]; !ok {
			logs.Logger.Warn("Metric config is nil of SYS.CFW of fw_instance_id.")
			return cfwServerInfo.LabelInfo, cfwServerInfo.FilterMetrics
		}

		metricNames := cfwConfigMap["fw_instance_id"]
		if len(metricNames) == 0 {
			logs.Logger.Warn("Metric config is empty of SYS.CFW of fw_instance_id.")
			return cfwServerInfo.LabelInfo, cfwServerInfo.FilterMetrics
		}

		servers, err := getResourcesBaseInfoFromRMS("cfw", "cfw_instance")
		if err != nil {
			logs.Logger.Errorf("Get resource base info from RMS Server error:", err.Error())
			return ecsInfo.LabelInfo, ecsInfo.FilterMetrics
		}

		for _, server := range servers {
			metrics := buildSingleDimensionMetrics(metricNames, "SYS.CFW", "fw_instance_id", server.ID)
			filterMetrics = append(filterMetrics, metrics...)
			info := labelInfo{
				Name:  []string{"name", "ep_id"},
				Value: []string{server.Name, server.EpId},
			}
			keys, values := getTags(server.Tags)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
		}

		cfwServerInfo.LabelInfo = resourceInfos
		cfwServerInfo.FilterMetrics = filterMetrics
		cfwServerInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return cfwServerInfo.LabelInfo, cfwServerInfo.FilterMetrics
}
