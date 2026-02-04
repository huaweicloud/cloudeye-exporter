package collector

import (
	"net/http"
	"time"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	clouttable "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cloudtable/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cloudtable/v2/model"
)

var cloudTableInfo serversInfo

type CloudTableInfo struct{}

type ClusterDetailReq struct {
	ClusterId string `json:"cluster_id"`
	XLanguage string `json:"X-Language"`
}
type ClusterDetailInfoResp struct {
	HttpStatusCode        int                   `json:"-"`
	DorisInformation      DorisInformation      `json:"doris_information,omitempty"`
	HbaseInformation      HbaseInformation      `json:"hbase_information,omitempty"`
	ClickhouseInformation ClickhouseInformation `json:"clickhouse_information,omitempty"`
	StarrocksInformation  StarrocksInformation  `json:"starrocks_information,omitempty"`
}
type DorisInformation struct {
	NodeNameIPMap map[string]string `json:"dimension_inst_name_ip_map,omitempty"`
}
type HbaseInformation struct {
	NodeNameIPMap map[string]string `json:"dimension_inst_name_ip_map,omitempty"`
}
type ClickhouseInformation struct {
	NodeNameIPMap map[string]string `json:"dimension_inst_name_ip_map,omitempty"`
}
type StarrocksInformation struct {
	NodeNameIPMap map[string]string `json:"dimension_inst_name_ip_map,omitempty"`
}

func getCloudTableClient() *clouttable.CloudTableClient {
	return clouttable.NewCloudTableClient(clouttable.CloudTableClientBuilder().WithCredential(
		authCredentialMap[conf.AuthMode](RegionServiceType)).
		WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
		WithEndpoint(getEndpoint("cloudtable", "v2")).Build())
}

func (ct CloudTableInfo) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	cloudTableInfo.Lock()
	defer cloudTableInfo.Unlock()
	if cloudTableInfo.LabelInfo == nil || time.Now().Unix() > cloudTableInfo.TTL {
		clusters, err := GetClusterInfo()
		if err != nil {
			logs.Logger.Errorf("Get cloud table resource info from service error: %s", err.Error())
			return cloudTableInfo.LabelInfo, cloudTableInfo.FilterMetrics
		}
		clusterInstanceMap := make(map[string]map[string]string)
		for _, cluster := range clusters {
			info := labelInfo{
				Name:  []string{"clusterName"},
				Value: []string{*cluster.ClusterName},
			}
			resourceInfos[*cluster.ClusterId] = info
			instanceMap := getInstanceByClusterId(*cluster.ClusterId)
			if instanceMap != nil && len(instanceMap) != 0 {
				clusterInstanceMap[*cluster.ClusterId] = instanceMap
			}
		}
		allMetrics, err := listAllMetrics("SYS.CloudTable")
		if err != nil {
			logs.Logger.Errorf("[%s] Get all metrics of SYS.CloudTable error: %s", err.Error())
			return cloudTableInfo.LabelInfo, cloudTableInfo.FilterMetrics
		}

		var filteredMetrics []cesmodel.MetricInfoList
		for _, metricInfo := range allMetrics {
			if IsMetricInfoInWhiteList(metricInfo) {
				filteredMetrics = append(filteredMetrics, metricInfo)
			}
		}

		for _, metric := range filteredMetrics {
			dimensions := metric.Dimensions
			clusterId := GetDimensionsValueByName(dimensions, "cluster_id")
			metricKey := GetResourceKeyFromMetricInfo(metric)
			if clusterInfo, ok := resourceInfos[clusterId]; ok {
				if ContainDimensionName(dimensions, "cluster_id") && ContainDimensionName(dimensions, "instance_name") {
					instanceName := GetDimensionsValueByName(dimensions, "instance_name")
					instanceInfoMap := clusterInstanceMap[clusterId]
					ip := ""
					if instanceInfoMap != nil && len(instanceInfoMap) != 0 {
						ip = instanceInfoMap[instanceName]
					}
					instanceInfo := labelInfo{
						Name:  []string{"clusterName", "nodeIp"},
						Value: []string{clusterInfo.Value[0], ip},
					}
					resourceInfos[metricKey] = instanceInfo
				} else {
					resourceInfos[metricKey] = clusterInfo
				}
			}
		}
		cloudTableInfo.LabelInfo = resourceInfos
		cloudTableInfo.FilterMetrics = filteredMetrics
		cloudTableInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return cloudTableInfo.LabelInfo, cloudTableInfo.FilterMetrics
}

func getInstanceByClusterId(clusterId string) map[string]string {
	request := &ClusterDetailReq{ClusterId: clusterId}
	resp, err := getHcClient(getEndpoint("cloudtable", "v2")).Sync(request, genReqDefForClusterDetail())
	if err != nil {
		logs.Logger.Errorf("Failed to get cluster detail: %s", err.Error())
		return nil
	}
	response, ok := resp.(*ClusterDetailInfoResp)
	if !ok {
		logs.Logger.Error("Failed to get ClusterDetailInfo: resp type is not ClusterDetailInfoResp")
		return nil
	}
	// DorisInformation, HbaseInformation, ClickhouseInformation, StarrocksInformation 这4个属性只会存在1个
	if &response.StarrocksInformation != nil && len((&response.StarrocksInformation).NodeNameIPMap) != 0 {
		return (&response.StarrocksInformation).NodeNameIPMap
	}
	if &response.HbaseInformation != nil && len((&response.HbaseInformation).NodeNameIPMap) != 0 {
		return (&response.HbaseInformation).NodeNameIPMap
	}
	if &response.ClickhouseInformation != nil && len((&response.ClickhouseInformation).NodeNameIPMap) != 0 {
		return (&response.ClickhouseInformation).NodeNameIPMap
	}
	if &response.DorisInformation != nil && len((&response.DorisInformation).NodeNameIPMap) != 0 {
		return (&response.DorisInformation).NodeNameIPMap
	}
	return nil
}

func genReqDefForClusterDetail() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().
		WithMethod(http.MethodGet).
		WithPath("/v2/{project_id}/clusters/{cluster_id}").
		WithResponse(new(ClusterDetailInfoResp)).
		WithContentType("application/json")

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("ClusterId").
		WithJsonTag("cluster_id").
		WithLocationType(def.Path))

	reqDefBuilder.WithRequestField(def.NewFieldDef().
		WithName("XLanguage").
		WithJsonTag("X-Language").
		WithLocationType(def.Header))

	requestDef := reqDefBuilder.Build()
	return requestDef
}

func GetClusterInfo() ([]model.ClusterDetail, error) {
	cloudTableClusterLimit := int32(100)
	cloudTableClusterOffset := int32(0)
	request := &model.ListClustersRequest{Limit: &cloudTableClusterLimit, Offset: &cloudTableClusterOffset}
	var clusters []model.ClusterDetail
	for {
		response, err := getCloudTableClient().ListClusters(request)
		if err != nil {
			logs.Logger.Errorf("list cloud table clusters error: %s, limit: %d, offset: %d", err.Error(),
				*request.Limit, *request.Offset)
			return nil, err
		}
		tempClusters := *response.Clusters
		if len(tempClusters) == 0 {
			break
		}
		clusters = append(clusters, tempClusters...)
		*request.Offset += cloudTableClusterLimit
	}
	return clusters, nil
}
