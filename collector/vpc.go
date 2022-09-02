package collector

import (
	"fmt"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

type Bandwidth struct {
	ResourceBaseInfo
	BandwidthProperties
}

type BandwidthProperties struct {
	BandwidthType string `json:"bandwidthType"`
	Type          string `json:"type"`
}

type PublicIp struct {
	ResourceBaseInfo
	PublicIpProperties
}

type PublicIpProperties struct {
	NetworkType     string `json:"networkType"`
	PublicIpAddress string `json:"publicIpAddress"`
	IpVersion       int    `json:"ipVersion"`
}

var vpcInfo serversInfo

type VPCInfo struct{}

func (getter VPCInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	vpcInfo.Lock()
	defer vpcInfo.Unlock()
	if vpcInfo.LabelInfo == nil || time.Now().Unix() > vpcInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.VPC")

		// bandwidths
		buildBandwidthsInfo(sysConfigMap, &filterMetrics, resourceInfos)

		// publicips
		buildPublicipsInfo(sysConfigMap, &filterMetrics, resourceInfos)

		vpcInfo.LabelInfo = resourceInfos
		vpcInfo.FilterMetrics = filterMetrics
		vpcInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return vpcInfo.LabelInfo, vpcInfo.FilterMetrics
}

func buildPublicipsInfo(sysConfigMap map[string][]string, filterMetrics *[]model.MetricInfoList, resourceInfos map[string]labelInfo) {
	publicips, err := getAllPublicIpFromRMS()
	if err != nil {
		return
	}
	for _, publicip := range publicips {
		if metricNames, ok := sysConfigMap["publicip_id"]; ok {
			metrics := buildSingleDimensionMetrics(metricNames, "SYS.VPC", "publicip_id", publicip.ID)
			*filterMetrics = append(*filterMetrics, metrics...)
			info := labelInfo{
				Name:  []string{"name", "epId", "networkType", "publicIpAddress", "ipVersion"},
				Value: []string{publicip.Name, publicip.EpId, publicip.NetworkType, publicip.PublicIpAddress, fmt.Sprintf("%d", publicip.IpVersion)},
			}
			keys, values := getTags(publicip.Tags)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
		}
	}
}

func buildBandwidthsInfo(sysConfigMap map[string][]string, filterMetrics *[]model.MetricInfoList, resourceInfos map[string]labelInfo) {
	bandwidths, err := getAllBandwidthFromRMS()
	if err != nil {
		return
	}
	for _, bandwidth := range bandwidths {
		if metricNames, ok := sysConfigMap["bandwidth_id"]; ok {
			metrics := buildSingleDimensionMetrics(metricNames, "SYS.VPC", "bandwidth_id", bandwidth.ID)
			*filterMetrics = append(*filterMetrics, metrics...)
			info := labelInfo{
				Name:  []string{"name", "epId", "bandwidthType", "type"},
				Value: []string{bandwidth.Name, bandwidth.EpId, bandwidth.BandwidthType, bandwidth.Type},
			}
			keys, values := getTags(bandwidth.Tags)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
		}
	}
}

func getAllBandwidthFromRMS() ([]Bandwidth, error) {
	resp, err := listResources("vpc", "bandwidths")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of vpc.bandwidths, error: %s", err.Error())
		return nil, err
	}
	bandwidths := make([]Bandwidth, 0, len(resp))
	for _, resource := range resp {
		var bandwidthProperties BandwidthProperties
		err := fmtResourceProperties(resource.Properties, &bandwidthProperties)
		if err != nil {
			logs.Logger.Errorf("fmt bandwidth properties error: %s", err.Error())
			continue
		}
		bandwidths = append(bandwidths, Bandwidth{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   *resource.Id,
				Name: *resource.Name,
				EpId: *resource.EpId,
				Tags: resource.Tags,
			},
			BandwidthProperties: bandwidthProperties,
		})
	}
	return bandwidths, nil
}

func getAllPublicIpFromRMS() ([]PublicIp, error) {
	resp, err := listResources("vpc", "publicips")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of vpc.publicips, error: %s", err.Error())
		return nil, err
	}
	publicips := make([]PublicIp, 0, len(resp))
	for _, resource := range resp {
		var publicIpProperties PublicIpProperties
		err := fmtResourceProperties(resource.Properties, &publicIpProperties)
		if err != nil {
			logs.Logger.Errorf("fmt publicIp properties error: %s", err.Error())
			continue
		}
		publicips = append(publicips, PublicIp{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   *resource.Id,
				Name: *resource.Name,
				EpId: *resource.EpId,
				Tags: resource.Tags},
			PublicIpProperties: publicIpProperties,
		})
	}
	return publicips, nil
}
