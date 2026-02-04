package collector

import (
	"fmt"
	"strings"
	"sync"
	"time"

	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	elb "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var elbInfo serversInfo
var listenersMap map[string]model.Listener
var poolsMap map[string]model.Pool
var availabilityZoneMap map[string]model.AvailabilityZone

func getELBClient() *elb.ElbClient {
	return elb.NewElbClient(elb.ElbClientBuilder().WithCredential(
		authCredentialMap[conf.AuthMode](RegionServiceType)).
		WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
		WithEndpoint(getEndpoint("elb", "v2")).Build())
}

func listLoadBalancers() ([]model.LoadBalancer, error) {
	limit := int32(1000)
	epIds := getEpIdRequestPart()
	request := &model.ListLoadBalancersRequest{Limit: &limit, EnterpriseProjectId: &epIds}
	var loadbalancers []model.LoadBalancer
	for {
		response, err := getELBClient().ListLoadBalancers(request)
		if err != nil {
			logs.Logger.Errorf("list LoadBalancers error: %s", err.Error())
			return loadbalancers, err
		}
		tempLoadbalancers := *response.Loadbalancers
		if len(tempLoadbalancers) == 0 {
			break
		}
		loadbalancers = append(loadbalancers, tempLoadbalancers...)
		request.Marker = &(tempLoadbalancers[len(tempLoadbalancers)-1].Id)
	}

	return loadbalancers, nil
}

func listListeners() []model.Listener {
	limit := int32(1000)
	epIds := getEpIdRequestPart()
	request := &model.ListListenersRequest{Limit: &limit, EnterpriseProjectId: &epIds}
	var listeners []model.Listener
	for {
		response, err := getELBClient().ListListeners(request)
		if err != nil {
			logs.Logger.Errorf("list Listeners error: %s", err.Error())
			return listeners
		}
		tempListeners := *response.Listeners
		if len(tempListeners) == 0 {
			break
		}
		listeners = append(listeners, tempListeners...)
		request.Marker = &(tempListeners[len(tempListeners)-1].Id)
	}

	return listeners
}

func listPools() []model.Pool {
	limit := int32(1000)
	epIds := getEpIdRequestPart()
	request := &model.ListPoolsRequest{Limit: &limit, EnterpriseProjectId: &epIds}
	var pools []model.Pool
	for {
		response, err := getELBClient().ListPools(request)
		if err != nil {
			logs.Logger.Errorf("list Pool error: %s", err.Error())
			return pools
		}
		tempPools := *response.Pools
		if len(tempPools) == 0 {
			break
		}
		pools = append(pools, tempPools...)
		request.Marker = &(tempPools[len(tempPools)-1].Id)
	}

	return pools
}

func listAvailabilityZones() []model.AvailabilityZone {
	request := &model.ListAvailabilityZonesRequest{}
	var resultZones []model.AvailabilityZone
	response, err := getELBClient().ListAvailabilityZones(request)
	if err != nil {
		logs.Logger.Errorf("list availability zones error: %s", err.Error())
		return resultZones
	}
	zones := *response.AvailabilityZones
	for _, subZones := range zones {
		resultZones = append(resultZones, subZones...)
	}

	return resultZones
}

type ELBInfo struct{}

func (getter ELBInfo) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]cesmodel.MetricInfoList, 0)
	elbInfo.Lock()
	defer elbInfo.Unlock()
	if elbInfo.LabelInfo == nil || time.Now().Unix() > elbInfo.TTL {
		getResourceMap()
		sysConfigMap := getMetricConfigMap("SYS.ELB")
		loadBalancers, err := listLoadBalancers()
		if err != nil {
			logs.Logger.Errorf("Get elb instances from services error: %s", err.Error())
			return elbInfo.LabelInfo, elbInfo.FilterMetrics
		}
		for index := range loadBalancers {
			loadBalancer := loadBalancers[index]
			if loadBalancerMetricNames, ok := sysConfigMap["lbaas_instance_id"]; ok {
				metrics := buildSingleDimensionMetrics(loadBalancerMetricNames, "SYS.ELB", "lbaas_instance_id", loadBalancer.Id)
				filterMetrics = append(filterMetrics, metrics...)
				info := labelInfo{
					Name:  []string{"name", "epId", "vip_address", "provider"},
					Value: []string{loadBalancer.Name, loadBalancer.EnterpriseProjectId, loadBalancer.VipAddress, loadBalancer.Provider},
				}
				keys, values := getElbTags(loadBalancer.Tags)
				info.Name = append(info.Name, keys...)
				info.Value = append(info.Value, values...)
				commonInfo := labelInfo{
					Name:  []string{},
					Value: []string{},
				}
				commonInfo.Name = append(commonInfo.Name, info.Name...)
				commonInfo.Value = append(commonInfo.Value, info.Value...)
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = commonInfo

				buildListenerInfo(sysConfigMap, &loadBalancer, commonInfo, &filterMetrics, resourceInfos)

				buildPoolInfo(sysConfigMap, &loadBalancer, commonInfo, &filterMetrics, resourceInfos)

				buildAvailabilityZoneInfo(sysConfigMap, &loadBalancer, commonInfo, &filterMetrics, resourceInfos)

				buildIpAddressInfo(sysConfigMap, &loadBalancer, commonInfo, &filterMetrics, resourceInfos)
			}
		}
		elbInfo.LabelInfo = resourceInfos
		elbInfo.FilterMetrics = filterMetrics
		elbInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return elbInfo.LabelInfo, elbInfo.FilterMetrics
}

