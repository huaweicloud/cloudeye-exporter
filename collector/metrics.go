package collector

import (
    "bytes"
    "context"
    "crypto/tls"
    "fmt"
    "io"
    "net"
    "net/http"
    "sync"
    "time"

    "github.com/golang/snappy"
    ces "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1"
    "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
    cesv2 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v2"
    cesv2model "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v2/model"
    "github.com/prometheus/client_golang/prometheus"
    "google.golang.org/protobuf/proto"

    "github.com/huaweicloud/cloudeye-exporter/logs"
)

type FilterMetricsCacheItem struct {
    TTL           int64
    FilterMetrics []model.MetricInfoList
}

var (
    host            string
    agentDimensions = sync.Map{}

    writeHttpClient *http.Client

    filterMetricsCache = sync.Map{}
)

const (
    ExporterWritePrometheusBatchSize = 5000
    BatchQueryMetricData429RetryTimes = 3
)

var stopSignals = make(chan struct{})

func init() {
    writeHttpClient = &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: true,
                MinVersion:         tls.VersionTLS12,
            },
            DialContext: (&net.Dialer{
                KeepAlive: 10 * time.Second,
            }).DialContext,
            MaxIdleConnsPerHost: 10,
        },
        Timeout: 10 * time.Second,
    }
}

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
    var resp *model.BatchListMetricDataResponse
    var err error
    retryTimes := 0
    for retryTimes < BatchQueryMetricData429RetryTimes {
        resp, err = getCESClient().BatchListMetricData(options)
        if err == nil {
            break
        }
        logs.Logger.Errorf("Failed to get metric data: %s", err.Error())
        if !isErrorTypeForTooManyRequests(err) {
            break
        }
        time.Sleep(time.Second)
        retryTimes++
    }
    if err != nil {
        return nil, err
    }
    if resp == nil {
        return nil, fmt.Errorf("batch query metric data response is nil")
    }
    if resp.Metrics == nil {
        return &[]model.BatchMetricData{}, nil
    }
    return resp.Metrics, nil
}

func listAllMetrics(namespace string) ([]model.MetricInfoList, error) {
    if value, loadOK := filterMetricsCache.Load(namespace); loadOK {
        cachedFilterMetricsItem, ok := value.(FilterMetricsCacheItem)
        if ok && cachedFilterMetricsItem.TTL > time.Now().Unix() {
            return cachedFilterMetricsItem.FilterMetrics, nil
        }
    }

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
    filterMetricsCache.Store(namespace, FilterMetricsCacheItem{
        FilterMetrics: metricData,
        TTL:           time.Now().Add(GetResourceInfoExpirationTime()).Unix(),
    })
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
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
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
            case <-stopSignals:
                logs.Logger.Info("Stopping agent dimensions cache cleanup")
                return
            }
        }
    }()
}

func StartWriteMetrics() {
    if !CloudConf.PromWriteConf.Enabled {
        logs.Logger.Info("Job for pushing metrics to prometheus is not enabled")
        return
    }

    for _, product := range CloudConf.PromWriteConf.Products {
        queryDuration := fmt.Sprintf("-%dm", CloudConf.Global.MetricQueryDuration)
        duration, err := time.ParseDuration(queryDuration)
        if err != nil {
            logs.Logger.Errorf("ParseDuration %s error: %s", queryDuration, err.Error())
            logs.FlushLogAndExit(1)
        }
        currentProduct := product
        metricsChannel := make(chan *TimeSeries, ExporterWritePrometheusBatchSize)
        go func(productConfig PromWriteProductConfig) {
            ticker := time.NewTicker(1 * time.Minute)
            defer ticker.Stop()
            for {
                select {
                case <-ticker.C:
                    now := time.Now()
                    from := now.Add(duration).UnixNano() / 1e6
                    to := now.UnixNano() / 1e6
                    queryAndWriteMetricData(metricsChannel, productConfig, from, to)
                case <-stopSignals:
                    logs.Logger.Infof("[%s] Stopping metrics writing and send tasks", productConfig.Namespace)
                    close(metricsChannel)
                    return
                }
            }
        }(currentProduct)
        go func(namespace string) {
            var tmpSeries []*TimeSeries
            workChan := make(chan struct{}, 50)
            ticker := time.NewTicker(10 * time.Second)
            defer close(workChan)
            for {
                select {
                case seriesItem, ok := <-metricsChannel:
                    if !ok {
                        logs.Logger.Infof("[%s]Metric channel closed.", namespace)
                        return
                    }
                    tmpSeries = append(tmpSeries, seriesItem)
                    if len(tmpSeries) == ExporterWritePrometheusBatchSize {
                        go func(currentSeries []*TimeSeries) {
                            workChan <- struct{}{}
                            defer func() {
                                <-workChan
                            }()
                            writeMetricsToPrometheus(currentSeries)
                        }(tmpSeries)
                        tmpSeries = []*TimeSeries{}
                    }
                case <-ticker.C:
                    if len(tmpSeries) > 0 {
                        writeMetricsToPrometheus(tmpSeries)
                        tmpSeries = []*TimeSeries{}
                    }
                case <-stopSignals:
                    logs.Logger.Infof("[%s]Stopping write metric data to prometheus.", namespace)
                    return
                }
            }
        }(currentProduct.Namespace)
    }
}

