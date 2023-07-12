package collector

import (
	"time"

	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

var asInfo serversInfo

type ASInfo struct{}

func (getter ASInfo) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]cesmodel.MetricInfoList, 0)
	asInfo.Lock()
	defer asInfo.Unlock()
	if asInfo.LabelInfo == nil || time.Now().Unix() > asInfo.TTL {
		if metricNames, ok := getMetricConfigMap("SYS.AS")["AutoScalingGroup"]; ok {
			if scalingGroups, err := getAllASFromRMS(); err == nil {
				for _, scalingGroup := range scalingGroups {
					metrics := buildSingleDimensionMetrics(metricNames, "SYS.AS", "AutoScalingGroup", scalingGroup.ID)
					filterMetrics = append(filterMetrics, metrics...)
					info := labelInfo{
						Name:  []string{"name", "epId"},
						Value: []string{scalingGroup.Name, scalingGroup.EpId},
					}
					keys, values := getTags(scalingGroup.Tags)
					info.Name = append(info.Name, keys...)
					info.Value = append(info.Value, values...)
					resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
				}
			}
		}
		asInfo.LabelInfo = resourceInfos
		asInfo.FilterMetrics = filterMetrics
		asInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return asInfo.LabelInfo, asInfo.FilterMetrics
}

func getAllASFromRMS() ([]ResourceBaseInfo, error) {
	return getResourcesBaseInfoFromRMS("as", "scalingGroups")
}
