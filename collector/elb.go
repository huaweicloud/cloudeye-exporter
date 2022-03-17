package collector

import (
	"fmt"
	"sync"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	elb "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var elbInfo serversInfo
var listenersMap map[string]model.Listener
var poolsMap map[string]model.Pool

func getELBClient() *elb.ElbClient {
	return elb.NewElbClient(elb.ElbClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(true)).
		WithEndpoint(getEndpoint("elb", "v2")).Build())
}

func listLoadBalancers() []model.LoadBalancer {
	limit := int32(1000)
	request := &model.ListLoadBalancersRequest{Limit: &limit}
	var loadbalancers []model.LoadBalancer
	for {
		response, err := getELBClient().ListLoadBalancers(request)
		if err != nil {
			logs.Logger.Errorf("list LoadBalancers error: %s", err.Error())
			return loadbalancers
		}
		tempLoadbalancers := *response.Loadbalancers
		if len(tempLoadbalancers) == 0 {
			break
		}
		loadbalancers = append(loadbalancers, tempLoadbalancers...)
		request.Marker = &(tempLoadbalancers[len(tempLoadbalancers)-1].Id)
	}

	return loadbalancers
}

func listListeners() []model.Listener {
	limit := int32(1000)
	request := &model.ListListenersRequest{Limit: &limit}
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
	request := &model.ListPoolsRequest{Limit: &limit}
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

func (exporter *BaseHuaweiCloudExporter) getElbResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]cesmodel.MetricInfoList, 0)
	elbInfo.Lock()
	defer elbInfo.Unlock()
	if elbInfo.LabelInfo == nil || time.Now().Unix() > elbInfo.TTL {
		getResourceMap()
		sysConfigMap := getMetricConfigMap("SYS.ELB")
		for _, loadBalancer := range listLoadBalancers() {
			if loadBalancerMetricNames, ok := sysConfigMap["lbaas_instance_id"]; ok {
				metrics := buildSingleDimensionMetrics(loadBalancerMetricNames, "SYS.ELB", "lbaas_instance_id", loadBalancer.Id)
				filterMetrics = append(filterMetrics, metrics...)
				info := labelInfo{
					Name:  []string{"name", "epId", "vip_address", "provider"},
					Value: []string{loadBalancer.Name, *loadBalancer.EnterpriseProjectId, loadBalancer.VipAddress, loadBalancer.Provider},
				}
				keys, values := getElbTags(loadBalancer.Tags)
				info.Name = append(info.Name, keys...)
				info.Value = append(info.Value, values...)
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info

				buildListenerInfo(sysConfigMap, &loadBalancer, info, &filterMetrics, resourceInfos)

				buildPoolInfo(sysConfigMap, &loadBalancer, info, &filterMetrics, resourceInfos)
			}
		}

		elbInfo.LabelInfo = resourceInfos
		elbInfo.FilterMetrics = filterMetrics
		elbInfo.TTL = time.Now().Add(TTL).Unix()
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

func getResourceMap() {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		listenersMap = getListenersInfoMap()
		wg.Done()
	}()
	go func() {
		poolsMap = getPoolsInfoMap()
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
