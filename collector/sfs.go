package collector

import (
	"errors"
	"time"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

var sfsInfo serversInfo

type SFSInfo struct{}

func (getter SFSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	sfsInfo.Lock()
	defer sfsInfo.Unlock()
	if sfsInfo.LabelInfo == nil || time.Now().Unix() > sfsInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.SFS")
		metricNames, ok := sysConfigMap["share_id"]
		if !ok {
			logs.Logger.Warnf("metric config is empty of share_id")
			return sfsInfo.LabelInfo, sfsInfo.FilterMetrics
		}
		shares, err := getAllShareInfo()
		if err != nil {
			logs.Logger.Errorf("Get all sfs share: %s", err.Error())
			return sfsInfo.LabelInfo, sfsInfo.FilterMetrics
		}
		for _, share := range shares {
			metrics := buildSingleDimensionMetrics(metricNames, "SYS.SFS", "share_id", share.ID)
			filterMetrics = append(filterMetrics, metrics...)
			info := labelInfo{
				Name:  []string{"name"},
				Value: []string{share.Name},
			}
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
		}

		sfsInfo.LabelInfo = resourceInfos
		sfsInfo.FilterMetrics = filterMetrics
		sfsInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return sfsInfo.LabelInfo, sfsInfo.FilterMetrics
}

type ListAllShareRequest struct {
	Offset *int32 `json:"offset,omitempty"`
	Limit  *int32 `json:"limit,omitempty"`
}

type ListAllShareResponse struct {
	Count          string  `json:"count"`
	Shares         []Share `json:"shares"`
	HttpStatusCode int     `json:"-"`
}

type Share struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func getAllShareInfo() ([]ResourceBaseInfo, error) {
	var shares []ResourceBaseInfo
	limit := int32(1000)
	offset := int32(0)
	request := &ListAllShareRequest{Limit: &limit, Offset: &offset}
	requestDef := genDefaultReqDefWithOffsetAndLimit("/v2/{project_id}/shares", new(ListAllShareResponse))
	for {
		resp, err := getHcClient(getEndpoint("sfs", "v2")).Sync(request, requestDef)
		if err != nil {
			return nil, err
		}
		response, ok := resp.(*ListAllShareResponse)
		if !ok {
			return nil, errors.New("resp type is not ListAllShareResponse")
		}
		if len(response.Shares) == 0 {
			break
		}
		for _, share := range response.Shares {
			shares = append(shares, ResourceBaseInfo{ID: share.ID, Name: share.Name})
		}
		*request.Offset += limit
	}
	return shares, nil
}
