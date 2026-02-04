package collector

import (
	"fmt"
	"time"

	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	cc "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cc/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cc/v3/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cc/v3/region"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var (
	ccInfo serversInfo
	limit  = int32(500)
)

const (
	CCNamespace      = "SYS.CC"
	CCConfigDimNames = "cloud_connect_id,bwp_id,region_bandwidth_id"
)

type CCInfo struct{}

func (getter CCInfo) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	ccInfo.Lock()
	defer ccInfo.Unlock()
	if ccInfo.LabelInfo == nil || time.Now().Unix() > ccInfo.TTL {
		sysConfigMap := getMetricConfigMap(CCNamespace)
		metricNames := sysConfigMap[CCConfigDimNames]
		if len(metricNames) == 0 {
			logs.Logger.Warn("Metric config is empty of SYS.CC.")
			return ccInfo.LabelInfo, ccInfo.FilterMetrics
		}

		connections, err := listCCConnections()
		if err != nil {
			logs.Logger.Errorf("Get all connections error: %s", err.Error())
			return ccInfo.LabelInfo, ccInfo.FilterMetrics
		}

		packages, err := listBandwidthPackages()
		if err != nil {
			logs.Logger.Errorf("Get all bandwidth packages error: %s", err.Error())
			return ccInfo.LabelInfo, ccInfo.FilterMetrics
		}

		bandwidths, err := listInterRegionBandwidths()
		if err != nil {
			logs.Logger.Errorf("Get all inter region bandwidths error: %s", err.Error())
			return ccInfo.LabelInfo, ccInfo.FilterMetrics
		}
		resourceInfos, filterMetrics := buildResourceInfoAndMetrics(metricNames, connections, packages, bandwidths)
		ccInfo.LabelInfo = resourceInfos
		ccInfo.FilterMetrics = filterMetrics
		ccInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return ccInfo.LabelInfo, ccInfo.FilterMetrics
}

func buildResourceInfoAndMetrics(metricNames []string, connections map[string]model.CloudConnection, packages map[string]model.BandwidthPackage, bandwidths []model.InterRegionBandwidth) (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]cesmodel.MetricInfoList, 0)
	for _, bandwidth := range bandwidths {
		if bandwidth.CloudConnectionId == "" || bandwidth.BandwidthPackageId == "" {
			continue
		}
		var info labelInfo
		connectionName, connectionValue := getConnectionInfo(connections, bandwidth.CloudConnectionId)
		info = appendNameValuePairToLabelInfo(info, connectionName, connectionValue)

		pkgName, pkgValue := getBandwidthPackageInfo(packages, bandwidth.BandwidthPackageId)
		info = appendNameValuePairToLabelInfo(info, pkgName, pkgValue)

		var regionBandWidthIds []string
		if bandwidth.InterRegions != nil && len(*bandwidth.InterRegions) != 0 {
			info.Name = append(info.Name, "interRegions", "bandwidthName")
			info.Value = append(info.Value, getInterRegionsInfo(bandwidth), getDefaultString(&bandwidth.Name))
			for _, interRegion := range *bandwidth.InterRegions {
				regionBandWidthIds = append(regionBandWidthIds, interRegion.Id)
			}
		}

		for _, regionBandWidthId := range regionBandWidthIds {
			metrics := buildDimensionMetrics(metricNames, CCNamespace,
				[]cesmodel.MetricsDimension{{Name: "cloud_connect_id", Value: bandwidth.CloudConnectionId},
					{Name: "bwp_id", Value: bandwidth.BandwidthPackageId},
					{Name: "region_bandwidth_id", Value: regionBandWidthId}})
			filterMetrics = append(filterMetrics, metrics...)
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
		}
	}
	return resourceInfos, filterMetrics
}

func getInterRegionsInfo(bandwidth model.InterRegionBandwidth) string {
	if bandwidth.InterRegions == nil || len(*bandwidth.InterRegions) == 0 {
		return ""
	}
	var localRegionId string
	var remoteRegionId string
	for _, interRegion := range *bandwidth.InterRegions {
		if *interRegion.LocalRegionId == conf.Region {
			localRegionId = getDefaultString(interRegion.LocalRegionId)
			remoteRegionId = getDefaultString(interRegion.RemoteRegionId)
			break
		}
	}
	if localRegionId == "" && remoteRegionId == "" {
		localRegionId = getDefaultString((*bandwidth.InterRegions)[0].LocalRegionId)
		remoteRegionId = getDefaultString((*bandwidth.InterRegions)[0].RemoteRegionId)
	}
	return fmt.Sprintf("%s->%s", localRegionId, remoteRegionId)
}

