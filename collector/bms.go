package collector

import (
	"time"

	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var bmsInfo serversInfo

type BMSInfo struct{}

func (getter BMSInfo) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]cesmodel.MetricInfoList, 0)
	extendInfo := make(map[string]map[string]string)
	bmsInfo.Lock()
	defer bmsInfo.Unlock()
	if bmsInfo.LabelInfo == nil || time.Now().Unix() > bmsInfo.TTL {
		services, err := getAllServerFromRMS("bms", "servers")
		if err != nil {
			logs.Logger.Error("Get all bms server error:", err.Error())
			return bmsInfo.LabelInfo, bmsInfo.FilterMetrics
		}
		sysConfigMap := getMetricConfigMap("SYS.BMS")
		if metricNames, ok := sysConfigMap["instance_id"]; ok {
			for _, instance := range services {
				metrics := buildSingleDimensionMetrics(metricNames, "SYS.BMS", "instance_id", instance.ID)
				filterMetrics = append(filterMetrics, metrics...)
				info := labelInfo{
					Name:  []string{"name", "epId", "ip", "fixedIp", "floatingIp"},
					Value: []string{instance.Name, instance.EpId, instance.IP, instance.FixedIP, instance.FloatingIP},
				}
				keys, values := getTags(instance.Tags)
				info.Name = append(info.Name, keys...)
				info.Value = append(info.Value, values...)
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
				extendInfo[instance.ID] = instance.DiskMap
			}
		}
		bmsInfo.LabelInfo = resourceInfos
		bmsInfo.FilterMetrics = filterMetrics
		tmpMap := make(map[string]interface{})
		for key, value := range extendInfo {
			tmpMap[key] = value
		}
		bmsInfo.ExtendInfo = tmpMap
		bmsInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return bmsInfo.LabelInfo, bmsInfo.FilterMetrics
}

type SERVICEBMSInfo struct{}

var serviceBmsInfo serversInfo

func (getter SERVICEBMSInfo) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	serviceBmsInfo.Lock()
	defer serviceBmsInfo.Unlock()
	if serviceBmsInfo.LabelInfo == nil || time.Now().Unix() > serviceBmsInfo.TTL {
		serviceBmsInfo.FilterMetrics = getServiceBMSMetrics()
		serviceBmsInfo.LabelInfo = bmsInfo.LabelInfo
		serviceBmsInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return serviceBmsInfo.LabelInfo, serviceBmsInfo.FilterMetrics
}

func getServiceBMSMetrics() []cesmodel.MetricInfoList {
	allMetrics, err := listAllMetrics("SERVICE.BMS")
	var filteredMetrics []cesmodel.MetricInfoList
	if err != nil {
		logs.Logger.Errorf("Get all metrics of SERVICE.BMS error: %s", err.Error())
		return filteredMetrics
	}
	if bmsInfo.LabelInfo == nil {
		logs.Logger.Info("No bms resource info found, skip to query agent metrics info")
		return filteredMetrics
	}
	bmsInfo.Lock()
	defer bmsInfo.Unlock()
	for _, metric := range allMetrics {
		serverKey := getServerResourceKeyFromMetricInfo(metric)
		if serverKey == "" {
			continue
		}
		if _, ok := bmsInfo.LabelInfo[serverKey]; !ok {
			continue
		}
		//白名单校验通过，查询当前指标对应指标数据
		if IsMetricInfoInWhiteList(metric) {
			filteredMetrics = append(filteredMetrics, metric)
		}
	}
	return filteredMetrics
}
