package collector

import (
	"strings"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	ecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	ecsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var ecsInfo serversInfo

func (exporter *BaseHuaweiCloudExporter) getEcsResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	ecsInfo.Lock()
	defer ecsInfo.Unlock()
	if ecsInfo.LabelInfo == nil || time.Now().Unix() > ecsInfo.TTL {
		var servers []ResourceBaseInfo
		var err error
		if getResourceFromRMS("SYS.ECS") {
			servers, err = getAllServerFromRMS()
		} else {
			servers, err = getAllServer()
		}
		if err != nil {
			logs.Logger.Error("Get all Server error:", err.Error())
			return ecsInfo.LabelInfo, ecsInfo.FilterMetrics
		}

		sysConfigMap := getMetricConfigMap("SYS.ECS")
		for _, server := range servers {
			if metricNames, ok := sysConfigMap["instance_id"]; ok {
				metrics := buildSingleDimensionMetrics(metricNames, "SYS.ECS", "instance_id", server.ID)
				filterMetrics = append(filterMetrics, metrics...)
				info := labelInfo{
					Name:  []string{"hostname", "epId"},
					Value: []string{server.Name, server.EpId},
				}
				keys, values := getTags(server.Tags)
				info.Name = append(info.Name, keys...)
				info.Value = append(info.Value, values...)
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			}
		}

		ecsInfo.LabelInfo = resourceInfos
		ecsInfo.FilterMetrics = filterMetrics
		ecsInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return ecsInfo.LabelInfo, ecsInfo.FilterMetrics
}

func getECSClient() *ecs.EcsClient {
	return ecs.NewEcsClient(ecs.EcsClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(true)).
		WithEndpoint(getEndpoint("ecs", "v2")).Build())
}

func getAllServer() ([]ResourceBaseInfo, error) {
	limit := int32(1000)
	offset := int32(1)
	options := &ecsmodel.ListServersDetailsRequest{
		Limit:  &limit,
		Offset: &offset,
	}
	var servers []ResourceBaseInfo
	for {
		response, err := getECSClient().ListServersDetails(options)
		if err != nil {
			return servers, err
		}
		serversInfo := *response.Servers
		if len(serversInfo) == 0 {
			break
		}
		for _, server := range serversInfo {
			tags := make(map[string]string, len(*server.Tags))
			for _, tag := range *server.Tags {
				tagArray := strings.Split(tag, "=")
				tags[tagArray[0]] = tagArray[1]
			}
			servers = append(servers, ResourceBaseInfo{ID: server.Id, Name: server.Name,
				Tags: tags, EpId: *server.EnterpriseProjectId})
		}
		*options.Offset += 1
	}
	return servers, nil
}

func getAllServerFromRMS() ([]ResourceBaseInfo, error) {
	resp, err := listResources("ecs", "cloudservers")
	if err != nil {
		logs.Logger.Error(err)
		return nil, err
	}
	services := make([]ResourceBaseInfo, len(resp))
	for index, resource := range resp {
		services[index].ID = *resource.Id
		services[index].Name = *resource.Name
		services[index].EpId = *resource.EpId
		services[index].Tags = resource.Tags
	}
	return services, nil
}

func (exporter *BaseHuaweiCloudExporter) getAGTResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	ecsInfo.Lock()
	defer ecsInfo.Unlock()
	return ecsInfo.LabelInfo, nil
}