func queryAndWriteMetricData(metricsResChannel chan<- *TimeSeries, product PromWriteProductConfig, from int64, to int64) {
    txnKey := fmt.Sprintf("%s-%d-%d", product.Namespace, from, to)
    logs.Logger.Debugf("[%s]Begin to write metric data to prometheus.", txnKey)
    resourceInfos := map[string]labelInfo{}
    filterMetrics := make([]model.MetricInfoList, 0)
    if serviceFunc, ok := serviceMap[product.Namespace]; ok {
        resourceInfos, filterMetrics = serviceFunc.GetResourceInfo()
    } else {
        var err error
        filterMetrics, err = listAllMetrics(product.Namespace)
        if err != nil {
            logs.Logger.Errorf("[%s] Get all metrics error for write metrics to prometheus: %s, %s", product.Namespace, err.Error())
            return
        }
    }
    if len(filterMetrics) == 0 {
        logs.Logger.Debugf("[%s]End   to write metric data to prometheus. metric count: [%d]", txnKey, 0)
        return
    }
    batchSize := CloudConf.Global.ScrapeBatchSize
    batches := make([][]model.MetricInfo, 0, (len(filterMetrics)+batchSize-1)/batchSize)
    tmpMetrics := make([]model.MetricInfo, 0, batchSize)
    count := 0
    for _, metric := range filterMetrics {
        tmpMetrics = append(tmpMetrics, transMetric(metric))
        count++
        if (len(tmpMetrics) == batchSize) || (count == len(filterMetrics)) {
            batches = append(batches, tmpMetrics)
            tmpMetrics = make([]model.MetricInfo, 0, batchSize)
        }
    }
    if product.MaxRoutines == 0 {
        product.MaxRoutines = 5
    }
    if product.MaxRoutines >= 50 {
        product.MaxRoutines = 50
    }
    workChan := make(chan struct{}, product.MaxRoutines)
    defer close(workChan)
    var wg sync.WaitGroup
    for i := range batches {
        wg.Add(1)
        go func(idx int) {
            workChan <- struct{}{}
            defer func() {
                <-workChan
                wg.Done()
            }()
            dataList, err := batchQueryMetricData(&batches[idx], from, to)
            if err != nil {
                return
            }
            renderAndWriteMetricData(dataList, resourceInfos, metricsResChannel)
        }(i)
    }
    wg.Wait()
    logs.Logger.Debugf("[%s]End to write metric data to prometheus, metric count: [%d].", txnKey)
}

func renderAndWriteMetricData(metrics *[]model.BatchMetricData, resourceInfos map[string]labelInfo, metricsResChannel chan<- *TimeSeries) {
    for _, metric := range *metrics {
        if len(metric.Datapoints) == 0 {
            logs.Logger.Debugf("Metric data point empty: metric_name: %s, dimension: %+v", metric.MetricName, metric.Dimensions)
            continue
        }
        dataPoints := metric.Datapoints
        dataPoint := &dataPoints[len(dataPoints)-1]
        label := getLabel(metric, resourceInfos)
        label = transformUnit(&metric, label)
        if duplicateLabelNames := getDuplicatedLabelNames(label.Name); duplicateLabelNames != "" {
            logs.Logger.Errorf("Duplicate label names were found between custom tags and system tags "+
                "and the metrics export is skipped for this resource(namespace: %s, dimensions info: %v,  labels: %+v).",
                *metric.Namespace, metric.Dimensions, duplicateLabelNames)
            continue
        }
        fqName := prometheus.BuildFQName(CloudConf.Global.Prefix, replaceName(*metric.Namespace), metric.MetricName)
        promPbLabels := []*Label{
            {
                Name:  "__name__",
                Value: fqName,
            },
        }
        promPbLabels = transformCESLabelToPromLabels(promPbLabels, label)
        seriesItem := &TimeSeries{
            Labels: promPbLabels,
            Samples: []*Sample{
                {
                    Value:     *dataPoint.Average,
                    Timestamp: dataPoint.Timestamp,
                },
            },
        }
        metricsResChannel <- seriesItem
    }
}

func writeMetricsToPrometheus(series []*TimeSeries) {
    req := &WriteRequest{
        Timeseries: series,
    }
    data, err := proto.Marshal(req)
    if err != nil {
        logs.Logger.Errorf("Marshal metric data by proto error: %s", err.Error())
        return
    }
    compressed := snappy.Encode(nil, data)
    httpReq, err := http.NewRequestWithContext(
        context.Background(),
        http.MethodPost,
        CloudConf.PromWriteConf.Url,
        bytes.NewReader(compressed),
    )
    if err != nil {
        logs.Logger.Errorf("Build http request for metrics pushing error: %s", err.Error())
        return
    }

    httpReq.Header.Set("Content-Type", "application/x-protobuf")
    httpReq.Header.Set("Content-Encoding", "snappy")
    httpReq.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")
    httpReq.Header.Set("User-Agent", "ces-exporter-go-client/1.0")
    httpResp, err := writeHttpClient.Do(httpReq)
    if err != nil {
        logs.Logger.Errorf("Send http request for metrics pushing error: %s", err.Error())
        return
    }
    defer func() {
        err = httpResp.Body.Close()
        if err != nil {
            logs.Logger.Errorf("Close prometheus response body error: %v", err)
        }
    }()
    if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
        notOkDesc, readErr := io.ReadAll(httpResp.Body)
        if readErr != nil {
            logs.Logger.Errorf("Read prometheus response body not ok: %v", readErr)
            return
        }
        logs.Logger.Errorf("Prometheus returned non-200 status: %s, %s", httpResp.Status, string(notOkDesc))
        return
    }
}

func transformCESLabelToPromLabels(promLabels []*Label, cesLabel labelInfo) []*Label {
    for index, name := range cesLabel.Name {
        promLabels = append(promLabels, &Label{
            Name:  name,
            Value: cesLabel.Value[index],
        })
    }
    return promLabels
}

func StopWriteMetrics() {
    if stopSignals != nil {
        // 发送方将channel关闭后，所有接受方都会收到0值信号
        close(stopSignals)
    }
}
