package collector

import (
	"strings"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	nosql "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/gaussdbfornosql/v3"
	nosqlmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/gaussdbfornosql/v3/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var (
	nosqlInfo serversInfo
	dimMap    = map[string]string{
		"cassandra": "cassandra_cluster_id,cassandra_node_id",
		"mongodb":   "mongodb_cluster_id,mongodb_node_id",
		"influxdb":  "influxdb_cluster_id,influxdb_node_id",
		"redis":     "redis_cluster_id,redis_node_id",
	}
)

type NoSQLInfo struct{}

func (getter NoSQLInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	nosqlInfo.Lock()
	defer nosqlInfo.Unlock()
	if nosqlInfo.LabelInfo == nil || time.Now().Unix() > nosqlInfo.TTL {
		instances, err := getAllNoSQLInstances()
		if err != nil {
			logs.Logger.Errorf("Get All NoSQL Instances error: %s", err.Error())
			return nosqlInfo.LabelInfo, nosqlInfo.FilterMetrics
		}
		for _, instance := range instances {
			buildMetricsAndInfo(instance, &filterMetrics, resourceInfos)
		}
		nosqlInfo.LabelInfo = resourceInfos
		nosqlInfo.FilterMetrics = filterMetrics
		nosqlInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return nosqlInfo.LabelInfo, nosqlInfo.FilterMetrics
}

func buildMetricsAndInfo(instance nosqlmodel.ListInstancesResult, filterMetrics *[]model.MetricInfoList, resourceInfos map[string]labelInfo) {
	dimStr, ok := dimMap[instance.Datastore.Type]
	if !ok {
		logs.Logger.Debugf("Instances type is invalid")
		return
	}
	metricNames := getMetricConfigMap("SYS.NoSQL")[dimStr]
	if len(metricNames) == 0 {
		logs.Logger.Debugf("metric names is empty: %s", dimStr)
		return
	}
	dimName := strings.Split(dimStr, ",")
	for _, group := range instance.Groups {
		for _, node := range group.Nodes {
			metrics := buildDimensionMetrics(metricNames, "SYS.NoSQL",
				[]model.MetricsDimension{{Name: dimName[0], Value: instance.Id}, {Name: dimName[1], Value: node.Id}})
			*filterMetrics = append(*filterMetrics, metrics...)
			nodeInfo := labelInfo{
				Name: []string{"instanceName", "lbIPAddress", "lbPort", "epId", "type", "nodeName", "nodePrivateIP", "nodePublicIp"},
				Value: []string{instance.Name, getDefaultString(instance.LbIpAddress), getDefaultString(instance.LbPort), instance.EnterpriseProjectId, instance.Datastore.Type,
					node.Name, node.PrivateIp, node.PublicIp},
			}
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = nodeInfo
		}
	}
}

func getAllNoSQLInstances() ([]nosqlmodel.ListInstancesResult, error) {
	limit := int32(100)
	offset := int32(0)
	options := &nosqlmodel.ListInstancesRequest{Limit: &limit, Offset: &offset}
	var instances []nosqlmodel.ListInstancesResult
	client := getNoSQLClient()
	for {
		response, err := client.ListInstances(options)
		if err != nil {
			logs.Logger.Errorf("list nosql instances: %s", err.Error())
			return instances, err
		}
		if len(*response.Instances) == 0 {
			break
		}
		instances = append(instances, *response.Instances...)
		*options.Offset += limit
	}
	return instances, nil
}

func getNoSQLClient() *nosql.GaussDBforNoSQLClient {
	return nosql.NewGaussDBforNoSQLClient(nosql.GaussDBforNoSQLClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(true)).
		WithEndpoint(getEndpoint("gaussdb-nosql", "v3")).Build())
}
