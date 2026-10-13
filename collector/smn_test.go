package collector

import (
	"fmt"
	"testing"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	smnModel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/smn/v2/model"
	"github.com/stretchr/testify/assert"
)

func TestSMNGetResourceInfo(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	smnInfo = serversInfo{}

	patches.ApplyFuncReturn(GetTopicInfo, []smnModel.ListTopicsItem{
		{Name: "topic-A", TopicId: "topic-id-1"},
		{Name: "topic-B", TopicId: "topic-id-2"},
	}, nil)

	metrics := []model.MetricInfoList{
		{Namespace: "SYS.SMN", MetricName: "msg_count"},
		{Namespace: "SYS.SMN", MetricName: "msg_size"},
	}
	patches.ApplyFuncReturn(listAllMetrics, metrics, nil)
	patches.ApplyFuncReturn(IsMetricInfoInWhiteList, true)
	patches.ApplyFuncReturn(GetResourceInfoExpirationTime, 5*time.Minute)

	var info SMNInfo
	labels, resultMetrics := info.GetResourceInfo()

	assert.Equal(t, 2, len(labels))
	assert.Equal(t, []string{"topicName"}, labels["topic-id-1"].Name)
	assert.Equal(t, []string{"topic-A"}, labels["topic-id-1"].Value)
	assert.Equal(t, []string{"topicName"}, labels["topic-id-2"].Name)
	assert.Equal(t, []string{"topic-B"}, labels["topic-id-2"].Value)
	assert.Equal(t, 2, len(resultMetrics))
	assert.Greater(t, smnInfo.TTL, time.Now().Unix()-1)
}

func TestSMNGetResourceInfoCache(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	cachedLabels := map[string]labelInfo{
		"cached-id": {Name: []string{"topicName"}, Value: []string{"cached-topic"}},
	}
	cachedMetrics := []model.MetricInfoList{
		{Namespace: "SYS.SMN", MetricName: "cached_metric"},
	}
	smnInfo = serversInfo{
		TTL:           time.Now().Add(10 * time.Minute).Unix(),
		LabelInfo:     cachedLabels,
		FilterMetrics: cachedMetrics,
	}

	patches.ApplyFunc(GetTopicInfo, func() ([]smnModel.ListTopicsItem, error) {
		t.Fatal("GetTopicInfo should not be called when cache is valid")
		return nil, nil
	})
	patches.ApplyFunc(listAllMetrics, func(namespace string) ([]model.MetricInfoList, error) {
		t.Fatal("listAllMetrics should not be called when cache is valid")
		return nil, nil
	})

	var info SMNInfo
	labels, resultMetrics := info.GetResourceInfo()

	assert.Equal(t, 1, len(labels))
	assert.Equal(t, []string{"cached-topic"}, labels["cached-id"].Value)
	assert.Equal(t, 1, len(resultMetrics))
	assert.Equal(t, "cached_metric", resultMetrics[0].MetricName)

	smnInfo = serversInfo{}
}

func TestSMNGetResourceInfoTopicError(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	smnInfo = serversInfo{}

	patches.ApplyFuncReturn(GetTopicInfo, nil, fmt.Errorf("network error"))
	patches.ApplyFuncReturn(GetResourceInfoExpirationTime, 5*time.Minute)

	var info SMNInfo
	labels, metrics := info.GetResourceInfo()

	assert.Nil(t, labels)
	assert.Nil(t, metrics)
}

func TestSMNGetResourceInfoMetricsError(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	smnInfo = serversInfo{}

	patches.ApplyFuncReturn(GetTopicInfo, []smnModel.ListTopicsItem{
		{Name: "topic-err", TopicId: "topic-err-id"},
	}, nil)
	patches.ApplyFuncReturn(listAllMetrics, nil, fmt.Errorf("metrics fetch error"))
	patches.ApplyFuncReturn(GetResourceInfoExpirationTime, 5*time.Minute)

	var info SMNInfo
	labels, metrics := info.GetResourceInfo()

	assert.Nil(t, labels)
	assert.Nil(t, metrics)
}

func TestSMNGetResourceInfoMetricFilter(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	smnInfo = serversInfo{}

	patches.ApplyFuncReturn(GetTopicInfo, []smnModel.ListTopicsItem{
		{Name: "topic-filter", TopicId: "topic-filter-id"},
	}, nil)

	allMetrics := []model.MetricInfoList{
		{Namespace: "SYS.SMN", MetricName: "msg_count"},
		{Namespace: "SYS.SMN", MetricName: "internal_debug"},
		{Namespace: "SYS.SMN", MetricName: "msg_size"},
	}
	patches.ApplyFuncReturn(listAllMetrics, allMetrics, nil)
	patches.ApplyFunc(IsMetricInfoInWhiteList, func(metricInfo model.MetricInfoList) bool {
		return metricInfo.MetricName != "internal_debug"
	})
	patches.ApplyFuncReturn(GetResourceInfoExpirationTime, 5*time.Minute)

	var info SMNInfo
	labels, resultMetrics := info.GetResourceInfo()

	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 2, len(resultMetrics))
	assert.Equal(t, "msg_count", resultMetrics[0].MetricName)
	assert.Equal(t, "msg_size", resultMetrics[1].MetricName)
}

func TestSMNGetResourceInfoEmptyTopics(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	smnInfo = serversInfo{}

	patches.ApplyFuncReturn(GetTopicInfo, []smnModel.ListTopicsItem{}, nil)
	patches.ApplyFuncReturn(listAllMetrics, []model.MetricInfoList{
		{Namespace: "SYS.SMN", MetricName: "msg_count"},
	}, nil)
	patches.ApplyFuncReturn(IsMetricInfoInWhiteList, true)
	patches.ApplyFuncReturn(GetResourceInfoExpirationTime, 5*time.Minute)

	var info SMNInfo
	labels, resultMetrics := info.GetResourceInfo()

	assert.Equal(t, 0, len(labels))
	assert.Equal(t, 1, len(resultMetrics))
}

func TestSMNGetResourceInfoCacheExpired(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	smnInfo = serversInfo{
		TTL:       time.Now().Add(-1 * time.Minute).Unix(),
		LabelInfo: map[string]labelInfo{"old-id": {Name: []string{"topicName"}, Value: []string{"old-topic"}}},
	}

	patches.ApplyFuncReturn(GetTopicInfo, []smnModel.ListTopicsItem{
		{Name: "new-topic", TopicId: "new-id"},
	}, nil)
	patches.ApplyFuncReturn(listAllMetrics, []model.MetricInfoList{
		{Namespace: "SYS.SMN", MetricName: "new_metric"},
	}, nil)
	patches.ApplyFuncReturn(IsMetricInfoInWhiteList, true)
	patches.ApplyFuncReturn(GetResourceInfoExpirationTime, 5*time.Minute)

	var info SMNInfo
	labels, resultMetrics := info.GetResourceInfo()

	assert.Equal(t, 1, len(labels))
	assert.Equal(t, []string{"new-topic"}, labels["new-id"].Value)
	assert.Equal(t, 1, len(resultMetrics))
	assert.Equal(t, "new_metric", resultMetrics[0].MetricName)

	smnInfo = serversInfo{}
}
