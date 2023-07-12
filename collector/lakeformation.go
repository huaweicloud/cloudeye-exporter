package collector

import (
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	lakeformation "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/lakeformation/v1"
	lakeformationmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/lakeformation/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var lakeformationInfo serversInfo

type LakeFormationInfo struct{}

func (getter LakeFormationInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	lakeformationInfo.Lock()
	defer lakeformationInfo.Unlock()
	if lakeformationInfo.LabelInfo == nil || time.Now().Unix() > lakeformationInfo.TTL {
		var instances []ResourceBaseInfo
		var err error
		if getResourceFromRMS("SYS.LakeFormation") {
			instances, err = getAllInstanceFromRMS()
		} else {
			instances, err = getAllLKInstance()
		}
		if err != nil {
			logs.Logger.Error("Get all instance error:", err.Error())
			return lakeformationInfo.LabelInfo, lakeformationInfo.FilterMetrics
		}

		sysConfigMap := getMetricConfigMap("SYS.LakeFormation")
		for _, instance := range instances {
			if metricNames, ok := sysConfigMap["instance_id"]; ok {
				metrics := buildSingleDimensionMetrics(metricNames, "SYS.LakeFormation", "instance_id", instance.ID)
				filterMetrics = append(filterMetrics, metrics...)
				info := labelInfo{
					Name:  []string{"name", "epId"},
					Value: []string{instance.Name, instance.EpId},
				}
				keys, values := getTags(instance.Tags)
				info.Name = append(info.Name, keys...)
				info.Value = append(info.Value, values...)
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			}
		}

		lakeformationInfo.LabelInfo = resourceInfos
		lakeformationInfo.FilterMetrics = filterMetrics
		lakeformationInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return lakeformationInfo.LabelInfo, lakeformationInfo.FilterMetrics
}

func getLakeFormationClient() *lakeformation.LakeFormationClient {
	return lakeformation.NewLakeFormationClient(lakeformation.LakeFormationClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(true)).
		WithEndpoint(getEndpoint("lakeformation", "v1")).Build())
}

func getAllLKInstance() ([]ResourceBaseInfo, error) {
	offset := 0
	limit := 1000
	request := &lakeformationmodel.ListLakeFormationInstancesRequest{
		InRecycleBin:        false,
		Offset:              int32(offset),
		Limit:               int32(limit),
		EnterpriseProjectId: "all_granted_eps",
	}
	var instances []lakeformationmodel.LakeFormationInstance
	for {
		response, err := getLakeFormationClient().ListLakeFormationInstances(request)
		if err != nil {
			logs.Logger.Errorf("Failed to get all lk instance : %s", err.Error())
			return nil, err
		}
		instances = append(instances, *response.Instances...)
		if len(*response.Instances) < limit {
			break
		}
		request.Offset += request.Limit
	}

	resources := make([]ResourceBaseInfo, len(instances))
	for i, instance := range instances {
		resources[i].ID = *instance.InstanceId
		resources[i].Name = *instance.Name
		resources[i].EpId = *instance.EnterpriseProjectId
	}
	return resources, nil
}

func getAllInstanceFromRMS() ([]ResourceBaseInfo, error) {
	return getResourcesBaseInfoFromRMS("lakeformation", "instance")
}
