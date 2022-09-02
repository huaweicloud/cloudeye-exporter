package collector

import (
	"fmt"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

type ConnectInfo struct {
	ResourceBaseInfo
	ConnectProperties
}
type ConnectProperties struct {
	DeviceID  string `json:"device_id"`
	Type      string `json:"type"`
	PortType  string `json:"port_type"`
	Provider  string `json:"provider"`
	ProductID string `json:"product_id"`
}

type VifInfo struct {
	ResourceBaseInfo
	VifProperties
}
type VifProperties struct {
	DeviceId          string `json:"device_id"`
	RouteMode         string `json:"route_mode"`
	AddressFamily     string `json:"address_family"`
	Vlan              int    `json:"vlan"`
	RemoteGatewayV4IP string `json:"remote_gateway_v4_ip"`
	LocalGatewayV4IP  string `json:"local_gateway_v4_ip"`
}

var dcaasInfo serversInfo

type DCAASInfo struct{}

func (getter DCAASInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	dcaasInfo.Lock()
	defer dcaasInfo.Unlock()
	if dcaasInfo.LabelInfo == nil || time.Now().Unix() > dcaasInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.DCAAS")

		// direct_connect
		buildConnectsInfo(sysConfigMap, &filterMetrics, resourceInfos)

		// virtual_interface
		buildVifsInfo(sysConfigMap, &filterMetrics, resourceInfos)

		// virtual_gateway
		buildVgwsInfo(sysConfigMap, &filterMetrics, resourceInfos)

		dcaasInfo.LabelInfo = resourceInfos
		dcaasInfo.FilterMetrics = filterMetrics
		dcaasInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return dcaasInfo.LabelInfo, dcaasInfo.FilterMetrics
}

func buildVgwsInfo(sysConfigMap map[string][]string, filterMetrics *[]model.MetricInfoList, resourceInfos map[string]labelInfo) {
	vgws, err := getDcaasVgwFromRMS()
	if err != nil {
		return
	}
	for index := range vgws {
		if metricNames, ok := sysConfigMap["virtual_gateway_id"]; ok {
			metrics := buildSingleDimensionMetrics(metricNames, "SYS.DCAAS", "virtual_gateway_id", vgws[index].ID)
			*filterMetrics = append(*filterMetrics, metrics...)
			info := labelInfo{
				Name:  []string{"name", "epId"},
				Value: []string{vgws[index].Name, vgws[index].EpId},
			}
			keys, values := getTags(vgws[index].Tags)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
		}
	}
}

func buildVifsInfo(sysConfigMap map[string][]string, filterMetrics *[]model.MetricInfoList, resourceInfos map[string]labelInfo) {
	vifs, err := getDcaasVifFromRMS()
	if err != nil {
		return
	}
	for index := range vifs {
		if metricNames, ok := sysConfigMap["virtual_interface_id"]; ok {
			metrics := buildSingleDimensionMetrics(metricNames, "SYS.DCAAS", "virtual_interface_id", vifs[index].ID)
			*filterMetrics = append(*filterMetrics, metrics...)
			info := labelInfo{
				Name: []string{"name", "epId", "device_id", "route_mode", "address_family", "vlan", "remote_gateway_v4_ip", "local_gateway_v4_ip"},
				Value: []string{vifs[index].Name, vifs[index].EpId, vifs[index].DeviceId, vifs[index].RouteMode,
					vifs[index].AddressFamily, fmt.Sprintf("%d", vifs[index].Vlan), vifs[index].RemoteGatewayV4IP, vifs[index].LocalGatewayV4IP},
			}
			keys, values := getTags(vifs[index].Tags)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)

			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
		}
	}
}

func buildConnectsInfo(sysConfigMap map[string][]string, filterMetrics *[]model.MetricInfoList, resourceInfos map[string]labelInfo) {
	connects, err := getDcaasConnectsFromRMS()
	if err != nil {
		return
	}
	for index := range connects {
		if metricNames, ok := sysConfigMap["direct_connect_id"]; ok {
			metrics := buildSingleDimensionMetrics(metricNames, "SYS.DCAAS", "direct_connect_id", connects[index].ID)
			*filterMetrics = append(*filterMetrics, metrics...)
			info := labelInfo{
				Name: []string{"name", "epId", "device_id", "type", "port_type", "provider", "product_id"},
				Value: []string{connects[index].Name, connects[index].EpId, connects[index].DeviceID, connects[index].Type,
					connects[index].PortType, connects[index].Provider, connects[index].ProductID},
			}
			keys, values := getTags(connects[index].Tags)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
		}
	}
}

func getDcaasConnectsFromRMS() ([]ConnectInfo, error) {
	resp, err := listResources("dcaas", "directConnect")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of dcaas.directConnect, error: %s", err.Error())
		return nil, err
	}
	connects := make([]ConnectInfo, 0, len(resp))
	for _, resource := range resp {
		var connectProperties ConnectProperties
		err := fmtResourceProperties(resource.Properties, &connectProperties)
		if err != nil {
			logs.Logger.Errorf("fmt dcaas connect properties error: %s", err.Error())
			continue
		}
		connects = append(connects, ConnectInfo{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   *resource.Id,
				Name: *resource.Name,
				EpId: *resource.EpId,
				Tags: resource.Tags},
			ConnectProperties: connectProperties,
		})
	}
	return connects, nil
}

func getDcaasVifFromRMS() ([]VifInfo, error) {
	resp, err := listResources("dcaas", "vif")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of dcaas.vif, error: %s", err.Error())
		return nil, err
	}
	vifs := make([]VifInfo, 0, len(resp))
	for _, resource := range resp {
		var vifProperties VifProperties
		err := fmtResourceProperties(resource.Properties, &vifProperties)
		if err != nil {
			logs.Logger.Errorf("fmt vip properties error: %s", err.Error())
			continue
		}
		vifs = append(vifs, VifInfo{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   *resource.Id,
				Name: *resource.Name,
				EpId: *resource.EpId,
				Tags: resource.Tags,
			},
			VifProperties: vifProperties,
		})
	}
	return vifs, nil
}

func getDcaasVgwFromRMS() ([]ResourceBaseInfo, error) {
	resp, err := listResources("dcaas", "vgw")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of dcaas.vgw, error: %s", err.Error())
		return nil, err
	}
	vgws := make([]ResourceBaseInfo, len(resp))
	for index, resource := range resp {
		vgws[index].ID = *resource.Id
		vgws[index].Name = *resource.Name
		vgws[index].EpId = *resource.EpId
		vgws[index].Tags = resource.Tags
	}
	return vgws, nil
}