func buildListenerInfo(sysConfigMap map[string][]string, loadBalancer *model.LoadBalancer, info labelInfo, filterMetrics *[]cesmodel.MetricInfoList, resourceInfos map[string]labelInfo) {
	listenerMetricNames, ok := sysConfigMap["lbaas_instance_id,lbaas_listener_id"]
	if !ok {
		logs.Logger.Warn("listener metric names not config")
		return
	}
	for _, listener := range loadBalancer.Listeners {
		listenersMetrics := buildDimensionMetrics(listenerMetricNames, "SYS.ELB",
			[]cesmodel.MetricsDimension{{Name: "lbaas_instance_id", Value: loadBalancer.Id}, {Name: "lbaas_listener_id", Value: listener.Id}})
		listenerInfo := info
		if detailListener, ok := listenersMap[listener.Id]; ok {
			listenerInfo.Name = append(listenerInfo.Name, []string{"listener_name", "port", "protocol"}...)
			listenerInfo.Value = append(listenerInfo.Value, []string{detailListener.Name, fmt.Sprintf("%d", detailListener.ProtocolPort), detailListener.Protocol}...)

			keys, values := getElbTags(detailListener.Tags)
			listenerInfo.Name = append(listenerInfo.Name, keys...)
			listenerInfo.Value = append(listenerInfo.Value, values...)
		}
		*filterMetrics = append(*filterMetrics, listenersMetrics...)
		resourceInfos[GetResourceKeyFromMetricInfo(listenersMetrics[0])] = listenerInfo
	}
}

func buildPoolInfo(sysConfigMap map[string][]string, loadBalancer *model.LoadBalancer, info labelInfo, filterMetrics *[]cesmodel.MetricInfoList, resourceInfos map[string]labelInfo) {
	poolMetricNames, ok := sysConfigMap["lbaas_instance_id,lbaas_pool_id"]
	if !ok {
		logs.Logger.Warn("pool metric names not config")
		return
	}
	for _, pool := range loadBalancer.Pools {
		poolsMetrics := buildDimensionMetrics(poolMetricNames, "SYS.ELB",
			[]cesmodel.MetricsDimension{{Name: "lbaas_instance_id", Value: loadBalancer.Id}, {Name: "lbaas_pool_id", Value: pool.Id}})
		poolInfo := info
		if detailPool, ok := poolsMap[pool.Id]; ok {
			poolInfo.Name = append(poolInfo.Name, []string{"pool_name", "pool_protocol"}...)
			poolInfo.Value = append(poolInfo.Value, []string{detailPool.Name, detailPool.Protocol}...)
		}
		*filterMetrics = append(*filterMetrics, poolsMetrics...)
		resourceInfos[GetResourceKeyFromMetricInfo(poolsMetrics[0])] = poolInfo
	}
}

