package collector

import (
	"time"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

var vpcEPInfo serversInfo

type VPCEPInfo struct{}

func (getter VPCEPInfo) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]cesmodel.MetricInfoList, 0)
	vpcEPInfo.Lock()
	defer vpcEPInfo.Unlock()
	if vpcEPInfo.LabelInfo == nil || time.Now().Unix() > vpcEPInfo.TTL {
		vpcEpEndpoints, err := getVpcEpEndpoints()
		if err != nil {
			logs.Logger.Error("Get all vpcep endpoints error:", err.Error())
			return vpcEPInfo.LabelInfo, vpcEPInfo.FilterMetrics
		}
		sysConfigMap := getMetricConfigMap("SYS.VPCEP")
		if metricNames, ok := sysConfigMap["ep_instance_id"]; ok {
			for _, endpoint := range vpcEpEndpoints {
				metrics := buildSingleDimensionMetrics(metricNames, "SYS.VPCEP", "ep_instance_id", endpoint.ID)
				filterMetrics = append(filterMetrics, metrics...)
				info := labelInfo{
					Name:  []string{"name", "epId", "ip", "vpcId"},
					Value: []string{endpoint.Name, endpoint.EpId, endpoint.IP, endpoint.VpcId},
				}
				keys, values := getTags(endpoint.Tags)
				info.Name = append(info.Name, keys...)
				info.Value = append(info.Value, values...)
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			}
		}
		vpcEPInfo.LabelInfo = resourceInfos
		vpcEPInfo.FilterMetrics = filterMetrics
		vpcEPInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return vpcEPInfo.LabelInfo, vpcEPInfo.FilterMetrics
}

type VpcEpEndpoint struct {
	ResourceBaseInfo
	EndpointProperties
}
type EndpointProperties struct {
	IP    string `json:"ip"`
	VpcId string `json:"vpc_id"`
}

func getVpcEpEndpoints() ([]VpcEpEndpoint, error) {
	endpoints, err := listResources("vpcep", "endpoints")
	if err != nil {
		return nil, err
	}
	vpcEpEndpoints := make([]VpcEpEndpoint, 0, len(endpoints))
	for _, endpoint := range endpoints {
		var endpointProperties EndpointProperties
		err := fmtResourceProperties(endpoint.Properties, &endpointProperties)
		if err != nil {
			logs.Logger.Errorf("fmt endpoint properties error: %s", err.Error())
			continue
		}
		vpcEpEndpoints = append(vpcEpEndpoints, VpcEpEndpoint{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   *endpoint.Id,
				Name: *endpoint.Name,
				EpId: *endpoint.EpId,
				Tags: endpoint.Tags,
			},
			EndpointProperties: endpointProperties,
		})
	}
	return vpcEpEndpoints, nil
}
