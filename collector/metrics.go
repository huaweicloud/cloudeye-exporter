package collector

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	ces "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var (
	host string
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
