package collector

import (
	"fmt"
	"sync"
	"time"

	http_client "github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	eps "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eps/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eps/v1/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/eps/v1/region"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var epsInfo = &EpsInfo{
	EpDetails: make([]model.EpDetail, 0),
	EpMap:     sync.Map{},
}

type EpsInfo struct {
	EpDetails []model.EpDetail
	EpMap     sync.Map
	TTL       int64
	sync.Mutex
}

const HelpInfo = `# HELP huaweicloud_epinfo huaweicloud_epinfo
# TYPE huaweicloud_epinfo gauge
`

func getEPSClient() *eps.EpsClient {
	return eps.NewEpsClient(getEPSClientBuilder().Build())
}

func getEPSClientBuilder() *http_client.HcHttpClientBuilder {
	builder := eps.EpsClientBuilder().WithCredential(authCredentialMap[conf.AuthMode](GlobalServiceType)).
		WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify))
	if endpoint, ok := endpointConfig["eps"]; ok {
		builder.WithEndpoint(endpoint)
	} else {
		builder.WithRegion(region.ValueOf("cn-north-4"))
	}
	return builder
}

func GetEPSInfo() (string, error) {
	result := HelpInfo
	epsInfo, err := listEps()
	if err != nil {
		return result, err
	}
	for _, detail := range epsInfo {
		result += fmt.Sprintf("%s_epinfo{epId=\"%s\",epName=\"%s\"} 1\n", CloudConf.Global.Prefix, detail.Id, detail.Name)
	}
	return result, nil
}

func listEps() ([]model.EpDetail, error) {
	if epsInfo != nil && time.Now().Unix() < epsInfo.TTL {
		return epsInfo.EpDetails, nil
	}
	epsInfo.Lock()
	defer epsInfo.Unlock()

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
		if len(resources) > MaxEpsCount {
			logs.Logger.Errorf("eps not allowed to exceed 10000")
			break
		}
		*req.Offset += limit
	}
	// 清空 epsInfo.EpMap 的所有键值对
	epsInfo.EpMap.Range(func(key, value interface{}) bool {
		epsInfo.EpMap.Delete(key) // 删除当前遍历到的键
		return true               // 返回 true 以继续遍历下一个键值对
	})
	for _, resource := range resources {
		epsInfo.EpMap.Store(resource.Id, resource.Name)
	}
	epsInfo.EpDetails = resources
	epsInfo.TTL = time.Now().Add(GetResourceInfoExpirationTime()).Unix()
	return epsInfo.EpDetails, nil
}

var once sync.Once

func GetEpNameByEpId(epId string) string {
	once.Do(func() {
		_, err := listEps()
		if err != nil {
			logs.Logger.Errorf("Get enterprise project name by enterprise project id(%s) error: %s", epId, err.Error())
		}
	})
	epName, ok := epsInfo.EpMap.Load(epId)
	if !ok {
		return ""
	}
	result, ok := epName.(string)
	if !ok {
		return ""
	}
	return result
}
