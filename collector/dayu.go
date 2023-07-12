package collector

import (
	"errors"
	"net/http"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var dayuInfo serversInfo

type DayuInfo struct{}

func (getter DayuInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	dayuInfo.Lock()
	defer dayuInfo.Unlock()
	if dayuInfo.LabelInfo == nil || time.Now().Unix() > dayuInfo.TTL {
		streams, err := getAllStreams()
		if err != nil {
			logs.Logger.Error("Get all dis Streams error:", err.Error())
			return dayuInfo.LabelInfo, dayuInfo.FilterMetrics
		}

		sysConfigMap := getMetricConfigMap("SYS.DAYU")
		for _, stream := range streams {
			if metricNames, ok := sysConfigMap["stream_id"]; ok {
				metrics := buildSingleDimensionMetrics(metricNames, "SYS.DAYU", "stream_id", stream.ID)
				filterMetrics = append(filterMetrics, metrics...)
				info := labelInfo{
					Name:  []string{"name", "epId"},
					Value: []string{stream.Name, stream.EpId},
				}
				keys, values := getTags(stream.Tags)
				info.Name = append(info.Name, keys...)
				info.Value = append(info.Value, values...)
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			}
		}

		dayuInfo.LabelInfo = resourceInfos
		dayuInfo.FilterMetrics = filterMetrics
		dayuInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return dayuInfo.LabelInfo, dayuInfo.FilterMetrics
}

type StreamInfo struct {
	Private                    bool   `json:"private"`
	StreamID                   string `json:"stream_id"`
	StreamName                 string `json:"stream_name"`
	CreateTime                 int64  `json:"create_time"`
	RetentionPeriod            int    `json:"retention_period"`
	Status                     string `json:"status"`
	StreamType                 string `json:"stream_type"`
	DataType                   string `json:"data_type"`
	PartitionCount             int    `json:"partition_count"`
	Tags                       []Tag  `json:"tags"`
	SysTags                    []Tag  `json:"sys_tags"`
	AutoScaleEnabled           bool   `json:"auto_scale_enabled"`
	AutoScaleMinPartitionCount int    `json:"auto_scale_min_partition_count"`
	AutoScaleMaxPartitionCount int    `json:"auto_scale_max_partition_count"`
}

type ListStreamsResp struct {
	TotalNumber    int          `json:"total_number"`
	StreamNames    []string     `json:"stream_names"`
	StreamInfoList []StreamInfo `json:"stream_info_list"`
	HasMoreStreams bool         `json:"has_more_streams"`
	HttpStatusCode int          `json:"-"`
}

type ListStreamsRep struct {
	Limit           string `json:"limit"`
	StartStreamName string `json:"start_stream_name"`
}

func genReqDefForListStreams() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().WithMethod(http.MethodGet).WithPath("/v2/{project_id}/streams").
		WithResponse(new(ListStreamsResp)).WithContentType("application/json")

	reqDefBuilder.WithRequestField(def.NewFieldDef().WithName("Limit").WithJsonTag("limit").WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().WithName("StartStreamName").WithJsonTag("start_stream_name").WithLocationType(def.Query))
	return reqDefBuilder.Build()
}

func listStreams() ([]StreamInfo, error) {
	disClient := getHcClient(getEndpoint("dis", "v2"))
	request := &ListStreamsRep{Limit: "100"}
	var streams []StreamInfo
	for {
		resp, err := disClient.Sync(request, genReqDefForListStreams())
		if err != nil {
			logs.Logger.Errorf("Failed to get list streams : %s", err.Error())
			return nil, err
		}
		response, ok := resp.(*ListStreamsResp)
		if !ok {
			err := errors.New("resp type is not ServiceDetail")
			logs.Logger.Errorf("Failed to get list streams : %s", err.Error())
			return nil, err
		}
		streams = append(streams, response.StreamInfoList...)
		if !response.HasMoreStreams {
			break
		}
		request.StartStreamName = response.StreamNames[len(response.StreamNames)-1]
	}
	return streams, nil
}

func getAllStreams() ([]ResourceBaseInfo, error) {
	streams, err := listStreams()
	if err != nil {
		logs.Logger.Errorf("Failed to get list streams : %s", err.Error())
		return nil, err
	}

	resources := make([]ResourceBaseInfo, len(streams))
	for i, stream := range streams {
		resources[i].ID = stream.StreamID
		resources[i].Name = stream.StreamName
		resources[i].EpId = fmtTags(stream.SysTags)["_sys_enterprise_project_id"]
		resources[i].Tags = fmtTags(stream.Tags)
	}
	return resources, nil
}
