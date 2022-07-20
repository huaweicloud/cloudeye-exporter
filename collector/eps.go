package collector

import (
	"fmt"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	eps "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eps/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eps/v1/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eps/v1/region"
)

var epsInfo = &EpsInfo{
	EpDetails: make([]model.EpDetail, 0),
}

type EpsInfo struct {
	EpDetails []model.EpDetail
	TTL       int64
}

const HelpInfo = `# HELP huaweicloud_epinfo huaweicloud_epinfo
# TYPE huaweicloud_epinfo gauge
`

func getEPSClient() *eps.EpsClient {
	return eps.NewEpsClient(eps.EpsClientBuilder().WithRegion(region.ValueOf("cn-north-4")).
		WithCredential(global.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).Build()).Build())
}

func GetEPSInfo() (string, error) {
	result := HelpInfo
	epsInfo, err := listEps()
	if err != nil {
		return result, err
	}
	for _, detail := range epsInfo {
		result += fmt.Sprintf("huaweicloud_epinfo{epId=\"%s\",epName=\"%s\"} 1\n", detail.Id, detail.Name)
	}
	return result, nil
}

func listEps() ([]model.EpDetail, error) {
	if epsInfo != nil && time.Now().Unix() < epsInfo.TTL {
		return epsInfo.EpDetails, nil
	}

	limit := int32(1000)
	Offset := int32(0)
	req := &model.ListEnterpriseProjectRequest{
		Limit:  &limit,
		Offset: &Offset,
	}

	client := getEPSClient()
	var resources []model.EpDetail

	for {
		response, err := client.ListEnterpriseProject(req)
		if err != nil {
			return resources, err
		}
		resources = append(resources, *response.EnterpriseProjects...)
		if len(*response.EnterpriseProjects) == 0 {
			break
		}
		Offset += limit
		req.Offset = &Offset
	}
	epsInfo.EpDetails = resources
	epsInfo.TTL = time.Now().Add(TTL).Unix()
	return epsInfo.EpDetails, nil
}
