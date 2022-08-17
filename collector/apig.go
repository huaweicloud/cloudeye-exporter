package collector

import (
	"errors"
	"net/http"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	apig "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/apig/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var apigInfo serversInfo

type APIGInfo struct{}

func (getter APIGInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]model.MetricInfoList, 0)
	apigInfo.Lock()
	defer apigInfo.Unlock()
	if apigInfo.LabelInfo == nil || time.Now().Unix() > apigInfo.TTL {
		sysConfigMap := getMetricConfigMap("SYS.APIG")
		metricNames, ok := sysConfigMap["api_id"]
		if !ok {
			return apigInfo.LabelInfo, apigInfo.FilterMetrics
		}
		apps, err := getAllAPIGAppsInstances()
		if err != nil {
			logs.Logger.Errorf("Failed to get all apig apps, error: %s", err.Error())
			return apigInfo.LabelInfo, apigInfo.FilterMetrics
		}
		for _, app := range apps {
			metrics := buildSingleDimensionMetrics(metricNames, "SYS.APIG", "api_id", app.ID)
			filterMetrics = append(filterMetrics, metrics...)
			info := labelInfo{
				Name:  []string{"name", "creator"},
				Value: []string{app.Name, app.Creator},
			}
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
		}
		apigInfo.LabelInfo = resourceInfos
		apigInfo.FilterMetrics = filterMetrics
		apigInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return apigInfo.LabelInfo, apigInfo.FilterMetrics
}

type ListAppsRequest struct {
	Offset *int32 `json:"offset,omitempty"`
	Limit  *int32 `json:"limit,omitempty"`
}

type ListAppsResponse struct {
	Total          *int32  `json:"total,omitempty"`
	Size           *int32  `json:"size,omitempty"`
	Apps           *[]Apps `json:"apps,omitempty"`
	HttpStatusCode int     `json:"-"`
}

type Apps struct {
	BindNum      int       `json:"bind_num"`
	Creator      string    `json:"creator"`
	UpdateTime   time.Time `json:"update_time"`
	AppKey       string    `json:"app_key"`
	Name         string    `json:"name"`
	Remark       string    `json:"remark"`
	ID           string    `json:"id"`
	AppSecret    string    `json:"app_secret"`
	RegisterTime time.Time `json:"register_time"`
	Status       int       `json:"status"`
}

func genReqDefForListApps() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().WithMethod(http.MethodGet).WithPath("/v1.0/apigw/apps").
		WithResponse(new(ListAppsResponse)).WithContentType("application/json")

	reqDefBuilder.WithRequestField(def.NewFieldDef().WithName("Offset").WithJsonTag("offset").WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().WithName("Limit").WithJsonTag("limit").WithLocationType(def.Query))

	return reqDefBuilder.Build()
}

type APIGAppsInfo struct {
	ResourceBaseInfo
	Creator string
}

func getAllAPIGAppsInstances() ([]APIGAppsInfo, error) {
	var apps []APIGAppsInfo
	limit := int32(200)
	offset := int32(0)
	request := &ListAppsRequest{Limit: &limit, Offset: &offset}
	requestDef := genReqDefForListApps()
	for {
		resp, err := getAPIGSClient().HcClient.Sync(request, requestDef)
		if err != nil {
			return nil, err
		}
		appsInfo, ok := resp.(*ListAppsResponse)
		if !ok {
			return nil, errors.New("resp type is not ListAppsResponse")
		}
		if len(*appsInfo.Apps) == 0 {
			break
		}
		for _, app := range *appsInfo.Apps {
			apps = append(apps, APIGAppsInfo{
				ResourceBaseInfo: ResourceBaseInfo{ID: app.ID, Name: app.Name},
				Creator:          app.Creator,
			})
		}
		*request.Offset += 1
	}
	return apps, nil
}

func getAPIGSClient() *apig.ApigClient {
	return apig.NewApigClient(apig.ApigClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(true)).
		WithEndpoint(getEndpoint("apig", "v1.0")).Build())
}
