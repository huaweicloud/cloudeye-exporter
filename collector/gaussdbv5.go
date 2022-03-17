package collector

import (
	"encoding/json"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	gaussdbforopengauss "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/gaussdbforopengauss/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/gaussdbforopengauss/v3/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

type GaussdbV5Node struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Role             string `json:"role"`
	Status           string `json:"status"`
	AvailabilityZone string `json:"availability_zone"`
}

var gaussdbV5Info serversInfo

func (exporter *BaseHuaweiCloudExporter) getGaussdbV5ResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]cesmodel.MetricInfoList, 0)
	gaussdbV5Info.Lock()
	defer gaussdbV5Info.Unlock()
	if gaussdbV5Info.LabelInfo == nil || time.Now().Unix() > gaussdbV5Info.TTL {
		sysConfigMap := getMetricConfigMap("SYS.GAUSSDBV5")
		if instanceMetricNames, ok := sysConfigMap["gaussdbv5_instance_id"]; ok {
			for _, instance := range listInstances() {
				metrics := buildSingleDimensionMetrics(instanceMetricNames, "SYS.GAUSSDBV5", "gaussdbv5_instance_id", instance.Id)
				filterMetrics = append(filterMetrics, metrics...)
				info := labelInfo{
					Name:  []string{"name"},
					Value: []string{instance.Name},
				}
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info

				if nodeMetricNames, ok := sysConfigMap["gaussdbv5_instance_id,gaussdbv5_node_id"]; ok {
					nodes, err := fmtNodesInfo(instance.Nodes)
					if err != nil {
						continue
					}
					for _, node := range *nodes {
						metrics := buildDimensionMetrics(nodeMetricNames, "SYS.GAUSSDBV5",
							[]cesmodel.MetricsDimension{{Name: "gaussdbv5_instance_id", Value: instance.Id}, {Name: "gaussdbv5_node_id", Value: node.ID}})
						filterMetrics = append(filterMetrics, metrics...)
						nodeInfo := labelInfo{
							Name:  []string{"nodeName", "role", "status", "availability_zone"},
							Value: []string{node.Name, node.Role, node.Status, node.AvailabilityZone},
						}
						nodeInfo.Name = append(nodeInfo.Name, info.Name...)
						nodeInfo.Value = append(nodeInfo.Value, info.Value...)
						resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = nodeInfo
					}
				}
			}

		}
		gaussdbV5Info.LabelInfo = resourceInfos
		gaussdbV5Info.FilterMetrics = filterMetrics
		gaussdbV5Info.TTL = time.Now().Add(TTL).Unix()
	}
	return gaussdbV5Info.LabelInfo, gaussdbV5Info.FilterMetrics
}

func getGaussdbforopengaussClient() *gaussdbforopengauss.GaussDBforopenGaussClient {
	return gaussdbforopengauss.NewGaussDBforopenGaussClient(gaussdbforopengauss.GaussDBforopenGaussClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(true)).
		WithEndpoint(getEndpoint("gaussdb-opengauss", "v3")).Build())
}

func listInstances() []model.ListInstanceResponse {
	limit := int32(100)
	request := &model.ListInstancesRequest{Limit: &limit}
	var instances []model.ListInstanceResponse
	for {
		response, err := getGaussdbforopengaussClient().ListInstances(request)
		if err != nil {
			logs.Logger.Errorf("list opengauss instances error: %s", err.Error())
			return instances
		}
		pageInstances := *response.Instances
		if len(pageInstances) == 0 {
			break
		}
		instances = append(instances, pageInstances...)
		offset := int32(len(instances))
		request.Offset = &offset
	}
	return instances
}

func fmtNodesInfo(nodeInfo []interface{}) (*[]GaussdbV5Node, error) {
	bytes, err := json.Marshal(nodeInfo)
	if err != nil {
		logs.Logger.Errorf("Marshal gaussdbv5 node error: %s", err.Error())
		return nil, err
	}
	var nodes []GaussdbV5Node
	err = json.Unmarshal(bytes, &nodes)
	if err != nil {
		logs.Logger.Errorf("Unmarshal to Gaussdbv5Node error: %s", err.Error())
		return nil, err
	}

	return &nodes, nil
}
