package collector

import (
	"time"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

type EfsInstanceInfo struct {
	ResourceBaseInfo
	EfsProperties map[string]string
}

var efsInfo serversInfo

type EFSInfo struct{}

func (getter EFSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	efsInfo.Lock()
	defer efsInfo.Unlock()
	if efsInfo.LabelInfo == nil || time.Now().Unix() > efsInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.EFS")
		metricNames, ok := sysConfigMap["efs_instance_id"]
		if !ok {
			return efsInfo.LabelInfo, efsInfo.FilterMetrics
		}
		shares, err := getAllEfsShareFromRMS()
		if err == nil {
			for _, share := range shares {
				metrics := buildSingleDimensionMetrics(metricNames, "SYS.EFS", "efs_instance_id", share.ID)
				filterMetrics = append(filterMetrics, metrics...)
				info := labelInfo{
					Name:  []string{"name", "epId"},
					Value: []string{share.Name, share.EpId},
				}
				keys, values := getTags(share.Tags)
				info.Name = append(info.Name, keys...)
				info.Value = append(info.Value, values...)
				propertiesKeys, propertiesValues := getTags(share.EfsProperties)
				info.Name = append(info.Name, propertiesKeys...)
				info.Value = append(info.Value, propertiesValues...)
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			}
		}

		efsInfo.LabelInfo = resourceInfos
		efsInfo.FilterMetrics = filterMetrics
		efsInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return efsInfo.LabelInfo, efsInfo.FilterMetrics
}

func getAllEfsShareFromRMS() ([]EfsInstanceInfo, error) {
	resources, err := listResources("sfsturbo", "shares")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of sfsturbo.shares, error: %s", err.Error())
		return nil, err
	}
	shares := make([]EfsInstanceInfo, len(resources))
	for index, resource := range resources {
		var efsProperties map[string]string
		err := fmtResourceProperties(resource.Properties, &efsProperties)
		if err != nil {
			// properties转化label失败，打印日志，继续增加其他字段
			logs.Logger.Errorf("Failed to fmt efs properties, error: %s", err.Error())
		}
		shares[index].ID = *resource.Id
		shares[index].Name = *resource.Name
		shares[index].EpId = *resource.EpId
		shares[index].Tags = resource.Tags
		shares[index].EfsProperties = efsProperties
	}
	return shares, nil
}
