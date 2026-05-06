package collector

import (
	"time"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	smn "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/smn/v2"
	smnModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/smn/v2/model"
)

var smnInfo serversInfo

type SMNInfo struct{}

func (getter SMNInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	smnInfo.Lock()
	defer smnInfo.Unlock()
	if smnInfo.LabelInfo == nil || time.Now().Unix() > smnInfo.TTL {
		topicList, err := GetTopicInfo()
		if err != nil {
			logs.Logger.Errorf("Get topic from service error: %s", err.Error())
			return smnInfo.LabelInfo, smnInfo.FilterMetrics
		}

		for _, topic := range topicList {
			info := labelInfo{
				Name:  []string{"topicName"},
				Value: []string{topic.Name},
			}
			resourceInfos[topic.TopicId] = info
		}

		allMetrics, err := listAllMetrics("SYS.SMN")
		if err != nil {
			logs.Logger.Errorf("Get all metrics of SYS.SMN error: %s", err.Error())
			return smnInfo.LabelInfo, smnInfo.FilterMetrics
		}

		var filteredMetrics []model.MetricInfoList
		for _, metricInfo := range allMetrics {
			if IsMetricInfoInWhiteList(metricInfo) {
				filteredMetrics = append(filteredMetrics, metricInfo)
			}
		}

		smnInfo.LabelInfo = resourceInfos
		smnInfo.FilterMetrics = filteredMetrics
		smnInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()

	}
	return smnInfo.LabelInfo, smnInfo.FilterMetrics
}

func getSMNClient() *smn.SmnClient {
	return smn.NewSmnClient(smn.SmnClientBuilder().WithCredential(
		authCredentialMap[conf.AuthMode](RegionServiceType)).
		WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
		WithEndpoint(getEndpoint("smn", "v2")).Build())
}

func GetTopicInfo() ([]smnModel.ListTopicsItem, error) {
	limit := int32(100)
	offset := int32(0)
	request := &smnModel.ListTopicsRequest{Limit: &limit, Offset: &offset}
	var topicList []smnModel.ListTopicsItem
	for {
		response, err := getSMNClient().ListTopics(request)
		if err != nil {
			logs.Logger.Errorf("list topics error: %s, limit: %d, offset: %d", err.Error(),
				*request.Limit, *request.Offset)
			return nil, err
		}
		tempClusters := *response.Topics
		if len(tempClusters) == 0 {
			break
		}
		topicList = append(topicList, tempClusters...)
		*request.Offset += limit
	}
	return topicList, nil
}
