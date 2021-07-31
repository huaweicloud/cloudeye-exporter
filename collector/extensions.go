package collector

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/huaweicloud/golangsdk/openstack/ces/v1/metrics"
	"github.com/huaweicloud/golangsdk/openstack/networking/v2/extensions/lbaas_v2/loadbalancers"
)

// If the extension labels have to added in this exporter, you only have
// to add the code to the following two parts.
// 1. Added the new labels name to defaultExtensionLabels
// 2. Added the new labels values to getAllResource
var defaultExtensionLabels = map[string][]string{
	"sys_elb":                 []string{"name", "provider", "vip_address"},
	"sys_elb_listener":        []string{"name", "port"},
	"sys_nat":                 []string{"name"},
	"sys_rds":                 []string{"name"},
	"sys_rds_instance":        []string{"port", "name", "role"},
	"sys_dcs":                 []string{"ip", "port", "name", "engine"},
	"sys_dms":                 []string{"name"},
	"sys_dms_instance":        []string{"name", "engine_version", "resource_spec_code", "connect_address", "port"},
	"sys_dms_instance_broker": []string{"name", "engine_version", "resource_spec_code", "connect_address", "port"},
	"sys_dms_instance_topics": []string{"name", "engine_version", "resource_spec_code", "connect_address", "port"},
	"sys_vpc_bandwidth":       []string{"name", "size", "share_type", "bandwidth_type", "charge_mode"},
	"sys_vpc_eip":             []string{"name", "public_ip_address", "type"},
	"sys_evs":                 []string{"name", "server_id", "device"},
	"sys_ecs":                 []string{"hostname"},
	"sys_as":                  []string{"name", "status"},
	"sys_functiongraph":       []string{"func_urn"},
}

const TTL = time.Hour * 3

var (
	elbInfo serversInfo
	natInfo serversInfo
	rdsInfo serversInfo
	dmsInfo serversInfo
	dcsInfo serversInfo
	vpcInfo serversInfo
	evsInfo serversInfo
	ecsInfo serversInfo
	asInfo  serversInfo
	fgsInfo serversInfo
)

type serversInfo struct {
	TTL           int64
	LenMetric     int
	Info          map[string][]string
	FilterMetrics []metrics.Metric
	sync.Mutex
}

func buildSingleDimensionMetrics(metricNames []string, namespace, dimName, dimValue string) []metrics.Metric {
	filterMetrics := make([]metrics.Metric, 0)
	for index := range metricNames {
		filterMetrics = append(filterMetrics, metrics.Metric{
			Namespace:  namespace,
			MetricName: metricNames[index],
			Dimensions: []metrics.Dimension{
				{
					Name:  dimName,
					Value: dimValue,
				},
			},
		})
	}
	return filterMetrics
}

