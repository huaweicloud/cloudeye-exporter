package collector

import (
	"time"

	cdm "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdm/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdm/v1/model"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var cdmInfo serversInfo

type CDMInfo struct{}

func (getter CDMInfo) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]cesmodel.MetricInfoList, 0)
	cdmInfo.Lock()
	defer cdmInfo.Unlock()
	if cdmInfo.LabelInfo == nil || time.Now().Unix() > cdmInfo.TTL {
		if metricNames, ok := getMetricConfigMap("SYS.CDM")["instance_id"]; ok {
			clusters, err := listClusters()
			if err != nil {
				logs.Logger.Errorf("Get cdm resource info from service error: %s", err.Error())
				return cdmInfo.LabelInfo, cdmInfo.FilterMetrics
			}
			for _, cluster := range *clusters {
				for _, instance := range *cluster.Instances {
					metrics := buildSingleDimensionMetrics(metricNames, "SYS.CDM", "instance_id", instance.Id)
					filterMetrics = append(filterMetrics, metrics...)
					info := labelInfo{
						Name:  []string{"clusterId", "clusterName", "instanceName", "trafficIp", "manage_ip", "private_ip"},
						Value: []string{cluster.Id, cluster.Name, instance.Name, getDefaultString(instance.TrafficIp), getDefaultString(instance.ManageIp), getDefaultString(instance.PrivateIp)},
					}
					resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
				}
			}
		}
		cdmInfo.LabelInfo = resourceInfos
		cdmInfo.FilterMetrics = filterMetrics
		cdmInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return cdmInfo.LabelInfo, cdmInfo.FilterMetrics
}

func getCDMClient() *cdm.CdmClient {
	return cdm.NewCdmClient(cdm.CdmClientBuilder().WithCredential(
		authCredentialMap[conf.AuthMode](RegionServiceType)).
		WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
		WithEndpoint(getEndpoint("cdm", "v1.1")).Build())
}

func listClusters() (*[]model.Clusters, error) {
	response, err := getCDMClient().ListClusters(&model.ListClustersRequest{})
	if err != nil {
		logs.Logger.Errorf("list cdm clusters error: %s", err.Error())
		return nil, err
	}
	return response.Clusters, nil
}
