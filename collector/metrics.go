package collector

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	ces "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	cesv2 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v2"
	cesv2model "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v2/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var (
	host string
	agentDimensions = make(map[string]string, 0)
)

func getCESClient() *ces.CesClient {
	return ces.NewCesClient(ces.CesClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(true)).
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
	return cesv2.NewCesClient(ces.CesClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(true)).
		WithEndpoint(getEndpoint("ces", "v2")).Build())
}

func getAgentOriginValue(value string) string {
	originValue, ok := agentDimensions[value]
	if ok {
		return originValue
	}
	return value
}

func loadAgentDimensions(instanceID string) {
	dimName := cesv2model.GetListAgentDimensionInfoRequestDimNameEnum()
	dimNames := []cesv2model.ListAgentDimensionInfoRequestDimName{dimName.DISK,
		dimName.MOUNT_POINT, dimName.GPU, dimName.PROC, dimName.RAID}
	for _, name := range dimNames {
		request := &cesv2model.ListAgentDimensionInfoRequest{
			ContentType: "application/json",
			InstanceId:  instanceID,
			DimName:     name,
		}
		response, err := getCESClientV2().ListAgentDimensionInfo(request)
		if err != nil {
			logs.Logger.Errorf("Failed to list agentDimensions: %s", err.Error())
			return
		}
		for _, dimension := range *response.Dimensions {
			agentDimensions[*dimension.Value] = *dimension.OriginValue
		}
	}
}
