package collector

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	v1 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1/region"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

type RmsErrorResp struct {
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

const (
	// ResourceTypeErrorCode 查询RMS侧指定类型的资源，资源类型参数校验不通过，会返回该错误码
	// 欧洲站等特殊局点与华为云大网有区别，部分云服务的指定资源（比如DCS的memcached资源类型），大网支持，但此类特殊局点不支持，强制查询会返回该错误码
	ResourceTypeErrorCode = "RMS.00010002"
)

func getRMSClient() *v1.RmsClient {
	return v1.NewRmsClient(getRMSClientBuilder().Build())
}

func getRMSClientBuilder() *http_client.HcHttpClientBuilder {
	builder := v1.RmsClientBuilder().
		WithCredential(authCredentialMap[conf.AuthMode](GlobalServiceType)).
		WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify))
	if endpoint, ok := endpointConfig["rms"]; ok {
		builder.WithEndpoint(endpoint)
	} else {
		builder.WithRegion(region.ValueOf("cn-north-4"))
	}
	return builder
}

func listResources(provider, resourceType string, optionalRegionID ...string) ([]model.ResourceEntity, error) {
	var result []model.ResourceEntity
	var err error
	retryTimes := 0
	for retryTimes < CloudConf.Global.RmsRetryTimes {
		result, err = listResourceNoRetry(provider, resourceType, optionalRegionID...)
		if err == nil {
			break
		}
		logs.Logger.Errorf("List resource error: %s", err.Error())
		retryTimes += 1
	}
	return result, err
}

func listResourceNoRetry(provider, resourceType string, optionalRegionID ...string) ([]model.ResourceEntity, error) {
	regionID := conf.Region
	if len(optionalRegionID) > 0 {
		regionID = optionalRegionID[0]
	}
	limit := int32(200)
	var resources []model.ResourceEntity
	req := &model.ListResourcesRequest{
		Provider: provider,
		Type:     resourceType,
		RegionId: &regionID,
		Limit:    &limit,
	}
	if CloudConf.Global.EpIds != "" {
		epIdArr := strings.Split(CloudConf.Global.EpIds, ",")
		for _, epID := range epIdArr {
			req.EpId = &epID
			resourceByEpID, err := getResourcesFromRMS(req)
			if err == nil {
				resources = append(resources, resourceByEpID...)
				continue
			}
			logs.Logger.Errorf("Get resources from rms by epID failed, epID is %s, error: %s", epID, err.Error())
			if errCode := getRmsErrorCode(err); errCode == ResourceTypeErrorCode {
				return []model.ResourceEntity{}, nil
			}
			return nil, errors.New(fmt.Sprintf("get resources from rms by epID failed: %s", err.Error()))
		}
		return resources, nil
	} else {
		resources, err := getResourcesFromRMS(req)
		if err == nil {
			return resources, nil
		}
		logs.Logger.Errorf("Get resources from rms failed, error: %s", err.Error())
		if errCode := getRmsErrorCode(err); errCode == ResourceTypeErrorCode {
			return []model.ResourceEntity{}, nil
		}
		return nil, errors.New(fmt.Sprintf("get resources from rms failed: %s", err.Error()))
	}
}

// 判断是否因为当前局点rms不支持当前资源类型，而导致查询失败（如果是，则忽略该错误，返回空的资源列表）
func getRmsErrorCode(err error) string {
	errorResp := RmsErrorResp{}
	unmarshalErr := json.Unmarshal([]byte(err.Error()), &errorResp)
	if unmarshalErr != nil {
		return ""
	}
	return errorResp.ErrorCode
}

func getResourcesFromRMS(req *model.ListResourcesRequest) ([]model.ResourceEntity, error) {
	var resources []model.ResourceEntity
	// 多EPID下，保证每个EPID查询开始时，Marker都是初始值
	req.Marker = nil
	for {
		response, err := getRMSClient().ListResources(req)
		if err != nil {
			return resources, err
		}
		resources = append(resources, *response.Resources...)
		if response.PageInfo.NextMarker == nil {
			break
		}
		req.Marker = response.PageInfo.NextMarker
	}

	if len(resources) == 0 {
		logs.Logger.Infof("Get empty resource list from rms")
	}

	return resources, nil
}
