package collector

import (
	"time"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	cbr "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cbr/v1"
	cbrmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cbr/v1/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

var cbrInfo serversInfo

type CBRInfo struct{}

func (getter CBRInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	cbrInfo.Lock()
	defer cbrInfo.Unlock()
	if cbrInfo.LabelInfo == nil || time.Now().Unix() > cbrInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.CBR")
		var cbrInstances []ResourceBaseInfo
		var err error
		if getResourceFromRMS("SYS.CBR") {
			cbrInstances, err = getAllCbrInstancesFromRMS()
		} else {
			cbrInstances, err = getAllCbrInstancesFromCBR()
		}
		if err != nil {
			logs.Logger.Errorf("Failed to get cbr instances, error: %s", err.Error())
			return nil, nil
		}

		for _, instance := range cbrInstances {
			if metricNames, ok := sysConfigMap["instance_id"]; ok {
				metrics := buildSingleDimensionMetrics(metricNames, "SYS.CBR", "instance_id", instance.ID)
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

		cbrInfo.LabelInfo = resourceInfos
		cbrInfo.FilterMetrics = filterMetrics
		cbrInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return cbrInfo.LabelInfo, cbrInfo.FilterMetrics
}

func getAllCbrInstancesFromRMS() ([]ResourceBaseInfo, error) {
	return getResourcesBaseInfoFromRMS("cbr", "vault")
}

func getAllCbrInstancesFromCBR() ([]ResourceBaseInfo, error) {
	limit := int32(1000)
	offset := int32(0)
	request := &cbrmodel.ListVaultRequest{Limit: &limit, Offset: &offset}
	var cbrInstances []ResourceBaseInfo
	for {
		response, err := getCBRClient().ListVault(request)
		if err != nil {
			logs.Logger.Errorf("Failed to list vault, error: %s", err.Error())
			return nil, err
		}
		if len(*response.Vaults) == 0 {
			break
		}
		for _, vault := range *response.Vaults {
			cbrInstances = append(cbrInstances, ResourceBaseInfo{
				ID: vault.Id, Name: vault.Name,
				EpId: *vault.EnterpriseProjectId, Tags: fmtTags(vault.Tags),
			})
		}
		*request.Offset += limit
	}
	return cbrInstances, nil
}

func getCBRClient() *cbr.CbrClient {
	return cbr.NewCbrClient(cbr.CbrClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(true)).
		WithEndpoint(getEndpoint("cbr", "v3")).Build())
}
