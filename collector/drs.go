package collector

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	rmsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1/model"
)

type DrsInstanceInfo struct {
	ResourceBaseInfo
	DrsProperties map[string]string
}

var drsInfo serversInfo

type DRSInfo struct{}

func (getter DRSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	drsInfo.Lock()
	defer drsInfo.Unlock()
	if drsInfo.LabelInfo == nil || time.Now().Unix() > drsInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.DRS")
		metricNames, ok := sysConfigMap["instance_id"]
		if !ok {
			return drsInfo.LabelInfo, drsInfo.FilterMetrics
		}
		drsJobs, err := getAllDrsJobsFromRMS()
		if err == nil {
			for _, job := range drsJobs {
				metrics := buildSingleDimensionMetrics(metricNames, "SYS.DRS", "instance_id", job.ID)
				filterMetrics = append(filterMetrics, metrics...)
				info := labelInfo{
					Name:  []string{"name", "epId"},
					Value: []string{job.Name, job.EpId},
				}
				keys, values := getTags(job.Tags)
				info.Name = append(info.Name, keys...)
				info.Value = append(info.Value, values...)
				propertiesKeys, propertiesValues := getTags(job.DrsProperties)
				info.Name = append(info.Name, propertiesKeys...)
				info.Value = append(info.Value, propertiesValues...)
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			}
		}

		drsInfo.LabelInfo = resourceInfos
		drsInfo.FilterMetrics = filterMetrics
		drsInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return drsInfo.LabelInfo, drsInfo.FilterMetrics
}

func getAllDrsJobsFromRMS() ([]DrsInstanceInfo, error) {
	var resources []rmsmodel.ResourceEntity
	types := []string{"migrationJob", "synchronizationJob", "dataGuardJob", "subscriptionJob", "backupMigrationJob"}
	for i := range types {
		resourceList, err := listResources("drs", types[i])
		if err != nil {
			logs.Logger.Errorf("Failed to list resource of %s, error: %s", fmt.Sprintf("drs.%s", types[i]), err.Error())
			continue
		}
		resources = append(resources, resourceList...)
	}

	drsJobs := make([]DrsInstanceInfo, len(resources))
	for index, resource := range resources {
		drsProperties, err := fmtDrsProperties(resource.Properties)
		if err != nil {
			// properties转化整label失败，打印日志，继续增加其他字段
			logs.Logger.Errorf("Failed to fmt drs properties, error: %s", err.Error())
		}
		drsJobs[index].ID = *resource.Id
		drsJobs[index].Name = *resource.Name
		drsJobs[index].EpId = *resource.EpId
		drsJobs[index].Tags = resource.Tags
		drsJobs[index].DrsProperties = drsProperties
	}
	return drsJobs, nil
}

func fmtDrsProperties(properties map[string]interface{}) (map[string]string, error) {
	bytes, err := json.Marshal(properties)
	if err != nil {
		logs.Logger.Errorf("Marshal rds instance properties error: %s", err.Error())
		return nil, err
	}
	drsProperties := make(map[string]string)
	err = json.Unmarshal(bytes, &drsProperties)
	if err != nil {
		logs.Logger.Errorf("Unmarshal to RdsInstanceProperties error: %s", err.Error())
		return nil, err
	}

	return drsProperties, nil
}
