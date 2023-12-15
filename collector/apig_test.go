package collector

import (
	"errors"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	"github.com/stretchr/testify/assert"
)

func TestApiGGetResourceInfo(t *testing.T) {
	respPage1 := ListAppsResponse{
		HttpStatusCode: 200,
		Apps: &[]Apps{
			{ID: "app-0001", Name: "app1"},
		},
	}
	respPage2 := ListAppsResponse{
		HttpStatusCode: 200,
		Apps:           &[]Apps{},
	}
	sysConfig := map[string][]string{"api_id": {"req_count"}}

	apigClient := getAPIGSClient()
	patches := gomonkey.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	patches.ApplyMethodFunc(apigClient.HcClient, "Sync", func(req interface{}, reqDef *def.HttpRequestDef) (interface{}, error) {
		request, ok := req.(*ListAppsRequest)
		if !ok {
			return nil, errors.New("test error")
		}
		if *request.Offset == 0 {
			return &respPage1, nil
		}
		return &respPage2, nil
	})
	defer patches.Reset()

	var getter APIGInfo
	labels, metrics := getter.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 1, len(metrics))
}
