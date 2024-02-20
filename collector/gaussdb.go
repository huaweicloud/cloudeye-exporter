package collector

import (
	"time"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	rmsModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1/model"
)

type GaussdbNodeInfo struct {
	ResourceBaseInfo
	NodeProperties
}

type NodeProperties struct {
	InstanceName string                   `json:"instance_name"`
	InstanceId   string                   `json:"instanceId"`
	InnerPort    string                   `json:"innerPort"`
	InnerIp      string                   `json:"innerIp"`
	Role         string                   `json:"role"`
	EngineName   string                   `json:"engineName"`
	Dimensions   []model.MetricsDimension `json:"dimensions"`
}

var gaussdbInfo serversInfo

type GAUSSDBInfo struct{}

func (getter GAUSSDBInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	gaussdbInfo.Lock()
	defer gaussdbInfo.Unlock()
	if gaussdbInfo.LabelInfo == nil || time.Now().Unix() > gaussdbInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.GAUSSDB")
		if metricNames, ok := sysConfigMap["gaussdb_mysql_instance_id,gaussdb_mysql_node_id"]; ok {
			nodes, err := getAllGaussdbNodesFromRMS()
			if err == nil {
				for _, node := range nodes {
					metrics := buildDimensionMetrics(metricNames, "SYS.GAUSSDB", node.Dimensions)
					filterMetrics = append(filterMetrics, metrics...)
					info := labelInfo{
						Name:  []string{"instanceName", "name", "epId", "innerPort", "innerIp", "role", "engineName"},
						Value: []string{node.InstanceName, node.Name, node.EpId, node.InnerPort, node.InnerIp, node.Role, node.EngineName},
					}
					keys, values := getTags(node.Tags)
					info.Name = append(info.Name, keys...)
					info.Value = append(info.Value, values...)
					resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
				}
			}
		}

		gaussdbInfo.LabelInfo = resourceInfos
		gaussdbInfo.FilterMetrics = filterMetrics
		gaussdbInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return gaussdbInfo.LabelInfo, gaussdbInfo.FilterMetrics
}

func getAllGaussdbNodesFromRMS() ([]GaussdbNodeInfo, error) {
	resp, err := listResources("gaussdb", "nodes")
	instanceResp, err := listResources("gaussdb", "instance")

	instanceMap := make(map[string]rmsModel.ResourceEntity)
	for _, entity := range instanceResp {
		instanceMap[*entity.Id] = entity
	}
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of gaussdb.nodes, error: %s", err.Error())
		return nil, err
	}
	nodes := make([]GaussdbNodeInfo, 0, len(resp))
	for _, resource := range resp {
		var properties NodeProperties
		err := fmtResourceProperties(resource.Properties, &properties)
		if err != nil {
			logs.Logger.Errorf("fmt gaussdb node properties error: %s", err.Error())
			continue
		}
		instanceResource, ok := instanceMap[properties.InstanceId]
		if ok {
			properties.InstanceName = *instanceResource.Name
		} else {
			logs.Logger.Errorf("Get gaussdb instance name empty, instance id is %s", properties.InstanceId)
		}
		nodes = append(nodes, GaussdbNodeInfo{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   *resource.Id,
				Name: *resource.Name,
				EpId: *resource.EpId,
				Tags: resource.Tags,
			},
			NodeProperties: properties,
		})

	}
	return nodes, nil
}
