package collector

import (
	"time"

	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

var hssInfo serversInfo

type HSSInfo struct{}

func (getter HSSInfo) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]cesmodel.MetricInfoList, 0)
	if hssInfo.LabelInfo == nil || time.Now().Unix() > hssInfo.TTL {
		enterpriseProjects, err := listEps()
		if err != nil {
			return hssInfo.LabelInfo, hssInfo.FilterMetrics
		}
		sysConfigMap := getMetricConfigMap("SYS.HSS")
		if nil == sysConfigMap {
			return hssInfo.LabelInfo, hssInfo.FilterMetrics
		}
		hssEpIdMetricNames, ok := sysConfigMap["hss_enterprise_project_id"]
		if !ok {
			return hssInfo.LabelInfo, hssInfo.FilterMetrics
		}
		for _, enterpriseProject := range enterpriseProjects {
			metrics := buildSingleDimensionMetrics(hssEpIdMetricNames, "SYS.HSS", "hss_enterprise_project_id", enterpriseProject.Id)
			filterMetrics = append(filterMetrics, metrics...)
			info := labelInfo{
				Name:  []string{"epId"},
				Value: []string{enterpriseProject.Id},
			}
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
		}
		hssInfo.LabelInfo = resourceInfos
		hssInfo.FilterMetrics = filterMetrics
		hssInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return hssInfo.LabelInfo, hssInfo.FilterMetrics
}