func getConnectionInfo(connections map[string]model.CloudConnection, connectionId string) ([]string, []string) {
	connection, ok := connections[connectionId]
	if ok {
		names := []string{"connectionName", "connectionEpId"}
		values := []string{connection.Name, *connection.EnterpriseProjectId}
		if connection.Tags != nil {
			tagKeys, tagValues := getTags(fmtTags(connection.Tags))
			names = append(names, tagKeys...)
			values = append(values, tagValues...)
		}
		return names, values
	}
	return nil, nil
}

func getBandwidthPackageInfo(packages map[string]model.BandwidthPackage, connectionId string) ([]string, []string) {
	pkg, ok := packages[connectionId]
	if !ok {
		return nil, nil
	}
	name := []string{"packageName", "packageEpId"}
	vale := []string{getDefaultString(&pkg.Name), getDefaultString(pkg.EnterpriseProjectId)}
	if pkg.Tags != nil {
		keys, values := getTags(fmtTags(pkg.Tags))
		name = append(name, keys...)
		vale = append(vale, values...)
	}
	return name, vale
}

func listCCConnections() (map[string]model.CloudConnection, error) {
	epIds := getEpIdRequestPart()
	request := &model.ListCloudConnectionsRequest{Limit: &limit, EnterpriseProjectId: &epIds}
	client := getCCClient()
	connections := make(map[string]model.CloudConnection, 0)
	for {
		response, err := client.ListCloudConnections(request)
		if err != nil {
			return connections, err
		}
		for _, connection := range response.CloudConnections {
			connections[connection.Id] = connection
		}
		if response.PageInfo.NextMarker == nil {
			break
		}
		request.Marker = response.PageInfo.NextMarker
	}
	return connections, nil
}

func listBandwidthPackages() (map[string]model.BandwidthPackage, error) {
	epIds := getEpIdRequestPart()
	request := &model.ListBandwidthPackagesRequest{Limit: &limit, EnterpriseProjectId: &epIds}
	client := getCCClient()
	bandwidthPackages := make(map[string]model.BandwidthPackage, 0)
	for {
		response, err := client.ListBandwidthPackages(request)
		if err != nil {
			logs.Logger.Errorf("Failed to list BandwidthPackages, error: %s", err.Error())
			return bandwidthPackages, err
		}
		for _, bandwidthPackage := range response.BandwidthPackages {
			bandwidthPackages[bandwidthPackage.Id] = bandwidthPackage
		}
		if response.PageInfo.NextMarker == nil {
			break
		}
		request.Marker = response.PageInfo.NextMarker
	}
	return bandwidthPackages, nil
}

func getCCClient() *cc.CcClient {
	return cc.NewCcClient(getCCClientBuilder().Build())
}

func getCCClientBuilder() *http_client.HcHttpClientBuilder {
	builder := cc.CcClientBuilder().WithCredential(authCredentialMap[conf.AuthMode](GlobalServiceType)).
		WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify))
	if endpoint, ok := endpointConfig["cc"]; ok {
		builder.WithEndpoint(endpoint)
	} else {
		builder.WithRegion(region.ValueOf("cn-north-4"))
	}
	return builder
}

func listInterRegionBandwidths() ([]model.InterRegionBandwidth, error) {
	epIds := getEpIdRequestPart()
	request := &model.ListInterRegionBandwidthsRequest{Limit: &limit, EnterpriseProjectId: &epIds}
	client := getCCClient()
	var resources []model.InterRegionBandwidth
	for {
		response, err := client.ListInterRegionBandwidths(request)
		if err != nil {
			logs.Logger.Errorf("Failed to list InterRegionBandwidths, error: %s", err.Error())
			return resources, err
		}
		resources = append(resources, response.InterRegionBandwidths...)
		if response.PageInfo.NextMarker == nil {
			break
		}
		request.Marker = response.PageInfo.NextMarker
	}
	return resources, nil
}
