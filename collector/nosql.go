package collector

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	nosql "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/gaussdbfornosql/v3"
	nosqlmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/gaussdbfornosql/v3/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var (
	nosqlInfo serversInfo
	dimMap    = map[string][]string{
		"cassandra": {"cassandra_cluster_id,cassandra_node_id"},
		"mongodb":   {"mongodb_cluster_id", "mongodb_cluster_id,mongodb_node_id"},
		"influxdb":  {"influxdb_cluster_id", "influxdb_cluster_id,influxdb_node_id"},
		"redis":     {"redis_cluster_id", "redis_cluster_id,redis_node_id"},
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
			buildMetricsAndInfo(instance, resourceInfos)
		}

		dynamoTables, err := getDynamoDbInstances()
		if err != nil {
			logs.Logger.Errorf("Get All NoSQL Dynamo Tables error: %s", err.Error())
			return nosqlInfo.LabelInfo, nosqlInfo.FilterMetrics
		}

		for _, table := range dynamoTables {
			buildDynamoTableResources(table, resourceInfos)
		}

		allMetrics, err := listAllMetrics("SYS.NoSQL")
		if err != nil {
			logs.Logger.Errorf("Get all metrics of SYS.NoSQLS error: %s", err.Error())
			return nosqlInfo.LabelInfo, nosqlInfo.FilterMetrics
		}

		for _, metricInfo := range allMetrics {
			resourceKey := GetResourceKeyFromMetricInfo(metricInfo)
			if resourceKey == "" {
				continue
			}
			if _, ok := resourceInfos[resourceKey]; !ok {
				continue
			}

			if IsMetricInfoInWhiteList(metricInfo) {
				filterMetrics = append(filterMetrics, metricInfo)
			}
		}

		nosqlInfo.LabelInfo = resourceInfos
		nosqlInfo.FilterMetrics = filterMetrics
		nosqlInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	}
	return nosqlInfo.LabelInfo, nosqlInfo.FilterMetrics
}

func buildDynamoTableResources(table DynamoTableInfo, infos map[string]labelInfo) {
	instanceInfo := labelInfo{
		Name:  []string{"name", "status"},
		Value: []string{table.Name, table.Status},
	}
	infos[table.Id] = instanceInfo
}

func buildMetricsAndInfo(instance nosqlmodel.ListInstancesResult, resourceInfos map[string]labelInfo) {
	dimStrArr, ok := dimMap[instance.Datastore.Type]
	if !ok {
		logs.Logger.Debugf("Instances type is invalid")
		return
	}
	for _, dimStr := range dimStrArr {
		dimName := strings.Split(dimStr, ",")
		if len(dimName) == 1 {
			buildClusterResources(instance, resourceInfos)
		} else {
			buildNodeDimResources(instance, dimName, resourceInfos)
		}
	}
}

func buildNodeDimResources(instance nosqlmodel.ListInstancesResult, dimName []string, resourceInfos map[string]labelInfo) {
	if resourceInfos == nil {
		resourceInfos = make(map[string]labelInfo)
	}
	for _, group := range instance.Groups {
		for _, node := range group.Nodes {
			nodeInfo := labelInfo{
				Name: []string{"instanceName", "lbIPAddress", "lbPort", "epId", "type", "mode", "nodeName", "nodePrivateIP", "nodePublicIp"},
				Value: []string{instance.Name, getDefaultString(instance.LbIpAddress), getDefaultString(instance.LbPort), instance.EnterpriseProjectId, instance.Datastore.Type, instance.Mode,
					node.Name, node.PrivateIp, node.PublicIp},
			}
			resourceInfos[GetResourceKeyFromDimensions([]model.MetricsDimension{{Name: dimName[0], Value: instance.Id}, {Name: dimName[1], Value: node.Id}})] = nodeInfo
		}
	}
}

func buildClusterResources(instance nosqlmodel.ListInstancesResult, resourceInfos map[string]labelInfo) {
	instanceInfo := labelInfo{
		Name:  []string{"instanceName", "lbIPAddress", "lbPort", "epId", "type", "mode"},
		Value: []string{instance.Name, getDefaultString(instance.LbIpAddress), getDefaultString(instance.LbPort), instance.EnterpriseProjectId, instance.Datastore.Type, instance.Mode},
	}
	resourceInfos[instance.Id] = instanceInfo
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

type DynamoTableInfo struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type DynamoTablesResp struct {
	TotalCount     int               `json:"total_count"`
	Tables         []DynamoTableInfo `json:"tables"`
	HttpStatusCode int               `json:"-"`
}

type ListDynamoTablesRep struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func genReqDefForListDynamoTables() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().WithMethod(http.MethodGet).WithPath("/v3/{project_id}/serverless/dynamodb/tables").
		WithResponse(new(DynamoTablesResp)).WithContentType("application/json")

	reqDefBuilder.WithRequestField(def.NewFieldDef().WithName("Limit").WithJsonTag("limit").WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().WithName("Offset").WithJsonTag("offset").WithLocationType(def.Query))
	return reqDefBuilder.Build()
}

func getDynamoDbInstances() ([]DynamoTableInfo, error) {
	// oidc场景下实际使用的是iam token鉴权，iam5接口无法使用token鉴权；
	// dynamotable查询接口仅支持iam5，不支持iam3；
	// 因此在oidc场景下，调用dynamotable接口查询资源会报错，影响其他类型nosql资源指标的导出，需去除对dynamotable指标导出的支持；
	// 后续等dynamotable接口支持iam3后，再开放dynamotable指标的导出
	if conf.AuthMode == AuthModeOidcToken {
		return []DynamoTableInfo{}, nil
	}

	client := getNoSQLClient()
	var tables []DynamoTableInfo
	request := &ListDynamoTablesRep{Limit: 100}
	for {
		resp, err := getDynamoTables(client, request)
		if err != nil {
			logs.Logger.Errorf("Failed to get list dynamo tables: %s", err.Error())
			return nil, err
		}
		response, ok := resp.(*DynamoTablesResp)
		if !ok {
			err := errors.New("resp type is not DynamoTablesResp")
			logs.Logger.Errorf("Failed to get list dynamo tables: %s", err.Error())
			return nil, err
		}
		if len(response.Tables) == 0 {
			break
		}
		tables = append(tables, response.Tables...)
		request.Offset += 100
	}
	return tables, nil
}

func getDynamoTables(client *nosql.GaussDBforNoSQLClient, request *ListDynamoTablesRep) (interface{}, error) {
	return client.HcClient.Sync(request, genReqDefForListDynamoTables())
}

func getNoSQLClient() *nosql.GaussDBforNoSQLClient {
	return nosql.NewGaussDBforNoSQLClient(nosql.GaussDBforNoSQLClientBuilder().WithCredential(
		authCredentialMap[conf.AuthMode](RegionServiceType)).
		WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
		WithEndpoint(getEndpoint("gaussdb-nosql", "v3")).Build())
}
