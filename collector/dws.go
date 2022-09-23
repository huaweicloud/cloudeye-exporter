package collector

import (
	"github.com/huaweicloud/cloudeye-exporter/logs"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	dws "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dws/v2"
	dwsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dws/v2/model"
)

var dwsInfo serversInfo

type DWSInfo struct{}

func (getter DWSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	dwsInfo.Lock()
	defer dwsInfo.Unlock()
	if dwsInfo.LabelInfo == nil || time.Now().Unix() > dwsInfo.TTL {
		resp, err := getDWSClient().ListClusters(&dwsmodel.ListClustersRequest{})
		if err != nil {
			logs.Logger.Errorf("List dws clusters error: %s", err.Error())
			return dwsInfo.LabelInfo, dwsInfo.FilterMetrics
		}
		for _, cluster := range *resp.Clusters {
			metrics := buildSingleDimensionMetrics(getMetricConfigMap("SYS.DWS")["datastore_id"], "SYS.DWS", "datastore_id", cluster.Id)
			filterMetrics = append(filterMetrics, metrics...)
			info := labelInfo{
				Name:  []string{"clusterName", "epId"},
				Value: []string{cluster.Name, cluster.EnterpriseProjectId},
			}
			keys, values := getTags(fmtTags(cluster.Tags))
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			for _, node := range cluster.Nodes {
				nodeMetrics := buildSingleDimensionMetrics(getMetricConfigMap("SYS.DWS")["dws_instance_id"], "SYS.DWS", "dws_instance_id", node.Id)
				filterMetrics = append(filterMetrics, nodeMetrics...)
				resourceInfos[GetResourceKeyFromMetricInfo(nodeMetrics[0])] = info
			}
		}
		dwsInfo.LabelInfo = resourceInfos
		dwsInfo.FilterMetrics = filterMetrics
		dwsInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return dwsInfo.LabelInfo, dwsInfo.FilterMetrics
}

func getDWSClient() *dws.DwsClient {
	return dws.NewDwsClient(dws.DwsClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(true)).
		WithEndpoint(getEndpoint("dws", "v1.0")).Build())
}
