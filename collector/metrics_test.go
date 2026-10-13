package collector

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/snappy"
	ces "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

// initTestLogger 初始化测试用的 logger
func initTestLogger() {
	tmpDir := os.TempDir()
	logsYml := filepath.Join(tmpDir, "test_logs.yml")
	content := []byte(`business:
  - level: debug
    type: CONSOLE
    enabled: true
    encoder: CONSOLE
    time_format: "2006-01-02 15:04:05"
`)
	_ = os.WriteFile(logsYml, content, 0644)
	logs.InitLog(logsYml)
}

// ======================== init / writeHttpClient Tests ========================

func TestWriteHttpClient_InitializedByInit(t *testing.T) {
	// init() 已经初始化了 writeHttpClient
	assert.NotNil(t, writeHttpClient)
	assert.Equal(t, 10*time.Second, writeHttpClient.Timeout)
}

func TestWriteHttpClient_TransportConfig(t *testing.T) {
	transport, ok := writeHttpClient.Transport.(*http.Transport)
	assert.True(t, ok)
	assert.NotNil(t, transport.TLSClientConfig)
	assert.Equal(t, uint16(0x0303), transport.TLSClientConfig.MinVersion) // TLS 1.2
}

// ======================== doWriteRequest / writeMetricsToPrometheus Tests ========================

func TestWriteMetricsToPrometheus_Success(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/x-protobuf", r.Header.Get("Content-Type"))
		assert.Equal(t, "snappy", r.Header.Get("Content-Encoding"))
		assert.Equal(t, "0.1.0", r.Header.Get("X-Prometheus-Remote-Write-Version"))

		var err error
		receivedBody, err = io.ReadAll(r.Body)
		assert.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	CloudConf.PromWriteConf.Url = server.URL

	series := []*TimeSeries{
		{
			Labels: []*Label{
				{Name: "__name__", Value: "huaweicloud_sys_ecs_cpu_util"},
				{Name: "instance_id", Value: "i-001"},
			},
			Samples: []*Sample{{Value: 42.0, Timestamp: 1700000000000}},
		},
	}
	writeMetricsToPrometheus(series)

	// 解压并反序列化接收到的数据
	decompressed, err := snappy.Decode(nil, receivedBody)
	assert.NoError(t, err)

	req := &WriteRequest{}
	err = proto.Unmarshal(decompressed, req)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(req.Timeseries))
	assert.Equal(t, "huaweicloud_sys_ecs_cpu_util", req.Timeseries[0].Labels[0].Value)
}

func TestWriteMetricsToPrometheus_ServerError(t *testing.T) {
	initTestLogger()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("internal error"))
		if err != nil {
			assert.NotNil(t, err)
		}
	}))
	defer server.Close()

	CloudConf.PromWriteConf.Url = server.URL

	series := []*TimeSeries{
		{
			Labels:  []*Label{{Name: "__name__", Value: "test"}},
			Samples: []*Sample{{Value: 1.0, Timestamp: 100}},
		},
	}
	// 不应 panic，错误只记录日志
	writeMetricsToPrometheus(series)
}

func TestWriteMetricsToPrometheus_BadRequest(t *testing.T) {
	initTestLogger()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("bad request"))
		if err != nil {
			assert.NotNil(t, err)
		}
	}))
	defer server.Close()

	CloudConf.PromWriteConf.Url = server.URL

	series := []*TimeSeries{
		{
			Labels:  []*Label{{Name: "__name__", Value: "test"}},
			Samples: []*Sample{{Value: 1.0, Timestamp: 100}},
		},
	}
	writeMetricsToPrometheus(series)
}

func TestWriteMetricsToPrometheus_EmptySeries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	CloudConf.PromWriteConf.Url = server.URL

	series := []*TimeSeries{}
	writeMetricsToPrometheus(series)
}

