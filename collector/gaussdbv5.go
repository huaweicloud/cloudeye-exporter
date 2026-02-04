package collector

import (
	"fmt"
	"net/http"
	"time"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

type GaussdbV5InstancesRequest struct {
	Body GaussdbV5InstancesRequestBody `json:"body,omitempty"`
}

type GaussdbV5InstancesRequestBody struct {
	Namespace string                           `json:"namespace,omitempty"`
	Start     int32                            `json:"start,omitempty"`
	Limit     int32                            `json:"limit,omitempty"`
	Query     []GaussdbV5InstanceSubDimReqInfo `json:"query,omitempty"`
}

type GaussdbV5InstanceSubDimReqInfo struct {
	DimName string `json:"dim_name,omitempty"`
}

type GaussdbV5InstancesResponse struct {
	Instances      []GaussdbV5Instances `json:"instances,omitempty"`
	Total          int32                `json:"total,omitempty"`
	HttpStatusCode int                  `json:"-"`
}

type GaussdbV5Instances struct {
	InstanceInfo GaussdbV5Instance     `json:"gaussdbv5_instance_id"`
	Pdb          GaussdbV5Pdb          `json:"gaussdbv5_pdb_id"`
	Component    GaussdbV5PdbComponent `json:"gaussdbv5_pdb_component_id"`
	Node         GaussdbV5Node         `json:"gaussdbv5_node_id"`
}

type GaussdbV5Instance struct {
	InstanceId          string         `json:"id,omitempty"`
	InstanceName        string         `json:"name,omitempty"`
	Status              string         `json:"status,omitempty"`
	Tags                []GaussdbV5Tag `json:"tags,omitempty"`
	EnterpriseProjectId string         `json:"enterprise_project_id,omitempty"`
}

type GaussdbV5Tag struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type GaussdbV5Pdb struct {
	PdbId   string `json:"id,omitempty"`
	PdbName string `json:"name,omitempty"`
}

type GaussdbV5PdbComponent struct {
	ComponentId            string `json:"id,omitempty"`
	ComponentType          string `json:"pdb_component_type,omitempty"`
	ComponentDistributedId string `json:"pdb_component_distributed_id,omitempty"`
	ComponentDetail        string `json:"pdb_component_detail,omitempty"`
}

type GaussdbV5Node struct {
	ID               string `json:"id,omitempty"`
	Name             string `json:"name,omitempty"`
	Role             string `json:"node_role,omitempty"`
	Status           string `json:"status,omitempty"`
	AvailabilityZone string `json:"availability_zone,omitempty"`
}

var gaussdbV5Info serversInfo

type GAUSSDBV5Info struct{}

func (getter GAUSSDBV5Info) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]cesmodel.MetricInfoList, 0)
	gaussdbV5Info.Lock()
	defer gaussdbV5Info.Unlock()
	if gaussdbV5Info.LabelInfo == nil || time.Now().Unix() > gaussdbV5Info.TTL {
		sysConfigMap := getMetricConfigMap("SYS.GAUSSDBV5")
		if instanceMetricNames, ok := sysConfigMap["gaussdbv5_instance_id"]; ok {
			instances, err := listInstances("gaussdbv5_instance_id")
			if err != nil {
				logs.Logger.Errorf("Get gauss db v5 instanceId data from service error: %s", err.Error())
				return gaussdbV5Info.LabelInfo, gaussdbV5Info.FilterMetrics
			}

			for _, instance := range instances {
				metrics := buildSingleDimensionMetrics(instanceMetricNames, "SYS.GAUSSDBV5", "gaussdbv5_instance_id", instance.InstanceInfo.InstanceId)
				filterMetrics = append(filterMetrics, metrics...)
				info := GetInstanceLabelInfo(instance)
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			}
		}

		if nodeMetricNames, ok := sysConfigMap["gaussdbv5_instance_id,gaussdbv5_node_id"]; ok {
			instances, err := listInstances("gaussdbv5_node_id")
			if err != nil {
				logs.Logger.Errorf("Get gauss db v5 NodeId data from service error: %s", err.Error())
				return gaussdbV5Info.LabelInfo, gaussdbV5Info.FilterMetrics
			}
			for _, instance := range instances {
				node := instance.Node
				metrics := buildDimensionMetrics(nodeMetricNames, "SYS.GAUSSDBV5",
					[]cesmodel.MetricsDimension{{Name: "gaussdbv5_instance_id", Value: instance.InstanceInfo.InstanceId}, {Name: "gaussdbv5_node_id", Value: node.ID}})
				filterMetrics = append(filterMetrics, metrics...)
				info := GetInstanceLabelInfo(instance)
				nodeInfo := labelInfo{
					Name:  []string{"nodeName", "role", "status", "availability_zone"},
					Value: []string{node.Name, node.Role, node.Status, node.AvailabilityZone},
				}
				nodeInfo.Name = append(nodeInfo.Name, info.Name...)
				nodeInfo.Value = append(nodeInfo.Value, info.Value...)
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = nodeInfo
			}
		}

		if nodeMetricNames, ok := sysConfigMap["gaussdbv5_instance_id,gaussdbv5_pdb_id"]; ok {
			instances, err := listInstances("gaussdbv5_pdb_id")
			if err != nil {
				logs.Logger.Errorf("Get gauss db v5 pdbId data from service error: %s", err.Error())
				return gaussdbV5Info.LabelInfo, gaussdbV5Info.FilterMetrics
			}
			for _, instance := range instances {
				metrics := buildDimensionMetrics(nodeMetricNames, "SYS.GAUSSDBV5",
					[]cesmodel.MetricsDimension{{Name: "gaussdbv5_instance_id", Value: instance.InstanceInfo.InstanceId}, {Name: "gaussdbv5_pdb_id", Value: instance.Pdb.PdbId}})
				filterMetrics = append(filterMetrics, metrics...)
				nodeInfo := labelInfo{
					Name:  []string{"name", "epId", "pdbName"},
					Value: []string{instance.InstanceInfo.InstanceName, instance.InstanceInfo.EnterpriseProjectId, instance.Pdb.PdbName},
				}
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = nodeInfo
			}
		}

		if nodeMetricNames, ok := sysConfigMap["gaussdbv5_instance_id,gaussdbv5_pdb_id,gaussdbv5_pdb_component_id"]; ok {
			instances, err := listInstances("gaussdbv5_pdb_component_id")
			if err != nil {
				logs.Logger.Errorf("Get gauss db v5 componentId data from service error: %s", err.Error())
				return gaussdbV5Info.LabelInfo, gaussdbV5Info.FilterMetrics
			}
			for _, instance := range instances {
				metrics := buildDimensionMetrics(nodeMetricNames, "SYS.GAUSSDBV5",
					[]cesmodel.MetricsDimension{{Name: "gaussdbv5_instance_id", Value: instance.InstanceInfo.InstanceId}, {Name: "gaussdbv5_pdb_id", Value: instance.Pdb.PdbId}, {Name: "gaussdbv5_pdb_component_id", Value: instance.Component.ComponentId}})
				filterMetrics = append(filterMetrics, metrics...)
				nodeInfo := labelInfo{
					Name:  []string{"name", "epId", "pdbName"},
					Value: []string{instance.InstanceInfo.InstanceName, instance.InstanceInfo.EnterpriseProjectId, instance.Pdb.PdbName},
				}
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = nodeInfo
			}
		}

		gaussdbV5Info.LabelInfo = resourceInfos
		gaussdbV5Info.FilterMetrics = filterMetrics
		gaussdbV5Info.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return gaussdbV5Info.LabelInfo, gaussdbV5Info.FilterMetrics
}