func buildAvailabilityZoneInfo(sysConfigMap map[string][]string, loadBalancer *model.LoadBalancer, info labelInfo, filterMetrics *[]cesmodel.MetricInfoList, resourceInfos map[string]labelInfo) {
	zoneMetricNames, ok := sysConfigMap["lbaas_instance_id,available_zone"]
	if !ok {
		logs.Logger.Warn("availability zone metric names not config")
		return
	}
	for _, zone := range loadBalancer.AvailabilityZoneList {
		zonesMetrics := buildDimensionMetrics(zoneMetricNames, "SYS.ELB",
			[]cesmodel.MetricsDimension{{Name: "lbaas_instance_id", Value: loadBalancer.Id}, {Name: "available_zone", Value: zone}})
		zoneInfo := info
		if detailZone, ok := availabilityZoneMap[zone]; ok {
			zoneInfo.Name = append(zoneInfo.Name, []string{"state", "public_border_group", "category", "protocol"}...)
			zoneInfo.Value = append(zoneInfo.Value, []string{detailZone.State, detailZone.PublicBorderGroup, fmt.Sprintf("%d", detailZone.Category), strings.Join(detailZone.Protocol, ",")}...)
		}
		*filterMetrics = append(*filterMetrics, zonesMetrics...)
		resourceInfos[GetResourceKeyFromMetricInfo(zonesMetrics[0])] = zoneInfo
	}
}

func buildIpAddressInfo(sysConfigMap map[string][]string, loadBalancer *model.LoadBalancer, info labelInfo, filterMetrics *[]cesmodel.MetricInfoList, resourceInfos map[string]labelInfo) {
	metricNames, ok := sysConfigMap["lbaas_instance_id,ip_address"]
	if !ok {
		logs.Logger.Warn("ip address metric names not config")
		return
	}
	var ipAddresses []string
	if loadBalancer.VipAddress != "" {
		ipAddresses = append(ipAddresses, loadBalancer.VipAddress)
	}
	if loadBalancer.Ipv6VipAddress != "" {
		// 指标上报时Ipv6地址被进行了符号转换
		ipv6VipAddress := strings.ReplaceAll(loadBalancer.Ipv6VipAddress, ":", "#")
		ipAddresses = append(ipAddresses, ipv6VipAddress)
	}
	for _, publicIp := range loadBalancer.Publicips {
		ipAddresses = append(ipAddresses, publicIp.PublicipAddress)
	}
	for _, ipAddress := range ipAddresses {
		metrics := buildDimensionMetrics(metricNames, "SYS.ELB",
			[]cesmodel.MetricsDimension{{Name: "lbaas_instance_id", Value: loadBalancer.Id}, {Name: "ip_address", Value: ipAddress}})
		*filterMetrics = append(*filterMetrics, metrics...)
		ipAddressInfo := labelInfo{
			Name:  []string{},
			Value: []string{},
		}
		ipAddressInfo.Name = append(ipAddressInfo.Name, info.Name...)
		ipAddressInfo.Value = append(ipAddressInfo.Value, info.Value...)
		resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = ipAddressInfo
	}
}

func getResourceMap() {
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		listenersMap = getListenersInfoMap()
		wg.Done()
	}()
	go func() {
		poolsMap = getPoolsInfoMap()
		wg.Done()
	}()
	go func() {
		availabilityZoneMap = getAvailabilityZoneInfoMap()
		wg.Done()
	}()
	wg.Wait()
}

func getListenersInfoMap() map[string]model.Listener {
	listeners := listListeners()
	listenersMap := make(map[string]model.Listener, len(listeners))
	for _, listener := range listeners {
		listenersMap[listener.Id] = listener
	}
	return listenersMap
}

func getPoolsInfoMap() map[string]model.Pool {
	pools := listPools()
	poolsMap := make(map[string]model.Pool, len(pools))
	for _, pool := range pools {
		poolsMap[pool.Id] = pool
	}
	return poolsMap
}

func getAvailabilityZoneInfoMap() map[string]model.AvailabilityZone {
	zones := listAvailabilityZones()
	zonesMap := make(map[string]model.AvailabilityZone, len(zones))
	for _, zone := range zones {
		zonesMap[zone.Code] = zone
	}
	return zonesMap
}

// 标签只允许大写字母，小写字母和下划线，过滤tags中有效的tag
func getElbTags(tags []model.Tag) ([]string, []string) {
	var keys, values []string
	for _, tag := range tags {
		valid := tagRegexp.MatchString(*tag.Key)
		if !valid {
			continue
		}
		keys = append(keys, *tag.Key)
		values = append(values, *tag.Value)
	}
	return keys, values
}
