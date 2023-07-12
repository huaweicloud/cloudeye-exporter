package collector

import (
	"time"

	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

var natInfo serversInfo

type NATInfo struct{}

func (getter NATInfo) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]cesmodel.MetricInfoList, 0)
	natInfo.Lock()
	defer natInfo.Unlock()
	if natInfo.LabelInfo == nil || time.Now().Unix() > natInfo.TTL {
		if metricNames, ok := getMetricConfigMap("SYS.NAT")["nat_gateway_id"]; ok {
			if natGateways, err := getAllNatFromRMS(); err == nil {
				for _, natGateway := range natGateways {
					metrics := buildSingleDimensionMetrics(metricNames, "SYS.NAT", "nat_gateway_id", natGateway.ID)
					filterMetrics = append(filterMetrics, metrics...)
					info := labelInfo{
						Name:  []string{"name", "epId"},
						Value: []string{natGateway.Name, natGateway.EpId},
					}
					keys, values := getTags(natGateway.Tags)
					info.Name = append(info.Name, keys...)
					info.Value = append(info.Value, values...)
					resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
				}
			}
		}
		natInfo.LabelInfo = resourceInfos
		natInfo.FilterMetrics = filterMetrics
		natInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return natInfo.LabelInfo, natInfo.FilterMetrics
}

func getAllNatFromRMS() ([]ResourceBaseInfo, error) {
	return getResourcesBaseInfoFromRMS("nat", "natGateways")
}
