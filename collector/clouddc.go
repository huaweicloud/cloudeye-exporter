package collector

import (
	"time"

	"github.com/huaweicloud/cloudeye-exporter/logs"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

type IMetalProperties struct {
	ImageName    string `json:"image_name"`
	ProductType  string `json:"product_type"`
	PowerStatus  string `json:"power_status"`
	IP           string `json:"ip"`
	HealthStatus string `json:"health_status"`
}

type IMetalInfo struct {
	ResourceBaseInfo
	IMetalProperties
}

type RackProperties struct {
	Description string `json:"description"`
	Location    string `json:"location"`
	Power       string `json:"power"`
	Dc          string `json:"dc"`
}

type RackInfo struct {
	ResourceBaseInfo
	RackProperties
}

type CloudDcInfo struct{}

var (
	cloudDcInfo serversInfo
)

func (getter CloudDcInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	cloudDcInfo.Lock()
	defer cloudDcInfo.Unlock()
	if cloudDcInfo.LabelInfo == nil || time.Now().Unix() > cloudDcInfo.TTL {
		var err error
		resourceInfos, filterMetrics, err = buildCloudDCResourceInfo()
		if err != nil {
			return cloudDcInfo.LabelInfo, cloudDcInfo.FilterMetrics
		}
		cloudDcInfo.LabelInfo = resourceInfos
		cloudDcInfo.FilterMetrics = filterMetrics
		cloudDcInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return cloudDcInfo.LabelInfo, cloudDcInfo.FilterMetrics
}

func buildCloudDCResourceInfo() (map[string]labelInfo, []model.MetricInfoList, error) {
	resourceInfos := map[string]labelInfo{}
	imetalInfos, err := getIMetalInstanceFromRMS()
	if err != nil {
		return nil, nil, err
	}
	rackInfos, err := getRackInstanceFromRMS()
	if err != nil {
		return nil, nil, err
	}

	for _, imetalInfo := range imetalInfos {
		info := labelInfo{
			Name:  []string{"name", "epId", "image_name", "product_type", "power_status", "ip", "health_status"},
			Value: []string{imetalInfo.Name, imetalInfo.EpId, imetalInfo.ImageName, imetalInfo.ProductType, imetalInfo.PowerStatus, imetalInfo.IP, imetalInfo.HealthStatus},
		}
		keys, values := getTags(imetalInfo.Tags)
		info.Name = append(info.Name, keys...)
		info.Value = append(info.Value, values...)
		resourceInfos[imetalInfo.ID] = info
	}

	for _, rackInfo := range rackInfos {
		info := labelInfo{
			Name:  []string{"name", "epId", "description", "location", "power", "dc"},
			Value: []string{rackInfo.Name, rackInfo.EpId, rackInfo.Description, rackInfo.Location, rackInfo.Power, rackInfo.Dc},
		}
		keys, values := getTags(rackInfo.Tags)
		info.Name = append(info.Name, keys...)
		info.Value = append(info.Value, values...)
		resourceInfos[rackInfo.ID] = info
	}

	allMetrics, err := listAllMetrics("SYS.CloudDC")
	if err != nil {
		logs.Logger.Errorf("Get all metrics of SYS.CloudDC error: %s", err.Error())
		return nil, nil, err
	}
	var filteredMetrics []model.MetricInfoList
	for _, metricInfo := range allMetrics {
		if !IsMetricInfoInWhiteList(metricInfo) {
			continue
		}
		resourceKey := getCloudDCResourceKey(model.BatchMetricData{Dimensions: &metricInfo.Dimensions})
		if _, ok := resourceInfos[resourceKey]; !ok {
			continue
		}
		filteredMetrics = append(filteredMetrics, metricInfo)
	}
	return resourceInfos, filteredMetrics, nil
}

func getIMetalInstanceFromRMS() ([]IMetalInfo, error) {
	imetalInfos := make([]IMetalInfo, 0)
	resps, err := listResources("clouddc", "imetal")
	if err != nil {
		logs.Logger.Errorf("Get clouddc imetal resources error: %s", err.Error())
		return nil, err
	}
	for _, resource := range resps {
		var imetalProperties IMetalProperties
		err = fmtResourceProperties(resource.Properties, &imetalProperties)
		if err != nil {
			logs.Logger.Errorf("fmt imetal properties error: %s", err.Error())
			continue
		}
		imetalInfos = append(imetalInfos, IMetalInfo{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   *resource.Id,
				Name: *resource.Name,
				EpId: *resource.EpId,
				Tags: resource.Tags,
			},
			IMetalProperties: imetalProperties,
		})
	}
	return imetalInfos, nil
}

func getRackInstanceFromRMS() ([]RackInfo, error) {
	rackInfos := make([]RackInfo, 0)
	resps, err := listResources("clouddc", "irack")
	if err != nil {
		logs.Logger.Errorf("Get clouddc rack resources error: %s", err.Error())
		return nil, err
	}
	for _, resource := range resps {
		var rackProperties RackProperties
		err = fmtResourceProperties(resource.Properties, &rackProperties)
		if err != nil {
			logs.Logger.Errorf("fmt rack properties error: %s", err.Error())
			continue
		}
		rackInfos = append(rackInfos, RackInfo{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   *resource.Id,
				Name: *resource.Name,
				EpId: *resource.EpId,
				Tags: resource.Tags,
			},
			RackProperties: rackProperties,
		})
	}
	return rackInfos, nil
}
