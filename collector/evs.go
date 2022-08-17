package collector

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	evs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2"
	evsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

type EvsInfo struct {
	ResourceBaseInfo
	DiskName string
	ServerId string
	Device   string
}

var evsInfo serversInfo

type EVSInfo struct{}

func (getter EVSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
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

		sysConfigMap := getMetricConfigMap("SYS.EVS")
		for _, volume := range volumes {
			if metricNames, ok := sysConfigMap["disk_name"]; ok {
				metrics := buildSingleDimensionMetrics(metricNames, "SYS.EVS", "disk_name", volume.DiskName)
				filterMetrics = append(filterMetrics, metrics...)
				info := labelInfo{
					Name:  []string{"id", "name", "epId", "serverId", "device"},
					Value: []string{volume.ID, volume.Name, volume.EpId, volume.ServerId, volume.Device},
				}
				keys, values := getTags(volume.Tags)
				info.Name = append(info.Name, keys...)
				info.Value = append(info.Value, values...)
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			}
		}

		evsInfo.LabelInfo = resourceInfos
		evsInfo.FilterMetrics = filterMetrics
		evsInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return evsInfo.LabelInfo, evsInfo.FilterMetrics
}

func getEVSClient() *evs.EvsClient {
	return evs.NewEvsClient(evs.EvsClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(true)).
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
				DiskName: getDiskName(disk.Attachments[index].ServerId, disk.Attachments[index].Device),
				ServerId: disk.Attachments[index].ServerId,
				Device:   disk.Attachments[index].Device})
		}
	}
	return volumesInfo, nil
}

func listVolumesFromEvs() ([]evsmodel.VolumeDetail, error) {
	limit := int32(1000)
	offset := int32(0)
	options := &evsmodel.ListVolumesRequest{
		Limit:  &limit,
		Offset: &offset,
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
		properties, err := fmtEvcProperties(resource.Properties)
		if err != nil {
			continue
		}
		for _, attachment := range properties.Attachments {
			volumes = append(volumes, EvsInfo{
				ResourceBaseInfo: ResourceBaseInfo{
					ID:   *resource.Id,
					Name: *resource.Name,
					EpId: *resource.EpId,
					Tags: resource.Tags,
				},
				DiskName: getDiskName(attachment.ServerId, attachment.Device),
				ServerId: attachment.ServerId,
				Device:   attachment.Device,
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

func fmtEvcProperties(properties map[string]interface{}) (*RmsEvcProperties, error) {
	bytes, err := json.Marshal(properties)
	if err != nil {
		logs.Logger.Errorf("Marshal evs properties error: %s", err.Error())
		return nil, err
	}
	var volumeDetail RmsEvcProperties
	err = json.Unmarshal(bytes, &volumeDetail)
	if err != nil {
		logs.Logger.Errorf("Unmarshal to RmsEvcProperties error: %s", err.Error())
		return nil, err
	}

	return &volumeDetail, nil
}
