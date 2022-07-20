package collector

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

func buildSingleDimensionMetrics(metricNames []string, namespace, dimName, dimValue string) []model.MetricInfoList {
	filterMetrics := make([]model.MetricInfoList, len(metricNames))
	for index := range metricNames {
		filterMetrics[index] = model.MetricInfoList{
			Namespace:  namespace,
			MetricName: metricNames[index],
			Dimensions: []model.MetricsDimension{
				{
					Name:  dimName,
					Value: dimValue,
				},
			},
		}
	}
	return filterMetrics
}

func buildDimensionMetrics(metricNames []string, namespace string, dimensions []model.MetricsDimension) []model.MetricInfoList {
	filterMetrics := make([]model.MetricInfoList, len(metricNames))
	for index := range metricNames {
		filterMetrics[index] = model.MetricInfoList{
			Namespace:  namespace,
			MetricName: metricNames[index],
			Dimensions: dimensions,
		}
	}
	return filterMetrics
}

func (exporter *BaseHuaweiCloudExporter) listAllResources(namespace string) (map[string]labelInfo, []model.MetricInfoList) {
	switch namespace {
	case "SYS.ECS":
		return exporter.getEcsResourceInfo()
	case "AGT.ECS":
		return exporter.getAGTResourceInfo()
	case "SYS.EVS":
		return exporter.getEvsResourceInfo()
	case "SYS.DCS":
		return exporter.getDcsResourceInfo()
	case "SYS.DCAAS":
		return exporter.getDcaasResourceInfo()
	case "SYS.VPC":
		return exporter.getVpcResourceInfo()
	case "SYS.ES":
		return exporter.getEsResourceInfo()
	case "SYS.RDS":
		return exporter.getRdsResourceInfo()
	case "SYS.ELB":
		return exporter.getElbResourceInfo()
	case "SYS.GAUSSDB":
		return exporter.getGaussdbResourceInfo()
	case "SYS.GAUSSDBV5":
		return exporter.getGaussdbV5ResourceInfo()
	case "SYS.NAT":
		return exporter.getNATResourceInfo()
	case "SYS.AS":
		return exporter.getASResourceInfo()
	case "SYS.FunctionGraph":
		return exporter.getFunctionGraphResourceInfo()
	case "SYS.DRS":
		return exporter.getDrsResourceInfo()
	default:
		return map[string]labelInfo{}, []model.MetricInfoList{}
	}
}
