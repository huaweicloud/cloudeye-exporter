package collector

import (
	"sync"
	"time"

	ces "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	cesv2 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v2"
	cesv2model "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v2/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var (
	host            string
	agentDimensions = sync.Map{}
)

type AgentDimensionsValue struct {
	originValue string
	ttl         int64
}

func getCESClient() *ces.CesClient {
	return ces.NewCesClient(ces.CesClientBuilder().
		WithCredential(authCredentialMap[conf.AuthMode](RegionServiceType)).
		WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
		WithEndpoint(getEndpoint("ces", "v1")).Build())
}

func batchQueryMetricData(metrics *[]model.MetricInfo, from, to int64) (*[]model.BatchMetricData, error) {
	options := &model.BatchListMetricDataRequest{
		Body: &model.BatchListMetricDataRequestBody{
			Metrics: *metrics,
			From:    from,
			To:      to,
			Period:  "1",
			Filter:  "average",
		},
	}

	v, err := getCESClient().BatchListMetricData(options)
	if err != nil {
		logs.Logger.Errorf("Failed to get metricdata: %s", err.Error())
		return nil, err
	}

	return v.Metrics, nil
}

func listAllMetrics(namespace string) ([]model.MetricInfoList, error) {
	limit := int32(1000)
	reqParam := &model.ListMetricsRequest{Limit: &limit, Namespace: &namespace}
	var metricData []model.MetricInfoList
	for {
		res, err := getCESClient().ListMetrics(reqParam)
		if err != nil {
			logs.Logger.Errorf("ListMetrics error, detail: %s", err.Error())
			break
		}
		if res.Metrics == nil {
			break
		}
		metrics := *(res.Metrics)
		if len(metrics) == 0 {
			break
		}
		reqParam.Start = &(res.MetaData.Marker)
		metricData = append(metricData, metrics...)
	}

	return metricData, nil
}

func getCESClientV2() *cesv2.CesClient {
	return cesv2.NewCesClient(ces.CesClientBuilder().WithCredential(authCredentialMap[conf.AuthMode](RegionServiceType)).
		WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
		WithEndpoint(getEndpoint("ces", "v2")).Build())
}

func getAgentOriginValue(instanceID, value string) string {
	originValue, ok := agentDimensions.Load(value)
	if !ok {
		err := loadAgentDimensions(instanceID)
		if err != nil {
			logs.Logger.Errorf("Load agent dimension error: %s", err.Error())
		}
		originValue, ok = agentDimensions.Load(value)
		if !ok {
			return value
		}
	}

	originV, ok := originValue.(AgentDimensionsValue)
	if !ok {
		return value
	}

	return originV.originValue
}

func loadAgentDimensions(instanceID string) error {
	dimName := cesv2model.GetListAgentDimensionInfoRequestDimNameEnum()
	dimNames := []cesv2model.ListAgentDimensionInfoRequestDimName{dimName.DISK,
		dimName.MOUNT_POINT, dimName.GPU, dimName.PROC, dimName.RAID}
	limit := int32(1000)
	offset := int32(0)
	for _, name := range dimNames {
		request := &cesv2model.ListAgentDimensionInfoRequest{
			InstanceId: instanceID,
			DimName:    name,
			Limit:      &limit,
			Offset:     &offset,
		}
		for {
			response, err := getCESClientV2().ListAgentDimensionInfo(request)
			if err != nil {
				logs.Logger.Errorf("Failed to list agent dimensions: %s", err.Error())
				return err
			}
			if response == nil || response.Dimensions == nil || len(*response.Dimensions) == 0 {
				break
			}
			for _, dimension := range *response.Dimensions {
				agentDimensions.Store(*dimension.Value, AgentDimensionsValue{
					originValue: *dimension.OriginValue,
					ttl:         time.Now().Add(time.Hour).Unix(),
				})
			}
			newOffset := *request.Offset + limit
			request.Offset = &newOffset
		}
	}
	return nil
}

func ClearAgentDimensionsCache() {
	go func() {
		ticker := time.NewTicker(time.Hour)
		for range ticker.C {
			agentDimensions.Range(func(key, value interface{}) bool {
				dimensionKey, keyOk := key.(string)
				if !keyOk {
					agentDimensions.Delete(dimensionKey)
					return false
				}
				dimensionValue, valueOk := value.(AgentDimensionsValue)
				if !valueOk {
					agentDimensions.Delete(dimensionKey)
					return false
				}
				unix := time.Now().Unix()
				if dimensionValue.ttl < unix {
					agentDimensions.Delete(dimensionKey)
				}
				return true
			})
		}
	}()
}
