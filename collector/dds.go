package collector

import (
	"time"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	dds "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dds/v3"
	ddsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dds/v3/model"
)

type DdsInstanceInfo struct {
	ResourceBaseInfo
	Mode             string
	Engine           string
	DatastoreType    string
	DatastoreVersion string
	Nodes            []ddsmodel.NodeItem
}

var ddsInfo serversInfo

type DDSInfo struct{}

func (getter DDSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	ddsInfo.Lock()
	defer ddsInfo.Unlock()
	if ddsInfo.LabelInfo == nil || time.Now().Unix() > ddsInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.DDS")
		metricNames, ok := sysConfigMap["mongodb_instance_id"]
		if !ok {
			return ddsInfo.LabelInfo, ddsInfo.FilterMetrics
		}
		ddsInstances, err := getAllDdsInstances()
		if err != nil {
			logs.Logger.Errorf("Get all dds instances: %s", err.Error())
			return ddsInfo.LabelInfo, ddsInfo.FilterMetrics
		}
		for _, instance := range ddsInstances {
			metrics := buildSingleDimensionMetrics(metricNames, "SYS.DDS", "mongodb_instance_id", instance.ID)
			filterMetrics = append(filterMetrics, metrics...)
			info := labelInfo{
				Name:  []string{"name", "epId", "mode", "engine", "datastoreType", "datastoreVersion"},
				Value: []string{instance.Name, instance.EpId, instance.Mode, instance.Engine, instance.DatastoreType, instance.DatastoreVersion},
			}
			keys, values := getTags(instance.Tags)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info

			nodeMetricNames, ok := sysConfigMap["mongodb_instance_id,mongodb_node_id"]
			if !ok {
				continue
			}
			for _, node := range instance.Nodes {
				metrics := buildDimensionMetrics(nodeMetricNames, "SYS.DDS",
					[]model.MetricsDimension{{Name: "mongodb_instance_id", Value: instance.ID}, {Name: "mongodb_node_id", Value: node.Id}})
				filterMetrics = append(filterMetrics, metrics...)
				nodeInfo := labelInfo{
					Name:  []string{"nodeName", "role", "privateIp", "publicIp"},
					Value: []string{node.Name, node.Role, node.PrivateIp, node.PublicIp},
				}
				nodeInfo.Name = append(nodeInfo.Name, info.Name...)
				nodeInfo.Value = append(nodeInfo.Value, info.Value...)
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = nodeInfo
			}
		}
		ddsInfo.LabelInfo = resourceInfos
		ddsInfo.FilterMetrics = filterMetrics
		ddsInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return ddsInfo.LabelInfo, ddsInfo.FilterMetrics
}

func getAllDdsInstances() ([]DdsInstanceInfo, error) {
	limit := int32(100)
	offset := int32(0)
	options := &ddsmodel.ListInstancesRequest{Limit: &limit, Offset: &offset}
	var instances []DdsInstanceInfo
	for {
		response, err := getDDSClient().ListInstances(options)
		if err != nil {
			logs.Logger.Errorf("list dds instances: %s", err.Error())
			return instances, err
		}
		instancesInfo := *response.Instances
		if len(instancesInfo) == 0 {
			break
		}
		for _, instance := range instancesInfo {
			instances = append(instances, fmtDdsInstance(instance))
		}
		*options.Offset += limit
	}
	return instances, nil
}

func fmtDdsInstance(instance ddsmodel.QueryInstanceResponse) DdsInstanceInfo {
	tags := make(map[string]string, len(instance.Tags))
	for _, tag := range instance.Tags {
		tags[tag.Key] = tag.Value
	}
	var nodes []ddsmodel.NodeItem
	for _, group := range instance.Groups {
		nodes = append(nodes, group.Nodes...)
	}
	return DdsInstanceInfo{
		ResourceBaseInfo: ResourceBaseInfo{
			ID: instance.Id, Name: instance.Name,
			Tags: tags, EpId: instance.EnterpriseProjectId,
		},
		Mode:             instance.Mode,
		Engine:           instance.Engine,
		DatastoreType:    instance.Datastore.Type,
		DatastoreVersion: instance.Datastore.Version,
		Nodes:            nodes,
	}
}

func getDDSClient() *dds.DdsClient {
	return dds.NewDdsClient(dds.DdsClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(true)).
		WithEndpoint(getEndpoint("dds", "v3")).Build())
}