func TestWriteMetricsToPrometheus_MultipleTimeseries(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	CloudConf.PromWriteConf.Url = server.URL

	series := []*TimeSeries{
		{Labels: []*Label{{Name: "__name__", Value: "metric_a"}}, Samples: []*Sample{{Value: 1.0, Timestamp: 100}}},
		{Labels: []*Label{{Name: "__name__", Value: "metric_b"}}, Samples: []*Sample{{Value: 2.0, Timestamp: 200}}},
		{Labels: []*Label{{Name: "__name__", Value: "metric_c"}}, Samples: []*Sample{{Value: 3.0, Timestamp: 300}}},
	}
	writeMetricsToPrometheus(series)
	assert.Equal(t, 1, callCount)
}

func TestWriteMetricsToPrometheus_InvalidURL(t *testing.T) {
	initTestLogger()
	CloudConf.PromWriteConf.Url = "http://invalid-host-that-does-not-exist:9999"

	series := []*TimeSeries{
		{Labels: []*Label{{Name: "__name__", Value: "test"}}, Samples: []*Sample{{Value: 1.0, Timestamp: 100}}},
	}
	// 不应 panic，错误只记录日志
	writeMetricsToPrometheus(series)
}

// ======================== transformCESLabelToPromLabels Tests ========================

func TestTransformCESLabelToPromLabels(t *testing.T) {
	promLabels := []*Label{{Name: "__name__", Value: "test_metric"}}
	cesLabel := labelInfo{
		Name:  []string{"instance_id", "az"},
		Value: []string{"i-001", "az1"},
	}

	result := transformCESLabelToPromLabels(promLabels, cesLabel)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "__name__", result[0].Name)
	assert.Equal(t, "instance_id", result[1].Name)
	assert.Equal(t, "i-001", result[1].Value)
	assert.Equal(t, "az", result[2].Name)
	assert.Equal(t, "az1", result[2].Value)
}

func TestTransformCESLabelToPromLabels_EmptyCESLabel(t *testing.T) {
	promLabels := []*Label{{Name: "__name__", Value: "test_metric"}}
	cesLabel := labelInfo{}

	result := transformCESLabelToPromLabels(promLabels, cesLabel)
	assert.Equal(t, 1, len(result))
}

func TestTransformCESLabelToPromLabels_EmptyPromLabels(t *testing.T) {
	promLabels := []*Label{}
	cesLabel := labelInfo{
		Name:  []string{"host"},
		Value: []string{"server01"},
	}

	result := transformCESLabelToPromLabels(promLabels, cesLabel)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "host", result[0].Name)
}

func TestTransformCESLabelToPromLabels_MultipleLabels(t *testing.T) {
	promLabels := []*Label{{Name: "__name__", Value: "cpu_util"}}
	cesLabel := labelInfo{
		Name:  []string{"instance_id", "az", "region", "epId"},
		Value: []string{"i-001", "az1", "cn-north-4", "ep-001"},
	}

	result := transformCESLabelToPromLabels(promLabels, cesLabel)
	assert.Equal(t, 5, len(result))
}

// ======================== getAgentOriginValue Tests ========================

func TestGetAgentOriginValue_Cached(t *testing.T) {
	agentDimensions.Store("test_hash_value", AgentDimensionsValue{
		originValue: "/dev/vda1",
		ttl:         time.Now().Add(time.Hour).Unix(),
	})
	defer agentDimensions.Delete("test_hash_value")

	result := getAgentOriginValue("i-001", "test_hash_value")
	assert.Equal(t, "/dev/vda1", result)
}

func TestGetAgentOriginValue_NotCached_LoadFails(t *testing.T) {
	patches := gomonkey.ApplyFunc(loadAgentDimensions, func(instanceID string) error {
		return nil
	})
	defer patches.Reset()

	result := getAgentOriginValue("i-001", "nonexistent_hash")
	assert.Equal(t, "nonexistent_hash", result)
}

