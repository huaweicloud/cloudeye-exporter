package collector

import (
	"testing"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	"github.com/stretchr/testify/assert"
)

func TestWanQMonitorGetResourceInfo(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()
	conf.DomainID = "123"

	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	conf.Region = "cn-test-01"
	metricConf = map[string]MetricConf{
		"SYS.WANQMonitor": {
			DimMetricName: map[string][]string{
				"task_id":               {"packet_loss_rate_avg"},
				"task_id,city":          {"packet_loss_rate_avg"},
				"task_id,operator":      {"packet_loss_rate_avg"},
				"task_id,operator,city": {"packet_loss_rate_avg"},
				"task_id,probe_id":      {"packet_loss_rate_avg"},
			},
		},
	}
	resp := QualityMonitorTasksResponse{
		Tasks: []Task{
			{
				TaskId:   "123123",
				TaskName: "test-name",
				EpID:     "0",
				Tags: []Tag{
					{
						Key:   "k",
						Value: "v",
					},
				},
				ProbeIds: []string{"1", "3"},
			},
		},
		Count:          1,
		HttpStatusCode: 200,
	}
	probeResp := ListProbeResponse{
		Probes: []Probe{
			{
				ProbeID:   "1",
				Operator:  "CMCC",
				Continent: "asia",
				Country:   "China",
				Province:  "zhejiang",
				City:      "hangzhou",
				Label: ProbeLabel{
					EnUs: LabelInfo{
						Name:     "china-zhejiang-hangzhou-CMCC",
						Country:  "china",
						City:     "hangzhou",
						Operator: "CMCC",
						Province: "zhejiang",
					},
					ZhCN: LabelInfo{
						Name:     "中国-浙江-杭州-中国移动",
						Country:  "中国",
						City:     "杭州",
						Operator: "中国移动",
						Province: "浙江",
					},
				},
			},
		},
		Count:          1,
		HttpStatusCode: 200,
	}
	patches.ApplyMethodFunc(getHcClient(getEndpoint("ces", "v2")), "Sync", func(req interface{}, reqDef *def.HttpRequestDef) (interface{}, error) {
		if reqDef.Path == "/v2/123/quality-monitor-tasks" {
			return &resp, nil
		}
		if reqDef.Path == "/v2/123/quality-monitor-probes" {
			return &probeResp, nil
		}
		return nil, nil
	})

	metrics := []model.MetricInfoList{
		{
			Namespace:  "SYS.WANQMonitor",
			MetricName: "packet_loss_rate_avg",
			Unit:       "%",
			Dimensions: []model.MetricsDimension{
				{
					Name:  "task_id",
					Value: "123123",
				},
			},
		},
		{
			Namespace:  "SYS.WANQMonitor",
			MetricName: "packet_loss_rate_avg",
			Unit:       "%",
			Dimensions: []model.MetricsDimension{
				{
					Name:  "task_id",
					Value: "123123",
				},
				{
					Name:  "city",
					Value: "hangzhou",
				},
			},
		},
		{
			Namespace:  "SYS.WANQMonitor",
			MetricName: "packet_loss_rate_avg",
			Unit:       "%",
			Dimensions: []model.MetricsDimension{
				{
					Name:  "task_id",
					Value: "123123",
				},
				{
					Name:  "operator",
					Value: "CMCC",
				},
			},
		},
		{
			Namespace:  "SYS.WANQMonitor",
			MetricName: "packet_loss_rate_avg",
			Unit:       "%",
			Dimensions: []model.MetricsDimension{
				{
					Name:  "task_id",
					Value: "123123",
				},
				{
					Name:  "operator",
					Value: "CMCC",
				},
				{
					Name:  "city",
					Value: "hangzhou",
				},
			},
		},
		{
			Namespace:  "SYS.WANQMonitor",
			MetricName: "packet_loss_rate_avg",
			Unit:       "%",
			Dimensions: []model.MetricsDimension{
				{
					Name:  "task_id",
					Value: "123123",
				},
				{
					Name:  "probe_id",
					Value: "1",
				},
			},
		},
	}
	patches.ApplyFuncReturn(listAllMetrics, metrics, nil)

	var wanQMonitor WanMonitorInfo
	resourceInfo, metrics := wanQMonitor.GetResourceInfo()
	assert.Equal(t, 5, len(resourceInfo))
	assert.Equal(t, 5, len(metrics))

}
