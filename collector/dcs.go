package collector

import (
	"fmt"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

type DcsInstancesInfo struct {
	ResourceBaseInfo
	RmsDcsInstanceProperties
}

var dcsInfo serversInfo

type DCSInfo struct{}

func (getter DCSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	dcsInfo.Lock()
	defer dcsInfo.Unlock()
	if dcsInfo.LabelInfo == nil || time.Now().Unix() > dcsInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.DCS")

		// redis instance
		buildRedisInstancesInfo(sysConfigMap, &filterMetrics, resourceInfos)

		// memcached instance
		buildMemcachedInstancesInfo(sysConfigMap, &filterMetrics, resourceInfos)

		// node
		buildDcsNodesInfo(sysConfigMap, &filterMetrics, resourceInfos)

		dcsInfo.LabelInfo = resourceInfos
		dcsInfo.FilterMetrics = filterMetrics
		dcsInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return dcsInfo.LabelInfo, dcsInfo.FilterMetrics
}

func buildRedisInstancesInfo(sysConfigMap map[string][]string, filterMetrics *[]model.MetricInfoList, resourceInfos map[string]labelInfo) {
	redisInstances, err := getRedisInstancesFromRMS()
	if err != nil {
		return
	}
	for index := range redisInstances {
		if metricNames, ok := sysConfigMap["dcs_instance_id"]; ok {
			metrics := buildSingleDimensionMetrics(metricNames, "SYS.DCS", "dcs_instance_id", redisInstances[index].ID)
			*filterMetrics = append(*filterMetrics, metrics...)
			info := labelInfo{
				Name: []string{"name", "epId", "engine", "ip", "port", "cache_mode", "engine_version"},
				Value: []string{redisInstances[index].Name, redisInstances[index].EpId, redisInstances[index].Engine,
					redisInstances[index].IP, fmt.Sprintf("%d", redisInstances[index].Port), redisInstances[index].CacheMode, redisInstances[index].EngineVersion},
			}
			keys, values := getTags(redisInstances[index].Tags)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
		}
	}
}

func buildMemcachedInstancesInfo(sysConfigMap map[string][]string, filterMetrics *[]model.MetricInfoList, resourceInfos map[string]labelInfo) {
	memcachedInstances, err := getMemcachedInstancesFromRMS()
	if err != nil {
		return
	}
	for index := range memcachedInstances {
		if metricNames, ok := sysConfigMap["dcs_memcached_instance_id"]; ok {
			metrics := buildSingleDimensionMetrics(metricNames, "SYS.DCS", "dcs_memcached_instance_id", memcachedInstances[index].ID)
			*filterMetrics = append(*filterMetrics, metrics...)
			info := labelInfo{
				Name: []string{"name", "epId", "engine", "ip", "port", "cache_mode", "engine_version"},
				Value: []string{memcachedInstances[index].Name, memcachedInstances[index].EpId, memcachedInstances[index].Engine,
					memcachedInstances[index].IP, fmt.Sprintf("%d", memcachedInstances[index].Port), memcachedInstances[index].CacheMode, memcachedInstances[index].EngineVersion},
			}
			keys, values := getTags(memcachedInstances[index].Tags)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
		}
	}
}

func buildDcsNodesInfo(sysConfigMap map[string][]string, filterMetrics *[]model.MetricInfoList, resourceInfos map[string]labelInfo) {
	nodes, err := getDcsNodesFromRMS()
	if err != nil {
		return
	}
	for index := range nodes {
		dimName := getDimsNameKey(nodes[index].Dimensions)
		if metricNames, ok := sysConfigMap[dimName]; ok {
			metrics := buildDimensionMetrics(metricNames, "SYS.DCS", nodes[index].Dimensions)
			*filterMetrics = append(*filterMetrics, metrics...)
			info := labelInfo{
				Name: []string{"node_type", "group_name", "res_subnet_ip", "private_ip", "private_port"},
				Value: []string{nodes[index].NodeType, nodes[index].GroupName, nodes[index].ResSubnetIp,
					nodes[index].PrivateIp, nodes[index].PrivatePort},
			}
			keys, values := getTags(nodes[index].Tags)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			if label, exist := resourceInfos[nodes[index].InstanceId]; exist {
				info.Name = append(info.Name, label.Name...)
				info.Value = append(info.Value, label.Value...)
			}
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
		}
	}
}

func getRedisInstancesFromRMS() ([]DcsInstancesInfo, error) {
	resp, err := listResources("dcs", "redis")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of dcs.redis, error: %s", err.Error())
		return nil, err
	}
	instances := make([]DcsInstancesInfo, 0, len(resp))
	for _, resource := range resp {
		var properties RmsDcsInstanceProperties
		err := fmtResourceProperties(resource.Properties, &properties)
		if err != nil {
			logs.Logger.Errorf("fmt dcs instance properties error: %s", err.Error())
			continue
		}
		instances = append(instances, DcsInstancesInfo{
			ResourceBaseInfo: ResourceBaseInfo{ID: *resource.Id,
				Name: *resource.Name,
				EpId: *resource.EpId,
				Tags: resource.Tags},
			RmsDcsInstanceProperties: properties,
		})
	}
	return instances, nil
}

func getMemcachedInstancesFromRMS() ([]DcsInstancesInfo, error) {
	resp, err := listResources("dcs", "memcached")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of dcs.memcached, error: %s", err.Error())
		return nil, err
	}
	instances := make([]DcsInstancesInfo, 0, len(resp))
	for _, resource := range resp {
		var properties RmsDcsInstanceProperties
		err := fmtResourceProperties(resource.Properties, &properties)
		if err != nil {
			logs.Logger.Errorf("fmt dcs instance properties error: %s", err.Error())
			continue
		}
		instances = append(instances, DcsInstancesInfo{
			ResourceBaseInfo: ResourceBaseInfo{ID: *resource.Id,
				Name: *resource.Name,
				EpId: *resource.EpId,
				Tags: resource.Tags},
			RmsDcsInstanceProperties: properties,
		})
	}
	return instances, nil
}

type DcsNodeInfo struct {
	Tags map[string]string
	RmsDcsNodeProperties
}

func getDcsNodesFromRMS() ([]DcsNodeInfo, error) {
	resp, err := listResources("dcs", "node")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of dcs.node, error: %s", err.Error())
		return nil, err
	}
	nodes := make([]DcsNodeInfo, 0, len(resp))
	for _, resource := range resp {
		var nodeProperties RmsDcsNodeProperties
		err := fmtResourceProperties(resource.Properties, &nodeProperties)
		if err != nil {
			logs.Logger.Errorf("fmt dcs node properties error: %s", err.Error())
			continue
		}
		nodes = append(nodes, DcsNodeInfo{
			Tags:                 resource.Tags,
			RmsDcsNodeProperties: nodeProperties,
		})
	}
	return nodes, nil
}

type RmsDcsNodeProperties struct {
	NodeType    string                   `json:"node_type"`
	GroupName   string                   `json:"group_name"`
	ResSubnetIp string                   `json:"res_subnet_ip"`
	PrivateIp   string                   `json:"private_ip"`
	PrivatePort string                   `json:"private_port"`
	InstanceId  string                   `json:"instance_id"`
	Dimensions  []model.MetricsDimension `json:"dimensions"`
}

type RmsDcsInstanceProperties struct {
	Engine        string `json:"engine"`
	IP            string `json:"ip"`
	Port          int    `json:"port"`
	CacheMode     string `json:"cache_mode"`
	EngineVersion string `json:"engine_version"`
}
