package collector

import (
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

type GESInstanceInfo struct {
	ResourceBaseInfo
	GESInstanceProperties
}
type GESInstanceProperties struct {
	PrivateIp string `json:"privateIp"`
}

var gesInfo serversInfo

type GESInfo struct{}

func (getter GESInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	gesInfo.Lock()
	defer gesInfo.Unlock()
	if gesInfo.LabelInfo == nil || time.Now().Unix() > gesInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.GES")
		gesInstances, err := getAllGesInstancesFromRMS()
		if err != nil {
			logs.Logger.Errorf("Get all ges instances error: %s", err.Error())
			return gesInfo.LabelInfo, gesInfo.FilterMetrics
		}
		metricNames := sysConfigMap["instance_id"]
		if len(metricNames) == 0 {
			logs.Logger.Warn("Metric config is empty of SYS.GES, dim_metric_name is instance_id.")
			return gesInfo.LabelInfo, gesInfo.FilterMetrics
		}
		for _, instance := range gesInstances {
			metrics := buildSingleDimensionMetrics(metricNames, "SYS.GES", "instance_id", instance.ID)
			filterMetrics = append(filterMetrics, metrics...)
			info := labelInfo{
				Name:  []string{"name", "epId", "privateIp"},
				Value: []string{instance.Name, instance.EpId, instance.PrivateIp},
			}
			keys, values := getTags(instance.Tags)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
		}

		gesInfo.LabelInfo = resourceInfos
		gesInfo.FilterMetrics = filterMetrics
		gesInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return gesInfo.LabelInfo, gesInfo.FilterMetrics
}

func getAllGesInstancesFromRMS() ([]GESInstanceInfo, error) {
	resp, err := listResources("ges", "graphs")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of ges.graphs, error: %s", err.Error())
		return nil, err
	}
	gesInstances := make([]GESInstanceInfo, 0, len(resp))
	for _, resource := range resp {
		var gesInstanceProperties GESInstanceProperties
		err := fmtResourceProperties(resource.Properties, &gesInstanceProperties)
		if err != nil {
			logs.Logger.Errorf("Fmt ges properties error: %s", err.Error())
			continue
		}
		gesInstances = append(gesInstances, GESInstanceInfo{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   *resource.Id,
				Name: *resource.Name,
				EpId: *resource.EpId,
				Tags: resource.Tags,
			},
			GESInstanceProperties: gesInstanceProperties,
		})
	}
	return gesInstances, nil
}
