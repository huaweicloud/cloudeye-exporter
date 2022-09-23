package collector

import (
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	ddm "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ddm/v1"
	ddmmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ddm/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

type DdmsInstanceInfo struct {
	Instance ddmmodel.ShowInstanceBeanResponse
	Nodes    []ddmmodel.ShowNodeResponse
}

var ddmsInfo serversInfo

type DDMSInfo struct{}

func (getter DDMSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	ddmsInfo.Lock()
	defer ddmsInfo.Unlock()
	if ddmsInfo.LabelInfo == nil || time.Now().Unix() > ddmsInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.DDMS")
		metricNames, ok := sysConfigMap["instance_id,node_id"]
		if !ok {
			return ddmsInfo.LabelInfo, ddmsInfo.FilterMetrics
		}
		instances, err := getAllDdmsInstances()
		if err != nil {
			logs.Logger.Errorf("Get all ddms instances: %s", err.Error())
			return ddmsInfo.LabelInfo, ddmsInfo.FilterMetrics
		}
		for _, instance := range instances {
			for _, node := range instance.Nodes {
				metrics := buildDimensionMetrics(metricNames, "SYS.DDMS",
					[]model.MetricsDimension{{Name: "instance_id", Value: instance.Instance.Id}, {Name: "node_id", Value: *node.NodeId}})
				filterMetrics = append(filterMetrics, metrics...)
				nodeInfo := labelInfo{
					Name: []string{"instanceName", "instanceAccessIP", "instanceAccessPost", "epId", "nodeName", "nodePrivateIP", "nodeFloatingIP", "nodeResSubnetIP"},
					Value: []string{instance.Instance.Name, instance.Instance.AccessIp, instance.Instance.AccessPort, instance.Instance.EnterpriseProjectId,
						*node.Name, getDefaultString(node.PrivateIp), getDefaultString(node.FloatingIp), getDefaultString(node.ResSubnetIp)},
				}
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = nodeInfo
			}
		}
		ddmsInfo.LabelInfo = resourceInfos
		ddmsInfo.FilterMetrics = filterMetrics
		ddmsInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return ddmsInfo.LabelInfo, ddmsInfo.FilterMetrics
}

func getAllDdmsInstances() ([]DdmsInstanceInfo, error) {
	limit := int32(100)
	offset := int32(0)
	options := &ddmmodel.ListInstancesRequest{Limit: &limit, Offset: &offset}
	var instances []DdmsInstanceInfo
	for {
		response, err := getDDMSClient().ListInstances(options)
		if err != nil {
			logs.Logger.Errorf("List ddms instances: %s", err.Error())
			return instances, err
		}
		instancesInfo := *response.Instances
		if len(instancesInfo) == 0 {
			break
		}
		for _, instance := range instancesInfo {
			nodes, err := getDdmsInstanceNodes(instance.Id)
			if err != nil {
				logs.Logger.Errorf("Get nodes of ddms instance [%s] error: %s", instance.Id, err.Error())
				continue
			}
			instances = append(instances, DdmsInstanceInfo{instance, nodes})
		}
		*options.Offset += limit
	}
	return instances, nil
}

func getDdmsInstanceNodes(instanceId string) ([]ddmmodel.ShowNodeResponse, error) {
	limit := int32(100)
	offset := int32(0)
	options := &ddmmodel.ListNodesRequest{InstanceId: instanceId, Limit: &limit, Offset: &offset}
	var nodes []ddmmodel.ShowNodeResponse
	client := getDDMSClient()
	for {
		response, err := client.ListNodes(options)
		if err != nil {
			return nodes, err
		}
		if len(*response.Nodes) == 0 {
			break
		}
		for _, node := range *response.Nodes {
			nodeInfo, err := client.ShowNode(&ddmmodel.ShowNodeRequest{InstanceId: instanceId, NodeId: *node.NodeId})
			if err != nil {
				return nodes, err
			}
			nodes = append(nodes, *nodeInfo)
		}
		*options.Offset += limit
	}
	return nodes, nil
}

func getDDMSClient() *ddm.DdmClient {
	return ddm.NewDdmClient(ddm.DdmClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(true)).
		WithEndpoint(getEndpoint("ddm", "v1")).Build())
}
