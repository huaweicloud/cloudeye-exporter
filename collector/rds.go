package collector

import (
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

type RdsInstanceInfo struct {
	ResourceBaseInfo
	RdsInstanceProperties
}
type RdsInstanceProperties struct {
	EngineVersion string `json:"engineVersion"`
	NodeNum       string `json:"nodeNum"`
	Port          string `json:"port"`
	DataVip       string `json:"dataVip"`
	EngineName    string `json:"engineName"`
}

var rdsInfo serversInfo

type RDSInfo struct{}

func (getter RDSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	rdsInfo.Lock()
	defer rdsInfo.Unlock()
	if rdsInfo.LabelInfo == nil || time.Now().Unix() > rdsInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.RDS")
		rdsInstances, err := getAllRdsInstanceSFromRMS()
		if err == nil {
			for _, instance := range rdsInstances {
				var dimName string
				switch instance.EngineName {
				case "mysql":
					dimName = "rds_cluster_id"
				case "postgresql":
					dimName = "postgresql_cluster_id"
				case "sqlserver":
					dimName = "rds_cluster_sqlserver_id"
				}
				if metricNames, ok := sysConfigMap[dimName]; ok {
					metrics := buildSingleDimensionMetrics(metricNames, "SYS.RDS", dimName, instance.ID)
					filterMetrics = append(filterMetrics, metrics...)
					info := labelInfo{
						Name:  []string{"name", "epId", "engineVersion", "nodeNum", "port", "dataVip", "engineName"},
						Value: []string{instance.Name, instance.EpId, instance.EngineVersion, instance.NodeNum, instance.Port, instance.DataVip, instance.EngineName},
					}
					keys, values := getTags(instance.Tags)
					info.Name = append(info.Name, keys...)
					info.Value = append(info.Value, values...)
					resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
				}
			}
		}

		rdsInfo.LabelInfo = resourceInfos
		rdsInfo.FilterMetrics = filterMetrics
		rdsInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return rdsInfo.LabelInfo, rdsInfo.FilterMetrics
}

func getAllRdsInstanceSFromRMS() ([]RdsInstanceInfo, error) {
	resp, err := listResources("rds", "instances")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of rds.instances, error: %s", err.Error())
		return nil, err
	}
	rdsInstances := make([]RdsInstanceInfo, 0, len(resp))
	for _, resource := range resp {
		var rdsInstanceProperties RdsInstanceProperties
		err := fmtResourceProperties(resource.Properties, &rdsInstanceProperties)
		if err != nil {
			logs.Logger.Errorf("fmt rds properties error: %s", err.Error())
			continue
		}
		rdsInstances = append(rdsInstances, RdsInstanceInfo{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   *resource.Id,
				Name: *resource.Name,
				EpId: *resource.EpId,
				Tags: resource.Tags,
			},
			RdsInstanceProperties: rdsInstanceProperties,
		})
	}
	return rdsInstances, nil
}
