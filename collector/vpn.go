package collector

import (
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var vpnInfo serversInfo

type VPNInfo struct{}

func (getter VPNInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	vpnInfo.Lock()
	defer vpnInfo.Unlock()
	if vpnInfo.LabelInfo == nil || time.Now().Unix() > vpnInfo.TTL {
		buildIpsecConnectionsInfo(&filterMetrics, resourceInfos)
		buildConnectionsInfo(&filterMetrics, resourceInfos)
		vpnInfo.LabelInfo = resourceInfos
		vpnInfo.FilterMetrics = filterMetrics
		vpnInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return vpnInfo.LabelInfo, vpnInfo.FilterMetrics
}

func buildIpsecConnectionsInfo(filterMetrics *[]model.MetricInfoList, resourceInfos map[string]labelInfo) {
	metricNames := getMetricConfigMap("SYS.VPN")["connection_id"]
	if len(metricNames) == 0 {
		logs.Logger.Debugf("metric names is empty: %s", "connection_id")
		return
	}
	for _, ipsecConnection := range getAllIpsecConnectionsFromRMS() {
		metrics := buildSingleDimensionMetrics(metricNames, "SYS.VPN", "connection_id", ipsecConnection.ID)
		*filterMetrics = append(*filterMetrics, metrics...)
		info := labelInfo{
			Name:  []string{"name", "epId", "peer_address"},
			Value: []string{ipsecConnection.Name, ipsecConnection.EpId, ipsecConnection.PeerAddress},
		}
		keys, values := getTags(ipsecConnection.Tags)
		info.Name = append(info.Name, keys...)
		info.Value = append(info.Value, values...)
		resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
	}
}

func buildConnectionsInfo(filterMetrics *[]model.MetricInfoList, resourceInfos map[string]labelInfo) {
	metricNames := getMetricConfigMap("SYS.VPN")["vpn_connection_id"]
	if len(metricNames) == 0 {
		logs.Logger.Debugf("metric names is empty: %s", "vpn_connection_id")
		return
	}
	for _, connection := range getAllConnectionsFromRMS() {
		metrics := buildSingleDimensionMetrics(metricNames, "SYS.VPN", "vpn_connection_id", connection.ID)
		*filterMetrics = append(*filterMetrics, metrics...)
		info := labelInfo{
			Name:  []string{"name", "epId", "peer_address"},
			Value: []string{connection.Name, connection.EpId, connection.PeerAddress},
		}
		keys, values := getTags(connection.Tags)
		info.Name = append(info.Name, keys...)
		info.Value = append(info.Value, values...)
		resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
	}
}

type ConnectionInfo struct {
	ResourceBaseInfo
	PeerAddress string
}

type ConnectionProperties struct {
	PeerAddress string `json:"peer_address"`
}

func getAllIpsecConnectionsFromRMS() []ConnectionInfo {
	resp, err := listResources("vpnaas", "ipsec-site-connections")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of vpnaas.ipsec-site-connections, error: %s", err.Error())
		return nil
	}
	connections := make([]ConnectionInfo, 0, len(resp))
	for _, resource := range resp {
		var properties ConnectionProperties
		err := fmtResourceProperties(resource.Properties, &properties)
		if err != nil {
			logs.Logger.Errorf("fmt vpn properties error: %s", err.Error())
			continue
		}
		connections = append(connections, ConnectionInfo{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   *resource.Id,
				Name: *resource.Name,
				EpId: *resource.EpId,
				Tags: resource.Tags,
			},
			PeerAddress: properties.PeerAddress,
		})
	}
	return connections
}

func getAllConnectionsFromRMS() []ConnectionInfo {
	resp, err := listResources("vpnaas", "vpnConnections")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of vpnaas.vpnConnections, error: %s", err.Error())
		return nil
	}
	connections := make([]ConnectionInfo, 0, len(resp))
	for _, resource := range resp {
		var properties ConnectionProperties
		err := fmtResourceProperties(resource.Properties, &properties)
		if err != nil {
			logs.Logger.Errorf("fmt vpn properties error: %s", err.Error())
			continue
		}
		connections = append(connections, ConnectionInfo{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   *resource.Id,
				Name: *resource.Name,
				EpId: *resource.EpId,
				Tags: resource.Tags,
			},
			PeerAddress: properties.PeerAddress,
		})
	}
	return connections
}