func GetInstanceLabelInfo(instance GaussdbV5Instances) labelInfo {
	info := labelInfo{
		Name:  []string{"name", "epId"},
		Value: []string{instance.InstanceInfo.InstanceName, instance.InstanceInfo.EnterpriseProjectId},
	}
	keys, values := getTags(fmtTags(instance.InstanceInfo.Tags))
	info.Name = append(info.Name, keys...)
	info.Value = append(info.Value, values...)
	return info
}

func listInstances(dim string) ([]GaussdbV5Instances, error) {
	gaussDBClient := getHcClient(getEndpoint("gaussdbV5", "v1.0"))
	urlPath := fmt.Sprintf("/v3/%s/ces/instances", conf.ProjectID)
	reqGaussDBV5Builder := def.NewHttpRequestDefBuilder().WithMethod(http.MethodPost).WithPath(urlPath).WithResponse(new(GaussdbV5InstancesResponse)).WithContentType("application/json").WithRequestField(def.NewFieldDef().WithName("Body").WithLocationType(def.Body)).Build()
	options := &GaussdbV5InstancesRequest{
		Body: GaussdbV5InstancesRequestBody{
			Namespace: "SYS.GAUSSDBV5",
			Start:     0,
			Limit:     100,
			Query: []GaussdbV5InstanceSubDimReqInfo{
				{
					dim,
				},
			},
		},
	}
	var result []GaussdbV5Instances
	for {
		resp, err := gaussDBClient.Sync(options, reqGaussDBV5Builder)
		if err != nil {
			logs.Logger.Errorf("List gaussdb instance error: %s", err.Error())
			return nil, err
		}
		gaussDBV5Resp, ok := resp.(*GaussdbV5InstancesResponse)
		if !ok {
			logs.Logger.Error("Convert response to GaussdbV5InstancesResponse failed")
			return nil, fmt.Errorf("convert response to GaussdbV5InstancesResponse")
		}
		if len(gaussDBV5Resp.Instances) == 0 {
			break
		}
		result = append(result, gaussDBV5Resp.Instances...)
		options.Body.Start += 100
	}
	return result, nil
}
