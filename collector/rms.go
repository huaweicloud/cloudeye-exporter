package collector

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1/region"
)

func getRMSClient() *v1.RmsClient {
	return v1.NewRmsClient(v1.RmsClientBuilder().WithRegion(region.ValueOf("cn-north-4")).
		WithCredential(global.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).Build()).Build())
}

func listResources(provider, rourceType string) ([]model.ResourceEntity, error) {
	limit := int32(200)
	req := &model.ListResourcesRequest{
		Provider: provider,
		Type:     rourceType,
		RegionId: &conf.Region,
		Limit:    &limit,
	}
	var resources []model.ResourceEntity
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
	return resources, nil
}
