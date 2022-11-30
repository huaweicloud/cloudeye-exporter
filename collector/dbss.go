package collector

import (
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var dbssInfo serversInfo

type DBSSInfo struct{}

func (getter DBSSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	dbssInfo.Lock()
	defer dbssInfo.Unlock()
	if dbssInfo.LabelInfo == nil || time.Now().Unix() > dbssInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.DBSS")
		servers, err := getAllDBSSServersFromRMS()
		if err != nil {
			logs.Logger.Errorf("Get all dbss servers error: %s", err.Error())
			return dbssInfo.LabelInfo, dbssInfo.FilterMetrics
		}
		metricNames := sysConfigMap["audit_id"]
		if len(metricNames) == 0 {
			logs.Logger.Warn("Metric config is empty of SYS.DBSS, dim_metric_name is audit_id.")
			return dbssInfo.LabelInfo, dbssInfo.FilterMetrics
		}
		for _, server := range servers {
			metrics := buildSingleDimensionMetrics(metricNames, "SYS.DBSS", "audit_id", server.ID)
			filterMetrics = append(filterMetrics, metrics...)
			info := labelInfo{
				Name:  []string{"name", "epId"},
				Value: []string{server.Name, server.EpId},
			}
			keys, values := getTags(server.Tags)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
		}

		dbssInfo.LabelInfo = resourceInfos
		dbssInfo.FilterMetrics = filterMetrics
		dbssInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return dbssInfo.LabelInfo, dbssInfo.FilterMetrics
}

func getAllDBSSServersFromRMS() ([]ResourceBaseInfo, error) {
	resp, err := listResources("dbss", "cloudservers")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of dbss.cloudservers, error: %s", err.Error())
		return nil, err
	}
	servers := make([]ResourceBaseInfo, 0, len(resp))
	for _, resource := range resp {
		var properties struct {
			ID string `json:"id"`
		}
		err := fmtResourceProperties(resource.Properties, &properties)
		if err != nil {
			logs.Logger.Errorf("Fmt server properties error: %s", err.Error())
			continue
		}
		servers = append(servers, ResourceBaseInfo{
			ID:   properties.ID,
			Name: *resource.Name,
			EpId: *resource.EpId,
			Tags: resource.Tags,
		})
	}
	return servers, nil
}
