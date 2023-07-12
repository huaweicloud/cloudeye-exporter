package collector

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

type ResourceInfoGetter interface {
	GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList)
}

var (
	serviceMap = map[string]ResourceInfoGetter{
		"SYS.ECS":           ECSInfo{},
		"AGT.ECS":           AGTECSInfo{},
		"SYS.EVS":           EVSInfo{},
		"SYS.DCS":           DCSInfo{},
		"SYS.DCAAS":         DCAASInfo{},
		"SYS.VPC":           VPCInfo{},
		"SYS.ES":            ESInfo{},
		"SYS.RDS":           RDSInfo{},
		"SYS.ELB":           ELBInfo{},
		"SYS.GAUSSDB":       GAUSSDBInfo{},
		"SYS.GAUSSDBV5":     GAUSSDBV5Info{},
		"SYS.NAT":           NATInfo{},
		"SYS.AS":            ASInfo{},
		"SYS.FunctionGraph": FunctionGraphInfo{},
		"SYS.DRS":           DRSInfo{},
		"SYS.WAF":           WAFInfo{},
		"SYS.DDS":           DDSInfo{},
		"SYS.APIG":          APIGInfo{},
		"SYS.CBR":           CBRInfo{},
		"SYS.DLI":           DLIInfo{},
		"SYS.SFS":           SFSInfo{},
		"SYS.EFS":           EFSInfo{},
		"SYS.VPN":           VPNInfo{},
		"SYS.CDM":           CDMInfo{},
		"SYS.DWS":           DWSInfo{},
		"SYS.DDOS":          DDOSInfo{},
		"SYS.NoSQL":         NoSQLInfo{},
		"SYS.DMS":           DMSInfo{},
		"SYS.DDMS":          DDMSInfo{},
		"SYS.APIC":          APICInfo{},
		"SYS.BMS":           BMSInfo{},
		"SERVICE.BMS":       SERVICEBMSInfo{},
		"SYS.VPCEP":         VPCEPInfo{},
		"SYS.ModelArts":     ModelArtsInfo{},
		"SYS.GES":           GESInfo{},
		"SYS.DBSS":          DBSSInfo{},
		"SYS.CC":            CCInfo{},
		"SYS.LakeFormation": LakeFormationInfo{},
		"SYS.MRS":           MRSInfo{},
		"SYS.DAYU":          DayuInfo{},
	}
)

func (exporter *BaseHuaweiCloudExporter) listAllResources(namespace string) (map[string]labelInfo, []model.MetricInfoList) {
	serviceFunc, ok := serviceMap[namespace]
	if !ok {
		return map[string]labelInfo{}, []model.MetricInfoList{}
	}
	return serviceFunc.GetResourceInfo()
}
