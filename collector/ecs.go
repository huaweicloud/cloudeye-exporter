package collector

import (
	"strings"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	ecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	ecsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var (
	ecsInfo    serversInfo
	agtEcsInfo serversInfo
)

type ECSInfo struct{}

func (getter ECSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	extendInfo := make(map[string]map[string]string)
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
					Name:  []string{"hostname", "epId", "hostIP", "fixedIP", "floatingIP"},
					Value: []string{server.Name, server.EpId, server.IP, server.FixedIP, server.FloatingIP},
				}
				keys, values := getTags(server.Tags)
				info.Name = append(info.Name, keys...)
				info.Value = append(info.Value, values...)
				extendInfo[server.ID] = server.DiskMap
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			}
		}
		ecsInfo.LabelInfo = resourceInfos
		ecsInfo.FilterMetrics = filterMetrics
		tmpMap := make(map[string]interface{})
		for key, value := range extendInfo {
			tmpMap[key] = value
		}
		ecsInfo.ExtendInfo = tmpMap
		ecsInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return ecsInfo.LabelInfo, ecsInfo.FilterMetrics
}

type EcsInstancesInfo struct {
	ResourceBaseInfo
	IP         string
	FloatingIP string
	FixedIP    string
	DiskMap    map[string]string // key: 磁盘名称(/dev/vda), value: evs_id
}

func getECSClient() *ecs.EcsClient {
	return ecs.NewEcsClient(ecs.EcsClientBuilder().WithCredential(
		authCredentialMap[conf.AuthMode](RegionServiceType)).
		WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
		WithEndpoint(getEndpoint("ecs", "v2")).Build())
}

func getAllServer() ([]EcsInstancesInfo, error) {
	var servers []EcsInstancesInfo
	epIds := getEpIdRequestPart()
	for _, epId := range epIds {
		tmpServers, err := getAllServerByEpId(epId)
		if err != nil {
			logs.Logger.Errorf("Failed to list ecs server, epId: %s, error: %s", epId, err.Error())
			return nil, err
		}
		servers = append(servers, tmpServers...)
	}
	return servers, nil
}

func getAllServerByEpId(epId string) ([]EcsInstancesInfo, error) {
	limit := int32(1000)
	offset := int32(1)
	options := &ecsmodel.ListServersDetailsRequest{
		Limit:               &limit,
		Offset:              &offset,
		EnterpriseProjectId: &epId,
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
			ips, fixedIps, floatingIps := getIPFromEcsInfo(server.Addresses)
			servers = append(servers, EcsInstancesInfo{
				ResourceBaseInfo: ResourceBaseInfo{
					ID: server.Id, Name: server.Name,
					Tags: tags, EpId: *server.EnterpriseProjectId},
				IP: ips, FixedIP: fixedIps, FloatingIP: floatingIps,
				DiskMap: getDiskInfoFromEcsInfo(server.OsExtendedVolumesvolumesAttached),
			})
		}
		*options.Offset += 1
	}
	return servers, nil
}

func getDiskInfoFromEcsInfo(attachedVolumes []ecsmodel.ServerExtendVolumeAttachment) map[string]string {
	evsMap := make(map[string]string, len(attachedVolumes))
	for i := range attachedVolumes {
		devicePath := strings.Split(attachedVolumes[i].Device, "/")
		evsMap[devicePath[len(devicePath)-1]] = attachedVolumes[i].Id
	}
	return evsMap
}

func getIPFromEcsInfo(addresses map[string][]ecsmodel.ServerAddress) (string, string, string) {
	var ips []string
	var fixedIps []string
	var floatingIps []string
	for _, address := range addresses {
		for i := range address {
			ips = append(ips, address[i].Addr)
			if address[i].OSEXTIPStype != nil && address[i].OSEXTIPStype.Value() == "fixed" {
				fixedIps = append(fixedIps, address[i].Addr)
			}
			if address[i].OSEXTIPStype != nil && address[i].OSEXTIPStype.Value() == "floating" {
				floatingIps = append(floatingIps, address[i].Addr)
			}
		}
	}
	return strings.Join(ips, ","), strings.Join(fixedIps, ","), strings.Join(floatingIps, ",")
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
		services[index].DiskMap = getEvsIDFromProperties(&properties)
		services[index].IP, services[index].FixedIP, services[index].FloatingIP = getIPInfoFromProperties(&properties)
	}
	return services, nil
}

type EcsProperties struct {
	Addresses []struct {
		Addr         string
		OsExtIpsType string
	} `json:"addresses"`
	ExtVolumesAttached []struct {
		Id     string `json:"id"`
		Device string `json:"device"`
	} `json:"ExtVolumesAttached"`
}

func getIPInfoFromProperties(properties *EcsProperties) (string, string, string) {
	var ips []string
	var fixedIps []string
	var floatingIps []string
	for i := range properties.Addresses {
		ips = append(ips, properties.Addresses[i].Addr)
		if properties.Addresses[i].OsExtIpsType == "fixed" {
			fixedIps = append(fixedIps, properties.Addresses[i].Addr)
		}
		if properties.Addresses[i].OsExtIpsType == "floating" {
			floatingIps = append(floatingIps, properties.Addresses[i].Addr)
		}
	}
	return strings.Join(ips, ","), strings.Join(fixedIps, ","), strings.Join(floatingIps, ",")
}

func getEvsIDFromProperties(properties *EcsProperties) map[string]string {
	evsMap := make(map[string]string, len(properties.ExtVolumesAttached))
	for i := range properties.ExtVolumesAttached {
		devicePath := strings.Split(properties.ExtVolumesAttached[i].Device, "/")
		evsMap[devicePath[len(devicePath)-1]] = properties.ExtVolumesAttached[i].Id
	}
	return evsMap
}

type AGTECSInfo struct{}

func (getter AGTECSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	agtEcsInfo.Lock()
	defer agtEcsInfo.Unlock()
	if agtEcsInfo.LabelInfo == nil {
		agtEcsInfo.FilterMetrics = getECSAGTMetrics()
		agtEcsInfo.LabelInfo = ecsInfo.LabelInfo
		agtEcsInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	if time.Now().Unix() > agtEcsInfo.TTL {
		go func() {
			metrics := getECSAGTMetrics()
			agtEcsInfo.Lock()
			defer agtEcsInfo.Unlock()
			agtEcsInfo.FilterMetrics = metrics
			agtEcsInfo.LabelInfo = ecsInfo.LabelInfo
			agtEcsInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
		}()
	}
	return agtEcsInfo.LabelInfo, agtEcsInfo.FilterMetrics
}

func getECSAGTMetrics() []model.MetricInfoList {
	allMetrics, err := listAllMetrics("AGT.ECS")
	var filteredMetrics []model.MetricInfoList
	if err != nil {
		logs.Logger.Errorf("Get all metrics of AGT.ECS error: %s", err.Error())
		return filteredMetrics
	}
	if ecsInfo.LabelInfo == nil {
		logs.Logger.Info("No ecs resource info found, skip to query agent metrics info")
		return filteredMetrics
	}
	ecsInfo.Lock()
	defer ecsInfo.Unlock()
	for _, metric := range allMetrics {
		serverKey := getServerResourceKeyFromMetricInfo(metric)
		if serverKey == "" {
			continue
		}
		if _, ok := ecsInfo.LabelInfo[serverKey]; !ok {
			continue
		}
		//白名单校验通过，查询当前指标对应指标数据
		if IsMetricInfoInWhiteList(metric) {
			filteredMetrics = append(filteredMetrics, metric)
		}
	}
	return filteredMetrics
}
