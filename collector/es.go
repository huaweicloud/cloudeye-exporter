package collector

import (
	"encoding/json"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

type EsInstanceInfo struct {
	ResourceBaseInfo
	Properties EsInstanceProperties
}
type EsInstanceProperties struct {
	ID          string     `json:"id"`
	ClusterMode string     `json:"clusterMode"`
	Endpoint    string     `json:"endpoint"`
	Instances   []Instance `json:"instances"`
}

type Instance struct {
	Role      string `json:"role"`
	SpecCode  string `json:"specCode"`
	PrivateIP string `json:"privateIp"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	ID        string `json:"id"`
	AzCode    string `json:"azCode"`
	IsFrozen  string `json:"isFrozen"`
	Group     string `json:"group"`
	Status    string `json:"status"`
}

var esInfo serversInfo

func (exporter *BaseHuaweiCloudExporter) getEsResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	esInfo.Lock()
	defer esInfo.Unlock()
	if esInfo.LabelInfo == nil || time.Now().Unix() > esInfo.TTL {
		buildEsResourceInfo(&filterMetrics, resourceInfos)
		esInfo.LabelInfo = resourceInfos
		esInfo.FilterMetrics = filterMetrics
		esInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return esInfo.LabelInfo, esInfo.FilterMetrics
}

func buildEsResourceInfo(filterMetrics *[]model.MetricInfoList, resourceInfos map[string]labelInfo) {
	sysConfigMap := getMetricConfigMap("SYS.ES")
	esInstances, err := getAllEsInstanceSFromRMS()
	if err != nil {
		return
	}
	for _, esInstance := range esInstances {
		if metricNames, ok := sysConfigMap["cluster_id"]; ok {
			metrics := buildSingleDimensionMetrics(metricNames, "SYS.ES", "cluster_id", esInstance.ID)
			*filterMetrics = append(*filterMetrics, metrics...)
			info := labelInfo{
				Name:  []string{"name", "epId", "clusterMode", "endpoint"},
				Value: []string{esInstance.Name, esInstance.EpId, esInstance.Properties.ClusterMode, esInstance.Properties.Endpoint},
			}
			keys, values := getTags(esInstance.Tags)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info

			if instanceMetricNames, ok := sysConfigMap["cluster_id,instance_id"]; ok {
				for _, instance := range esInstance.Properties.Instances {
					instanceInfo := labelInfo{
						Name:  []string{"instanceName", "type", "privateIp", "role"},
						Value: []string{instance.Name, instance.Type, instance.PrivateIP, instance.Role},
					}
					instanceMetrics := buildDimensionMetrics(instanceMetricNames, "SYS.ES",
						[]model.MetricsDimension{{Name: "cluster_id", Value: esInstance.ID}, {Name: "instance_id", Value: instance.ID}})
					*filterMetrics = append(*filterMetrics, instanceMetrics...)
					instanceInfo.Name = append(instanceInfo.Name, info.Name...)
					instanceInfo.Value = append(instanceInfo.Value, info.Value...)
					resourceInfos[GetResourceKeyFromMetricInfo(instanceMetrics[0])] = instanceInfo
				}
			}
		}
	}
}

func getAllEsInstanceSFromRMS() ([]EsInstanceInfo, error) {
	resp, err := listResources("css", "clusters")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of css.clusters, error: %s", err.Error())
		return nil, err
	}
	esInstances := make([]EsInstanceInfo, 0, len(resp))
	for _, resource := range resp {
		bandwidthProperties, err := fmtEsInstanceProperties(resource.Properties)
		if err != nil {
			continue
		}
		esInstances = append(esInstances, EsInstanceInfo{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   *resource.Id,
				Name: *resource.Name,
				EpId: *resource.EpId,
				Tags: resource.Tags,
			},
			Properties: *bandwidthProperties,
		})
	}
	return esInstances, nil
}

func fmtEsInstanceProperties(properties map[string]interface{}) (*EsInstanceProperties, error) {
	bytes, err := json.Marshal(properties)
	if err != nil {
		logs.Logger.Errorf("Marshal es instance properties error: %s", err.Error())
		return nil, err
	}
	var esInstanceProperties EsInstanceProperties
	err = json.Unmarshal(bytes, &esInstanceProperties)
	if err != nil {
		logs.Logger.Errorf("Unmarshal to EsInstanceProperties error: %s", err.Error())
		return nil, err
	}

	return &esInstanceProperties, nil
}