func (exporter *BaseHuaweiCloudExporter) getElbResourceInfo() (map[string][]string, *[]metrics.Metric) {
	resourceInfos := make(map[string][]string)
	filterMetrics := make([]metrics.Metric, 0)
	elbInfo.Lock()
	defer elbInfo.Unlock()
	if elbInfo.Info == nil || time.Now().Unix() > elbInfo.TTL {
		allELBs, err := getAllLoadBalancer(exporter.ClientConfig)
		if err != nil {
			logs.Logger.Errorln("Get all LoadBalancer error:", err.Error())
			return elbInfo.Info, &elbInfo.FilterMetrics
		}
		if allELBs == nil {
			return elbInfo.Info, &elbInfo.FilterMetrics
		}
		configMap := getMetricConfigMap("SYS.ELB")
		for _, elb := range *allELBs {
			resourceInfos[elb.ID] = []string{elb.Name, elb.Provider, elb.VipAddress}
			if configMap == nil {
				continue
			}
			if metricNames, ok := configMap["lbaas_instance_id"]; ok {
				filterMetrics = append(filterMetrics, buildSingleDimensionMetrics(metricNames, "SYS.ELB", "lbaas_instance_id", elb.ID)...)
			}
			if metricNames, ok := configMap["lbaas_instance_id,lbaas_listener_id"]; ok {
				filterMetrics = append(filterMetrics, buildListenerMetrics(metricNames, &elb)...)
			}
			if metricNames, ok := configMap["lbaas_instance_id,lbaas_pool_id"]; ok {
				filterMetrics = append(filterMetrics, buildPoolMetrics(metricNames, &elb)...)
			}
		}

		allListeners, err := getAllListener(exporter.ClientConfig)
		if err != nil {
			logs.Logger.Errorln("Get all Listener error:", err.Error())
		}
		if allListeners != nil {
			for _, listener := range *allListeners {
				resourceInfos[listener.ID] = []string{listener.Name, fmt.Sprintf("%d", listener.ProtocolPort)}
			}
		}

		elbInfo.Info = resourceInfos
		elbInfo.FilterMetrics = filterMetrics
		elbInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return elbInfo.Info, &elbInfo.FilterMetrics
}

func buildListenerMetrics(metricNames []string, elb *loadbalancers.LoadBalancer) []metrics.Metric {
	filterMetrics := make([]metrics.Metric, 0)
	for listenerIndex := range elb.Listeners {
		for index := range metricNames {
			filterMetrics = append(filterMetrics, metrics.Metric{
				Namespace:  "SYS.ELB",
				MetricName: metricNames[index],
				Dimensions: []metrics.Dimension{
					{
						Name:  "lbaas_instance_id",
						Value: elb.ID,
					},
					{
						Name:  "lbaas_listener_id",
						Value: elb.Listeners[listenerIndex].ID,
					},
				},
			})
		}
	}
	return filterMetrics
}

func buildPoolMetrics(metricNames []string, elb *loadbalancers.LoadBalancer) []metrics.Metric {
	filterMetrics := make([]metrics.Metric, 0)
	for poolIndex := range elb.Pools {
		for index := range metricNames {
			filterMetrics = append(filterMetrics, metrics.Metric{
				Namespace:  "SYS.ELB",
				MetricName: metricNames[index],
				Dimensions: []metrics.Dimension{
					{
						Name:  "lbaas_instance_id",
						Value: elb.ID,
					},
					{
						Name:  "lbaas_pool_id",
						Value: elb.Pools[poolIndex].ID,
					},
				},
			})
		}
	}
	return filterMetrics
}

func (exporter *BaseHuaweiCloudExporter) getNatResourceInfo() (map[string][]string, *[]metrics.Metric) {
	resourceInfos := make(map[string][]string)
	filterMetrics := make([]metrics.Metric, 0)
	natInfo.Lock()
	defer natInfo.Unlock()
	if natInfo.Info == nil || time.Now().Unix() > natInfo.TTL {
		allnat, err := getAllNat(exporter.ClientConfig)
		if err != nil {
			logs.Logger.Errorln("Get all Nat error:", err.Error())
			return natInfo.Info, &natInfo.FilterMetrics
		}
		if allnat == nil {
			return natInfo.Info, &natInfo.FilterMetrics
		}
		configMap := getMetricConfigMap("SYS.NAT")
		for _, nat := range *allnat {
			resourceInfos[nat.ID] = []string{nat.Name}
			if configMap == nil {
				continue
			}
			if metricNames, ok := configMap["nat_gateway_id"]; ok {
				filterMetrics = append(filterMetrics, buildSingleDimensionMetrics(metricNames, "SYS.NAT", "nat_gateway_id", nat.ID)...)
			}
		}

		natInfo.Info = resourceInfos
		natInfo.FilterMetrics = filterMetrics
		natInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return natInfo.Info, &natInfo.FilterMetrics
}

func (exporter *BaseHuaweiCloudExporter) getRdsResourceInfo() (map[string][]string, *[]metrics.Metric) {
	resourceInfos := make(map[string][]string)
	filterMetrics := make([]metrics.Metric, 0)
	rdsInfo.Lock()
	defer rdsInfo.Unlock()
	if rdsInfo.Info == nil || time.Now().Unix() > rdsInfo.TTL {
		allrds, err := getAllRds(exporter.ClientConfig)
		if err != nil {
			logs.Logger.Errorln("Get all Rds error:", err.Error())
			return rdsInfo.Info, &rdsInfo.FilterMetrics
		}
		if allrds == nil {
			return rdsInfo.Info, &rdsInfo.FilterMetrics
		}
		configMap := getMetricConfigMap("SYS.RDS")
		for _, rds := range allrds.Instances {
			resourceInfos[rds.Id] = []string{rds.Name}
			for _, node := range rds.Nodes {
				resourceInfos[node.Id] = []string{fmt.Sprintf("%d", rds.Port), node.Name, node.Role}
			}
			if configMap == nil {
				continue
			}
			var dimName string
			switch rds.DataStore.Type {
			case "MySQL":
				dimName = "rds_cluster_id"
			case "PostgreSQL":
				dimName = "postgresql_cluster_id"
			case "SQLServer":
				dimName = "rds_cluster_sqlserver_id"
			}
			if metricNames, ok := configMap[dimName]; ok {
				filterMetrics = append(filterMetrics, buildSingleDimensionMetrics(metricNames, "SYS.RDS", dimName, rds.Id)...)
			}
		}

		rdsInfo.Info = resourceInfos
		rdsInfo.FilterMetrics = filterMetrics
		rdsInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return rdsInfo.Info, &rdsInfo.FilterMetrics
}

func (exporter *BaseHuaweiCloudExporter) getDmsResourceInfo() (map[string][]string, *[]metrics.Metric) {
	resourceInfos := map[string][]string{}
	dmsInfo.Lock()
	defer dmsInfo.Unlock()
	if dmsInfo.Info == nil || time.Now().Unix() > dmsInfo.TTL {
		allDmsInstance, err := getAllDms(exporter.ClientConfig)
		if err != nil {
			logs.Logger.Errorln("Get all Dms error:", err.Error())
			return dmsInfo.Info, &dmsInfo.FilterMetrics
		}
		if allDmsInstance == nil {
			return dmsInfo.Info, &dmsInfo.FilterMetrics
		}

		for _, dms := range allDmsInstance.Instances {
			resourceInfos[dms.InstanceID] = []string{dms.Name, dms.EngineVersion, dms.ResourceSpecCode, dms.ConnectAddress,
				fmt.Sprintf("%d", dms.Port)}
		}

		allQueues, err := getAllDmsQueue(exporter.ClientConfig)
		if err != nil {
			logs.Logger.Errorln("Get all Dms Queue error:", err.Error())
		}
		if allQueues != nil {
			for _, queue := range *allQueues {
				resourceInfos[queue.ID] = []string{queue.Name}
			}
		}

		dmsInfo.Info = resourceInfos
		dmsInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return dmsInfo.Info, &dmsInfo.FilterMetrics
}

func (exporter *BaseHuaweiCloudExporter) getDcsResourceInfo() (map[string][]string, *[]metrics.Metric) {
	resourceInfos := make(map[string][]string)
	filterMetrics := make([]metrics.Metric, 0)
	dcsInfo.Lock()
	defer dcsInfo.Unlock()
	if dcsInfo.Info == nil || time.Now().Unix() > dcsInfo.TTL {
		allDcs, err := getAllDcs(exporter.ClientConfig)
		if err != nil {
			logs.Logger.Errorln("Get all Dcs error:", err.Error())
			return dcsInfo.Info, &dcsInfo.FilterMetrics
		}
		if allDcs == nil {
			return dcsInfo.Info, &dcsInfo.FilterMetrics
		}
		configMap := getMetricConfigMap("SYS.DCS")
		for _, dcs := range allDcs.Instances {
			resourceInfos[dcs.InstanceID] = []string{dcs.IP, fmt.Sprintf("%d", dcs.Port), dcs.Name, dcs.Engine}
			if configMap == nil {
				continue
			}
			var dimName string
			switch dcs.Engine {
			case "Redis":
				dimName = "dcs_instance_id"
			case "Memcached":
				dimName = "dcs_memcached_instance_id"
			}
			if metricNames, ok := configMap[dimName]; ok {
				filterMetrics = append(filterMetrics, buildSingleDimensionMetrics(metricNames, "SYS.DCS", dimName, dcs.InstanceID)...)
			}
		}

		dcsInfo.Info = resourceInfos
		dcsInfo.FilterMetrics = filterMetrics
		dcsInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return dcsInfo.Info, &dcsInfo.FilterMetrics
}

func (exporter *BaseHuaweiCloudExporter) getVpcResourceInfo() (map[string][]string, *[]metrics.Metric) {
	resourceInfos := map[string][]string{}
	vpcInfo.Lock()
	defer vpcInfo.Unlock()
	if vpcInfo.Info == nil || time.Now().Unix() > vpcInfo.TTL {
		allPublicIps, err := getAllPublicIp(exporter.ClientConfig)
		if err != nil {
			logs.Logger.Errorln("Get all PublicIp error:", err.Error())
		}
		if allPublicIps != nil {
			for _, publicIp := range *allPublicIps {
				resourceInfos[publicIp.ID] = []string{publicIp.BandwidthName, publicIp.PublicIpAddress, publicIp.Type}
			}
		}

		allBandwidth, err := getAllBandwidth(exporter.ClientConfig)
		if err != nil {
			logs.Logger.Errorln("Get all Bandwidth error:", err.Error())
			return resourceInfos, &vpcInfo.FilterMetrics
		}
		if allBandwidth != nil {
			for _, bandwidth := range *allBandwidth {
				resourceInfos[bandwidth.ID] = []string{bandwidth.Name, fmt.Sprintf("%d", bandwidth.Size), bandwidth.ShareType, bandwidth.BandwidthType, bandwidth.ChargeMode}
			}
		}

		vpcInfo.Info = resourceInfos
		vpcInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return vpcInfo.Info, &vpcInfo.FilterMetrics
}

func (exporter *BaseHuaweiCloudExporter) getEvsResourceInfo() (map[string][]string, *[]metrics.Metric) {
	resourceInfos := map[string][]string{}
	evsInfo.Lock()
	defer evsInfo.Unlock()
	if evsInfo.Info == nil || time.Now().Unix() > evsInfo.TTL {
		allVolumes, err := getAllVolume(exporter.ClientConfig)
		if err != nil {
			logs.Logger.Errorf("Get all Volume error: %s", err.Error())
			return evsInfo.Info, &evsInfo.FilterMetrics
		}
		if allVolumes == nil {
			return evsInfo.Info, &evsInfo.FilterMetrics
		}

		for _, volume := range *allVolumes {
			if len(volume.Attachments) > 0 {
				device := strings.Split(volume.Attachments[0].Device, "/")
				resourceInfos[fmt.Sprintf("%s-%s", volume.Attachments[0].ServerID, device[len(device)-1])] = []string{volume.Name, volume.Attachments[0].ServerID, volume.Attachments[0].Device}
			}
		}

		evsInfo.Info = resourceInfos
		evsInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return evsInfo.Info, &evsInfo.FilterMetrics
}

func (exporter *BaseHuaweiCloudExporter) getEcsResourceInfo() (map[string][]string, *[]metrics.Metric) {
	resourceInfos := map[string][]string{}
	ecsInfo.Lock()
	defer ecsInfo.Unlock()
	if ecsInfo.Info == nil || time.Now().Unix() > ecsInfo.TTL {
		allServers, err := getAllServer(exporter.ClientConfig)
		if err != nil {
			logs.Logger.Errorln("Get all Server error:", err.Error())
			return ecsInfo.Info, &ecsInfo.FilterMetrics
		}
		if allServers == nil {
			return ecsInfo.Info, &ecsInfo.FilterMetrics
		}

		for _, server := range *allServers {
			resourceInfos[server.ID] = []string{server.Name}
		}

		ecsInfo.Info = resourceInfos
		ecsInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return ecsInfo.Info, &ecsInfo.FilterMetrics
}

func (exporter *BaseHuaweiCloudExporter) getAsResourceInfo() (map[string][]string, *[]metrics.Metric) {
	resourceInfos := map[string][]string{}
	asInfo.Lock()
	defer asInfo.Unlock()
	if asInfo.Info == nil || time.Now().Unix() > asInfo.TTL {
		allGroups, err := getAllGroup(exporter.ClientConfig)
		if err != nil {
			logs.Logger.Errorln("Get all Group error:", err.Error())
			return asInfo.Info, &asInfo.FilterMetrics
		}
		if allGroups == nil {
			return asInfo.Info, &asInfo.FilterMetrics
		}

		for _, group := range *allGroups {
			resourceInfos[group.ID] = []string{group.Name, group.Status}
		}

		asInfo.Info = resourceInfos
		asInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return asInfo.Info, &asInfo.FilterMetrics
}

func (exporter *BaseHuaweiCloudExporter) getFunctionGraphResourceInfo() (map[string][]string, *[]metrics.Metric) {
	resourceInfos := map[string][]string{}
	fgsInfo.Lock()
	defer fgsInfo.Unlock()
	if fgsInfo.Info == nil || time.Now().Unix() > fgsInfo.TTL {
		functionList, err := getAllFunction(exporter.ClientConfig)
		if err != nil {
			logs.Logger.Errorln("Get all Function error:", err.Error())
			return fgsInfo.Info, &fgsInfo.FilterMetrics
		}
		if functionList == nil {
			return fgsInfo.Info, &fgsInfo.FilterMetrics
		}

		for _, function := range functionList.Functions {
			resourceInfos[fmt.Sprintf("%s-%s", function.Package, function.FuncName)] = []string{function.FuncUrn}
		}

		fgsInfo.Info = resourceInfos
		fgsInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return fgsInfo.Info, &fgsInfo.FilterMetrics
}

func (exporter *BaseHuaweiCloudExporter) getAllResource(namespace string) (map[string][]string, *[]metrics.Metric) {
	switch namespace {
	case "SYS.ELB":
		return exporter.getElbResourceInfo()
	case "SYS.NAT":
		return exporter.getNatResourceInfo()
	case "SYS.RDS":
		return exporter.getRdsResourceInfo()
	case "SYS.DMS":
		return exporter.getDmsResourceInfo()
	case "SYS.DCS":
		return exporter.getDcsResourceInfo()
	case "SYS.VPC":
		return exporter.getVpcResourceInfo()
	case "SYS.EVS":
		return exporter.getEvsResourceInfo()
	case "SYS.ECS":
		return exporter.getEcsResourceInfo()
	case "SYS.AS":
		return exporter.getAsResourceInfo()
	case "SYS.FunctionGraph":
		return exporter.getFunctionGraphResourceInfo()
	default:
		return map[string][]string{}, &[]metrics.Metric{}
	}
}