func TestGetAgentOriginValue_WrongType(t *testing.T) {
	agentDimensions.Store("wrong_type_key_test", "not_an_AgentDimensionsValue")
	defer agentDimensions.Delete("wrong_type_key_test")

	result := getAgentOriginValue("i-001", "wrong_type_key_test")
	assert.Equal(t, "wrong_type_key_test", result)
}

func TestGetAgentOriginValue_CachedButExpired(t *testing.T) {
	// 存储一个已过期的条目（但仍在 map 中）
	agentDimensions.Store("expired_hash", AgentDimensionsValue{
		originValue: "/dev/vda2",
		ttl:         time.Now().Add(-time.Hour).Unix(),
	})
	defer agentDimensions.Delete("expired_hash")

	// 即使 TTL 过期，只要 key 还在 map 中，仍然返回 originValue
	// （过期清理由 ClearAgentDimensionsCache 的定时任务负责）
	result := getAgentOriginValue("i-001", "expired_hash")
	assert.Equal(t, "/dev/vda2", result)
}

// ======================== listAllMetrics Caching Tests ========================

func TestListAllMetrics_CacheHit(t *testing.T) {
	namespace := "SYS.ECS_TEST_CACHE_HIT"
	cachedMetrics := []model.MetricInfoList{
		{Namespace: namespace, MetricName: "cpu_util"},
	}
	filterMetricsCache.Store(namespace, FilterMetricsCacheItem{
		FilterMetrics: cachedMetrics,
		TTL:           time.Now().Add(time.Hour).Unix(),
	})
	defer filterMetricsCache.Delete(namespace)

	result, err := listAllMetrics(namespace)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "cpu_util", result[0].MetricName)
}

func TestListAllMetrics_CacheExpired_CallsAPI(t *testing.T) {
	initTestLogger()
	namespace := "SYS.ECS_TEST_CACHE_EXPIRED"

	// 先存一个过期的缓存
	filterMetricsCache.Store(namespace, FilterMetricsCacheItem{
		FilterMetrics: []model.MetricInfoList{},
		TTL:           time.Now().Add(-time.Hour).Unix(),
	})
	defer filterMetricsCache.Delete(namespace)

	// mock getCESClient 返回一个 mock client
	mockMetrics := []model.MetricInfoList{
		{Namespace: namespace, MetricName: "mem_util"},
	}
	mockResp := &model.ListMetricsResponse{
		Metrics: &mockMetrics,
		MetaData: &model.MetaData{
			Marker: "next_marker",
		},
	}

	// 第二次调用返回空，终止分页
	emptyResp := &model.ListMetricsResponse{
		Metrics: &[]model.MetricInfoList{},
	}

	callCount := 0
	patches := gomonkey.ApplyFunc(getCESClient, func() *ces.CesClient {
		return nil
	})
	defer patches.Reset()

	patches.ApplyMethodFunc(&ces.CesClient{}, "ListMetrics",
		func(request *model.ListMetricsRequest) (*model.ListMetricsResponse, error) {
			callCount++
			if callCount == 1 {
				return mockResp, nil
			}
			return emptyResp, nil
		})

	result, err := listAllMetrics(namespace)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "mem_util", result[0].MetricName)
}

func TestListAllMetrics_APIError(t *testing.T) {
	initTestLogger()
	namespace := "SYS.ECS_TEST_API_ERROR"

	// 确保缓存中没有该 namespace
	filterMetricsCache.Delete(namespace)

	patches := gomonkey.ApplyFunc(getCESClient, func() *ces.CesClient {
		return nil
	})
	defer patches.Reset()

	patches.ApplyMethodFunc(&ces.CesClient{}, "ListMetrics",
		func(request *model.ListMetricsRequest) (*model.ListMetricsResponse, error) {
			return nil, fmt.Errorf("API error")
		})

	result, err := listAllMetrics(namespace)
	assert.NoError(t, err) // API 错误不返回 error，只记录日志
	assert.Equal(t, 0, len(result))
}

