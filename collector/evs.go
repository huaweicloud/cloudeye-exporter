package collector

import (
	"fmt"
	"strings"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	evs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2"
	evsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

type EvsInfo struct {
	ResourceBaseInfo
	DiskName     string
	VolumeDiskId string
	ServerId     string
	Device       string
}

var evsInfo serversInfo

type EVSInfo struct{}

func (getter EVSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	evsInfo.Lock()
	defer evsInfo.Unlock()
	if evsInfo.LabelInfo == nil || time.Now().Unix() > evsInfo.TTL {
		var volumes []EvsInfo
		var err error
		if getResourceFromRMS("SYS.EVS") {
			volumes, err = getAllVolumeFromRMS()
		} else {
			volumes, err = getAllVolume()
		}
		if err != nil {
			logs.Logger.Error("Get all volumes error:", err.Error())
			return evsInfo.LabelInfo, evsInfo.FilterMetrics
		}

		allMetrics, err := listAllMetrics("SYS.EVS")
		if err != nil {
			logs.Logger.Error("Get all evs metrics error:", err.Error())
			return evsInfo.LabelInfo, evsInfo.FilterMetrics
		}

		metricMap := map[string]model.MetricInfoList{}
		var filteredMetrics []model.MetricInfoList
		for _, metric := range allMetrics {
			if !IsMetricInfoInWhiteList(metric) {
				continue
			}
			filteredMetrics = append(filteredMetrics, metric)
			resourceKey := GetResourceKeyFromMetricInfo(metric)
			metricMap[resourceKey] = metric
		}

		sysConfigMap := getMetricConfigMap("SYS.EVS")
		for _, volume := range volumes {
			if _, ok := sysConfigMap["disk_name"]; ok {
				info := labelInfo{
					Name:  []string{"id", "name", "epId", "serverId", "device"},
					Value: []string{volume.ID, volume.Name, volume.EpId, volume.ServerId, volume.Device},
				}
				keys, values := getTags(volume.Tags)
				info.Name = append(info.Name, keys...)
				info.Value = append(info.Value, values...)

				if _, ok := metricMap[volume.DiskName]; ok {
					resourceInfos[volume.DiskName] = info
					continue
				}
				if _, ok := metricMap[volume.VolumeDiskId]; ok {
					resourceInfos[volume.VolumeDiskId] = info
				}
			}
		}
		evsInfo.LabelInfo = resourceInfos
		evsInfo.FilterMetrics = filteredMetrics
		evsInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return evsInfo.LabelInfo, evsInfo.FilterMetrics
}

func getEVSClient() *evs.EvsClient {
	return evs.NewEvsClient(evs.EvsClientBuilder().WithCredential(
		authCredentialMap[conf.AuthMode](RegionServiceType)).
		WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
		WithEndpoint(getEndpoint("evs", "v2")).Build())
}

func getAllVolume() ([]EvsInfo, error) {
	volumes, err := listVolumesFromEvs()
	if err != nil {
		return nil, err
	}
	volumesInfo := make([]EvsInfo, 0, len(volumes))
	for _, disk := range volumes {
		for index := range disk.Attachments {
			volumesInfo = append(volumesInfo, EvsInfo{
				ResourceBaseInfo: ResourceBaseInfo{
					ID:   disk.Id,
					Name: disk.Name,
					Tags: disk.Tags,
					EpId: *disk.EnterpriseProjectId,
				},
				DiskName:     getDiskName(disk.Attachments[index].ServerId, disk.Attachments[index].Device),
				VolumeDiskId: getVolumeDiskId(disk.Attachments[index].ServerId, disk.Id),
				ServerId:     disk.Attachments[index].ServerId,
				Device:       disk.Attachments[index].Device})
		}
	}
	return volumesInfo, nil
}

func getVolumeDiskId(serverId string, diskId string) string {
	return fmt.Sprintf("%s-volume-%s", serverId, diskId)
}

func listVolumesFromEvs() ([]evsmodel.VolumeDetail, error) {
	var volumes []evsmodel.VolumeDetail
	epIds := getEpIdRequestPart()
	for _, epId := range epIds {
		tmpVolumes, err := listVolumesFromEvsByEpId(epId)
		if err != nil {
			logs.Logger.Errorf("Failed to list evs volumes, epId: %s, error: %s", epId, err.Error())
			return nil, err
		}
		volumes = append(volumes, tmpVolumes...)
	}
	return volumes, nil
}

func listVolumesFromEvsByEpId(epId string) ([]evsmodel.VolumeDetail, error) {
	limit := int32(1000)
	offset := int32(0)
	options := &evsmodel.ListVolumesRequest{
		Limit:               &limit,
		Offset:              &offset,
		EnterpriseProjectId: &epId,
	}
	var volumes []evsmodel.VolumeDetail
	for {
		response, err := getEVSClient().ListVolumes(options)
		if err != nil {
			return volumes, err
		}
		disksInfo := *response.Volumes
		if len(disksInfo) == 0 {
			break
		}
		volumes = append(volumes, *response.Volumes...)
		*options.Offset += 1
	}
	return volumes, nil
}

func getAllVolumeFromRMS() ([]EvsInfo, error) {
	resp, err := listResources("evs", "volumes")
	if err != nil {
		return nil, err
	}
	var volumes []EvsInfo
	for _, resource := range resp {
		var evsProperties RmsEvcProperties
		err := fmtResourceProperties(resource.Properties, &evsProperties)
		if err != nil {
			logs.Logger.Errorf("fmt evs properties error: %s", err.Error())
			continue
		}
		for _, attachment := range evsProperties.Attachments {
			volumes = append(volumes, EvsInfo{
				ResourceBaseInfo: ResourceBaseInfo{
					ID:   *resource.Id,
					Name: *resource.Name,
					EpId: *resource.EpId,
					Tags: resource.Tags,
				},
				DiskName:     getDiskName(attachment.ServerId, attachment.Device),
				VolumeDiskId: getVolumeDiskId(attachment.ServerId, *resource.Id),
				ServerId:     attachment.ServerId,
				Device:       attachment.Device,
			})
		}
	}
	return volumes, nil
}

func getDiskName(serverID, device string) string {
	deviceInfo := strings.Split(device, "/")
	if len(deviceInfo) > 0 {
		return fmt.Sprintf("%s-%s", serverID, deviceInfo[len(deviceInfo)-1])
	}
	return ""
}

type RmsEvcProperties struct {
	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	Device   string `json:"device"`
	ServerId string `json:"serverId"`
}
