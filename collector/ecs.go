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

type ECSInfo struct{}

func (getter ECSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	ecsInfo.Lock()
	defer ecsInfo.Unlock()
	if ecsInfo.LabelInfo == nil || time.Now().Unix() > ecsInfo.TTL {
		var servers []EcsInstancesInfo
		var err error
		if getResourceFromRMS("SYS.ECS") {
			servers, err = getAllServerFromRMS("ecs", "cloudservers")
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
					Name:  []string{"hostname", "epId", "hostIP"},
					Value: []string{server.Name, server.EpId, server.IP},
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

type EcsInstancesInfo struct {
	ResourceBaseInfo
	IP string
}

func getECSClient() *ecs.EcsClient {
	return ecs.NewEcsClient(ecs.EcsClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(true)).
		WithEndpoint(getEndpoint("ecs", "v2")).Build())
}
func getAllServer() ([]EcsInstancesInfo, error) {
	limit := int32(1000)
	offset := int32(1)
	options := &ecsmodel.ListServersDetailsRequest{
		Limit:  &limit,
		Offset: &offset,
	}
	var servers []EcsInstancesInfo
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
				if len(tagArray) == 2 {
					tags[tagArray[0]] = tagArray[1]
				}
			}
			servers = append(servers, EcsInstancesInfo{
				ResourceBaseInfo: ResourceBaseInfo{
					ID: server.Id, Name: server.Name,
					Tags: tags, EpId: *server.EnterpriseProjectId},
				IP: getIPFromEcsInfo(server.Addresses),
			})
		}
		*options.Offset += 1
	}
	return servers, nil
}

func getIPFromEcsInfo(addresses map[string][]ecsmodel.ServerAddress) string {
	var ips []string
	for _, address := range addresses {
		for i := range address {
			ips = append(ips, address[i].Addr)
		}
	}
	return strings.Join(ips, ",")
}

func getAllServerFromRMS(provider, resourceType string) ([]EcsInstancesInfo, error) {
	resp, err := listResources(provider, resourceType)
	if err != nil {
		return nil, err
	}
	services := make([]EcsInstancesInfo, len(resp))
	for index, resource := range resp {
		var properties EcsProperties
		err := fmtResourceProperties(resource.Properties, &properties)
		if err != nil {
			logs.Logger.Errorf("fmt ecs properties error: %s", err.Error())
			continue
		}
		services[index].ID = *resource.Id
		services[index].Name = *resource.Name
		services[index].EpId = *resource.EpId
		services[index].Tags = resource.Tags
		services[index].IP = getIPInfoFromProperties(&properties)
	}
	return services, nil
}

type EcsProperties struct {
	Addresses []struct {
		Addr string
	} `json:"addresses"`
}

func getIPInfoFromProperties(properties *EcsProperties) string {
	var ips []string
	for i := range properties.Addresses {
		ips = append(ips, properties.Addresses[i].Addr)
	}
	return strings.Join(ips, ",")
}

type AGTECSInfo struct{}

func (getter AGTECSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	ecsInfo.Lock()
	defer ecsInfo.Unlock()
	return ecsInfo.LabelInfo, nil
}
