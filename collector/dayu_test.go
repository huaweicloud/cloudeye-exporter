package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	"github.com/stretchr/testify/assert"
)

func TestDayuGetResourceInfo(t *testing.T) {
	sysConfig := map[string][]string{"stream_id": {"dis11_stream_record_retention_time"}}
	patches := gomonkey.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	dayuClient := &core.HcHttpClient{}
	patches.ApplyFuncReturn(getHcClient, dayuClient)
	resp := ListStreamsResp{
		HttpStatusCode: 200,
		HasMoreStreams: false,
		StreamInfoList: []StreamInfo{
			{StreamID: "stream-0001-0001", StreamName: "Stream01"},
		},
	}
	patches.ApplyMethodFunc(dayuClient, "Sync", func(req interface{}, reqDef *def.HttpRequestDef) (interface{}, error) {
		return &resp, nil
	})
	defer patches.Reset()

	var dayugetter DayuInfo
	labels, metrics := dayugetter.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 1, len(metrics))
}