// ======================== renderAndWriteMetricData Tests ========================

func TestRenderAndWriteMetricData(t *testing.T) {
	initTestLogger()
	CloudConf.Global.Prefix = "huaweicloud"
	namespace := "SYS.ECS"
	avg := 85.5
	unit := "%"
	metrics := []model.BatchMetricData{
		{
			Namespace:  &namespace,
			MetricName: "cpu_util",
			Dimensions: &[]model.MetricsDimension{
				{Name: "instance_id", Value: "i-001"},
			},
			Datapoints: []model.DatapointForBatchMetric{
				{Average: &avg, Timestamp: time.Now().UnixMilli()},
			},
			Unit: &unit,
		},
	}
	resourceInfos := map[string]labelInfo{
		"i-001": {Name: []string{"name"}, Value: []string{"server01"}},
	}

	ch := make(chan *TimeSeries, 10)
	renderAndWriteMetricData(&metrics, resourceInfos, ch)
	close(ch)

	var results []*TimeSeries
	for ts := range ch {
		results = append(results, ts)
	}
	require.Equal(t, 1, len(results), "Should have exactly one result")
	require.NotNil(t, results[0], "Result should not be nil")
	require.Len(t, results[0].Samples, 1, "Should have exactly one sample")
	assert.Equal(t, 85.5, results[0].Samples[0].Value, "Sample value should be 85.5")
	assert.NotZero(t, results[0].Samples[0].Timestamp, "Sample timestamp should not be zero")
	require.NotEmpty(t, results[0].Labels, "Result should have labels")
}

func TestRenderAndWriteMetricData_EmptyDatapoints(t *testing.T) {
	initTestLogger()
	namespace := "SYS.ECS"
	metrics := []model.BatchMetricData{
		{
			Namespace:  &namespace,
			MetricName: "cpu_util",
			Dimensions: &[]model.MetricsDimension{},
			Datapoints: []model.DatapointForBatchMetric{},
		},
	}
	resourceInfos := map[string]labelInfo{}
	ch := make(chan *TimeSeries, 10)

	renderAndWriteMetricData(&metrics, resourceInfos, ch)
	close(ch)

	var results []*TimeSeries
	for ts := range ch {
		results = append(results, ts)
	}
	assert.Equal(t, 0, len(results))
}

func TestRenderAndWriteMetricData_MultipleDatapoints_UsesLatest(t *testing.T) {
	initTestLogger()
	CloudConf.Global.Prefix = "huaweicloud"
	namespace := "SYS.ECS"
	unit := "%"
	avg1 := 50.0
	avg2 := 80.0
	metrics := []model.BatchMetricData{
		{
			Namespace:  &namespace,
			MetricName: "cpu_util",
			Unit:       &unit,
			Dimensions: &[]model.MetricsDimension{
				{Name: "instance_id", Value: "i-001"},
			},
			Datapoints: []model.DatapointForBatchMetric{
				{Average: &avg1, Timestamp: 1000},
				{Average: &avg2, Timestamp: 2000},
			},
		},
	}
	resourceInfos := map[string]labelInfo{}
	ch := make(chan *TimeSeries, 10)

	renderAndWriteMetricData(&metrics, resourceInfos, ch)
	close(ch)

	var results []*TimeSeries
	for ts := range ch {
		results = append(results, ts)
	}
	assert.Equal(t, 1, len(results))
	// 使用最后一个 datapoint
	assert.Equal(t, 80.0, results[0].Samples[0].Value)
	assert.Equal(t, int64(2000), results[0].Samples[0].Timestamp)
}

