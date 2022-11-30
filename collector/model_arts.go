package collector

import (
	"errors"
	"net/http"
	"time"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

var modelArtsInfo serversInfo

type ModelArtsInfo struct{}

func (getter ModelArtsInfo) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]cesmodel.MetricInfoList, 0)
	modelArtsInfo.Lock()
	defer modelArtsInfo.Unlock()
	if modelArtsInfo.LabelInfo == nil || time.Now().Unix() > modelArtsInfo.TTL {
		services, err := getModelArtsServices()
		if err != nil {
			logs.Logger.Error("Get all ModelArts services error:", err.Error())
			return modelArtsInfo.LabelInfo, modelArtsInfo.FilterMetrics
		}
		sysConfigMap := getMetricConfigMap("SYS.ModelArts")
		metricNames := sysConfigMap["service_id"]
		modelMetricNames := sysConfigMap["service_id,model_id"]
		if len(metricNames) == 0 {
			logs.Logger.Warn("Metric config is empty of SYS.ModelArts, dim_metric_name is service_id")
			return modelArtsInfo.LabelInfo, modelArtsInfo.FilterMetrics
		}
		for _, service := range services {
			metrics := buildSingleDimensionMetrics(metricNames, "SYS.ModelArts", "service_id", service.ID)
			filterMetrics = append(filterMetrics, metrics...)
			info := labelInfo{
				Name:  []string{"name", "epId"},
				Value: []string{service.Name, service.EpId},
			}
			keys, values := getTags(service.Tags)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
			if len(modelMetricNames) == 0 {
				continue
			}
			for _, model := range service.Models {
				metrics := buildDimensionMetrics(modelMetricNames, "SYS.ModelArts",
					[]cesmodel.MetricsDimension{{Name: "service_id", Value: service.ID}, {Name: "model_id", Value: model.ModelID}})
				filterMetrics = append(filterMetrics, metrics...)
				modelInfo := labelInfo{
					Name:  []string{"modelName"},
					Value: []string{model.ModelName},
				}
				modelInfo.Name = append(modelInfo.Name, info.Name...)
				modelInfo.Value = append(modelInfo.Value, info.Value...)
				resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = modelInfo
			}
		}
		modelArtsInfo.LabelInfo = resourceInfos
		modelArtsInfo.FilterMetrics = filterMetrics
		modelArtsInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return modelArtsInfo.LabelInfo, modelArtsInfo.FilterMetrics
}

type ShowServiceRequest struct {
	ServiceId string `json:"service_id"`
}

type Model struct {
	ModelID   string `json:"model_id"`
	ModelName string `json:"model_name"`
}

type ServiceDetail struct {
	Config         []Model `json:"config"`
	HttpStatusCode int     `json:"-"`
}

func genReqDefForShowService() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().WithMethod(http.MethodGet).WithPath("/v1/{project_id}/services/{service_id}").
		WithResponse(new(ServiceDetail)).WithContentType("application/json")

	reqDefBuilder.WithRequestField(def.NewFieldDef().WithName("ServiceId").WithJsonTag("service_id").WithLocationType(def.Path))
	return reqDefBuilder.Build()
}

func showService(serviceId string) (*ServiceDetail, error) {
	request := &ShowServiceRequest{ServiceId: serviceId}
	resp, err := getHcClient(getEndpoint("modelarts", "v1")).Sync(request, genReqDefForShowService())
	if err != nil {
		return nil, err
	}
	response, ok := resp.(*ServiceDetail)
	if !ok {
		return nil, errors.New("resp type is not ServiceDetail")
	}
	return response, nil
}

type ModelArtService struct {
	ResourceBaseInfo
	Models []Model
}

func getModelArtsServices() ([]ModelArtService, error) {
	services, err := listResources("modelarts", "service")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of modelarts.service, error: %s", err.Error())
		return nil, err
	}
	servicesInfo := make([]ModelArtService, 0, len(services))
	for _, service := range services {
		tmpService := ModelArtService{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   *service.Id,
				Name: *service.Name,
				EpId: *service.EpId,
				Tags: service.Tags},
		}
		serviceInfo, err := showService(*service.Id)
		if err != nil {
			logs.Logger.Errorf("Failed to get model of service of %s, error: %s", *service.Id, err.Error())
		}
		tmpService.Models = serviceInfo.Config
		servicesInfo = append(servicesInfo, tmpService)
	}
	return servicesInfo, nil
}
