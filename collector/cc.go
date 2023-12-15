package collector

import (
	"fmt"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	cc "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cc/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cc/v3/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cc/v3/region"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var (
	ccInfo serversInfo
	limit  = int32(2000)
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
		ccInfo.LabelInfo, ccInfo.FilterMetrics = buildResourceInfoAndMetrics(metricNames, connections, packages, bandwidths)
		ccInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return ccInfo.LabelInfo, ccInfo.FilterMetrics
}

func buildResourceInfoAndMetrics(metricNames []string, connections map[string]model.CloudConnection, packages map[string]model.BandwidthPackage, bandwidths []model.InterRegionBandwidth) (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]cesmodel.MetricInfoList, 0)
	for _, bandwidth := range bandwidths {
		if *bandwidth.CloudConnectionId == "" || *bandwidth.BandwidthPackageId == "" {
			continue
		}
		metrics := buildDimensionMetrics(metricNames, CCNamespace,
			[]cesmodel.MetricsDimension{{Name: "cloud_connect_id", Value: *bandwidth.CloudConnectionId},
				{Name: "bwp_id", Value: *bandwidth.BandwidthPackageId},
				{Name: "region_bandwidth_id", Value: *bandwidth.Id}})
		filterMetrics = append(filterMetrics, metrics...)

		var info labelInfo
		connectionName, connectionValue := getConnectionInfo(connections, *bandwidth.CloudConnectionId)
		info.Name = append(info.Name, connectionName...)
		info.Value = append(info.Value, connectionValue...)

		pkgName, pkgValue := getBandwidthPackageInfo(packages, *bandwidth.BandwidthPackageId)
		info.Name = append(info.Name, pkgName...)
		info.Value = append(info.Value, pkgValue...)

		if bandwidth.InterRegions != nil && len(*bandwidth.InterRegions) != 0 {
			info.Name = append(info.Name, "interRegions", "bandwidthName")
			info.Value = append(info.Value, fmt.Sprintf("%s<->%s",
				getDefaultString((*bandwidth.InterRegions)[0].LocalRegionId), getDefaultString((*bandwidth.InterRegions)[0].RemoteRegionId)),
				getDefaultString(bandwidth.Name))
		}
		resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
	}
	return resourceInfos, filterMetrics
}

func getConnectionInfo(connections map[string]model.CloudConnection, connectionId string) ([]string, []string) {
	connection, ok := connections[connectionId]
	if ok {
		return []string{"connectionName", "connectionEpId"}, []string{*connection.Name, *connection.EnterpriseProjectId}
	}
	return nil, nil
}

func getBandwidthPackageInfo(packages map[string]model.BandwidthPackage, connectionId string) ([]string, []string) {
	pkg, ok := packages[connectionId]
	if !ok {
		return nil, nil
	}
	name := []string{"packageName", "packageEpId"}
	vale := []string{getDefaultString(pkg.Name), getDefaultString(pkg.EnterpriseProjectId)}
	if pkg.Tags != nil {
		keys, values := getTags(fmtTags(pkg.Tags))
		name = append(name, keys...)
		vale = append(vale, values...)
	}
	return name, vale
}

func listCCConnections() (map[string]model.CloudConnection, error) {
	request := &model.ListCloudConnectionsRequest{Limit: &limit}
	client := getCCClient()
	connections := make(map[string]model.CloudConnection, 0)
	for {
		response, err := client.ListCloudConnections(request)
		if err != nil {
			return connections, err
		}
		for _, connection := range *response.CloudConnections {
			connections[*connection.Id] = connection
		}
		if response.PageInfo.NextMarker == nil {
			break
		}
		request.Marker = response.PageInfo.NextMarker
	}
	return connections, nil
}

func listBandwidthPackages() (map[string]model.BandwidthPackage, error) {
	request := &model.ListBandwidthPackagesRequest{Limit: &limit}
	client := getCCClient()
	bandwidthPackages := make(map[string]model.BandwidthPackage, 0)
	for {
		response, err := client.ListBandwidthPackages(request)
		if err != nil {
			logs.Logger.Errorf("Failed to list BandwidthPackages, error: %s", err.Error())
			return bandwidthPackages, err
		}
		for _, bandwidthPackage := range *response.BandwidthPackages {
			bandwidthPackages[*bandwidthPackage.Id] = bandwidthPackage
		}
		if response.PageInfo.NextMarker == nil {
			break
		}
		request.Marker = response.PageInfo.NextMarker
	}
	return bandwidthPackages, nil
}

func getCCClient() *cc.CcClient {
	return cc.NewCcClient(cc.CcClientBuilder().WithRegion(region.ValueOf("cn-north-4")).
		WithCredential(global.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).Build()).Build())
}

func listInterRegionBandwidths() ([]model.InterRegionBandwidth, error) {
	request := &model.ListInterRegionBandwidthsRequest{Limit: &limit}
	client := getCCClient()
	var resources []model.InterRegionBandwidth
	for {
		response, err := client.ListInterRegionBandwidths(request)
		if err != nil {
			logs.Logger.Errorf("Failed to list InterRegionBandwidths, error: %s", err.Error())
			return resources, err
		}
		resources = append(resources, *response.InterRegionBandwidths...)
		if response.PageInfo.NextMarker == nil {
			break
		}
		request.Marker = response.PageInfo.NextMarker
	}
	return resources, nil
}