func TestRenderAndWriteMetricData_DuplicateLabels_Skipped(t *testing.T) {
	initTestLogger()
	CloudConf.Global.Prefix = "huaweicloud"
	namespace := "SYS.ECS"
	unit := "%"
	avg := 50.0
	metrics := []model.BatchMetricData{
		{
			Namespace:  &namespace,
			MetricName: "cpu_util",
			Unit:       &unit,
			Dimensions: &[]model.MetricsDimension{
				{Name: "instance_id", Value: "i-001"},
			},
			Datapoints: []model.DatapointForBatchMetric{
				{Average: &avg, Timestamp: 1000},
			},
		},
	}
	// resourceInfos 中包含与系统标签同名的自定义标签，导致重复
	resourceInfos := map[string]labelInfo{
		"i-001": {Name: []string{"instance_id"}, Value: []string{"i-001"}}, // 与系统标签重复
	}
	ch := make(chan *TimeSeries, 10)

	renderAndWriteMetricData(&metrics, resourceInfos, ch)
	close(ch)

	var results []*TimeSeries
	for ts := range ch {
		results = append(results, ts)
	}
	// 由于标签重复，该 metric 应被跳过
	assert.Equal(t, 0, len(results))
}

// ======================== StartWriteMetrics Tests ========================

func TestStartWriteMetrics_Disabled(t *testing.T) {
	CloudConf.PromWriteConf.Enabled = false
	// 不会 panic，直接返回
	StartWriteMetrics()
}

func TestStartWriteMetrics_Enabled_NoProducts(t *testing.T) {
	CloudConf.PromWriteConf.Enabled = true
	CloudConf.PromWriteConf.Products = []PromWriteProductConfig{}

	// 不会 panic
	StartWriteMetrics()
}

// ======================== StopWriteMetrics Tests ========================

func TestStopWriteMetrics(t *testing.T) {
	// 重置 stopSignals
	oldStopSignals := stopSignals
	stopSignals = make(chan struct{})
	defer func() { stopSignals = oldStopSignals }()

	// 启动一个 goroutine 监听信号
	done := make(chan struct{})
	go func() {
		<-stopSignals
		close(done)
	}()

	StopWriteMetrics()

	select {
	case <-done:
		// 成功收到信号
	case <-time.After(time.Second):
		t.Fatal("StopWriteMetrics did not close stopSignals")
	}
}

// ======================== batchQueryMetricData Tests ========================

func TestBatchQueryMetricData(t *testing.T) {
	initTestLogger()
	metrics := []model.MetricInfo{
		{Namespace: "SYS.ECS", MetricName: "cpu_util", Dimensions: []model.MetricsDimension{{Name: "instance_id", Value: "i-001"}}},
	}

	avg := 50.0
	mockResp := &model.BatchListMetricDataResponse{
		Metrics: &[]model.BatchMetricData{
			{
				MetricName: "cpu_util",
				Datapoints: []model.DatapointForBatchMetric{
					{Average: &avg, Timestamp: time.Now().UnixMilli()},
				},
			},
		},
	}

	patches := gomonkey.ApplyFunc(getCESClient, func() *ces.CesClient {
		return nil
	})
	defer patches.Reset()

	patches.ApplyMethodFunc(&ces.CesClient{}, "BatchListMetricData",
		func(request *model.BatchListMetricDataRequest) (*model.BatchListMetricDataResponse, error) {
			return mockResp, nil
		})

	result, err := batchQueryMetricData(&metrics, time.Now().Add(-10*time.Minute).UnixMilli(), time.Now().UnixMilli())
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, len(*result))
}

func TestBatchQueryMetricData_APIError(t *testing.T) {
	initTestLogger()
	metrics := []model.MetricInfo{
		{Namespace: "SYS.ECS", MetricName: "cpu_util", Dimensions: []model.MetricsDimension{{Name: "instance_id", Value: "i-001"}}},
	}

	patches := gomonkey.ApplyFunc(getCESClient, func() *ces.CesClient {
		return nil
	})
	defer patches.Reset()

	patches.ApplyMethodFunc(&ces.CesClient{}, "BatchListMetricData",
		func(request *model.BatchListMetricDataRequest) (*model.BatchListMetricDataResponse, error) {
			return nil, fmt.Errorf("API error")
		})

	result, err := batchQueryMetricData(&metrics, 0, 0)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// ======================== AgentDimensionsValue Tests ========================

func TestAgentDimensionsValue_TTL(t *testing.T) {
	val := AgentDimensionsValue{
		originValue: "/dev/vda1",
		ttl:         time.Now().Add(time.Hour).Unix(),
	}
	assert.True(t, val.ttl > time.Now().Unix())
}

func TestAgentDimensionsValue_Expired(t *testing.T) {
	val := AgentDimensionsValue{
		originValue: "/dev/vda1",
		ttl:         time.Now().Add(-time.Hour).Unix(),
	}
	assert.True(t, val.ttl < time.Now().Unix())
}

// ======================== FilterMetricsCacheItem Tests ========================

func TestFilterMetricsCacheItem_Valid(t *testing.T) {
	item := FilterMetricsCacheItem{
		TTL: time.Now().Add(time.Hour).Unix(),
		FilterMetrics: []model.MetricInfoList{
			{Namespace: "SYS.ECS", MetricName: "cpu_util"},
		},
	}
	assert.True(t, item.TTL > time.Now().Unix())
	assert.Equal(t, 1, len(item.FilterMetrics))
}

func TestFilterMetricsCacheItem_Expired(t *testing.T) {
	item := FilterMetricsCacheItem{
		TTL:           time.Now().Add(-time.Hour).Unix(),
		FilterMetrics: []model.MetricInfoList{},
	}
	assert.True(t, item.TTL < time.Now().Unix())
}

// ======================== writeMetricsToPrometheus Proto Integration ========================

func TestWriteMetricsToPrometheus_ProtoMarshalIntegration(t *testing.T) {
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		receivedBody, err = io.ReadAll(r.Body)
		assert.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	CloudConf.PromWriteConf.Url = server.URL

	series := []*TimeSeries{
		{
			Labels: []*Label{
				{Name: "__name__", Value: "huaweicloud_sys_ecs_cpu_util"},
				{Name: "instance_id", Value: "i-abc123"},
				{Name: "az", Value: "cn-north-4a"},
			},
			Samples: []*Sample{{Value: 85.5, Timestamp: 1700000000000}},
		},
	}
	writeMetricsToPrometheus(series)

	// snappy 解压
	decompressed, err := snappy.Decode(nil, receivedBody)
	assert.NoError(t, err)

	req := &WriteRequest{}
	err = proto.Unmarshal(decompressed, req)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(req.Timeseries))
	assert.Equal(t, 3, len(req.Timeseries[0].Labels))
	assert.Equal(t, "huaweicloud_sys_ecs_cpu_util", req.Timeseries[0].Labels[0].Value)
	assert.Equal(t, "i-abc123", req.Timeseries[0].Labels[1].Value)
	assert.Equal(t, 85.5, req.Timeseries[0].Samples[0].Value)
	assert.Equal(t, int64(1700000000000), req.Timeseries[0].Samples[0].Timestamp)
}

func TestWriteMetricsToPrometheus_3xxStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMovedPermanently)
		_, err := w.Write([]byte("redirect"))
		if err != nil {
			assert.NotNil(t, err)
		}
	}))
	defer server.Close()

	CloudConf.PromWriteConf.Url = server.URL

	series := []*TimeSeries{
		{Labels: []*Label{{Name: "__name__", Value: "test"}}, Samples: []*Sample{{Value: 1.0, Timestamp: 100}}},
	}
	// 3xx 状态码会被当作非 2xx 处理，只记录日志不 panic
	writeMetricsToPrometheus(series)
}

// ======================== Helper ========================

func floatPtr(f float64) *float64 {
	return &f
}
