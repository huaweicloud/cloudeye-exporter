package collector

import (
	"fmt"
	"strings"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	mrs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/mrs/v1"
	mrsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/mrs/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var mrsInfo serversInfo

type MRSInfo struct{}

func (getter MRSInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	mrsInfo.Lock()
	defer mrsInfo.Unlock()
	if mrsInfo.LabelInfo == nil {
		mrsInfo.LabelInfo, mrsInfo.FilterMetrics = getMRSResourceAndMetrics()
		mrsInfo.TTL = time.Now().Add(TTL).Unix()
	}
	if time.Now().Unix() > mrsInfo.TTL {
		go func() {
			label, metrics := getMRSResourceAndMetrics()
			mrsInfo.Lock()
			defer mrsInfo.Unlock()
			mrsInfo.LabelInfo = label
			mrsInfo.FilterMetrics = metrics
			mrsInfo.TTL = time.Now().Add(TTL).Unix()
		}()
	}
	return mrsInfo.LabelInfo, mrsInfo.FilterMetrics
}

func getMRSResourceAndMetrics() (map[string]labelInfo, []model.MetricInfoList) {
	clusters, err := getClusterInfo()
	if err != nil {
		logs.Logger.Errorf("[%s] Get all clusters error: %s", err.Error())
		return nil, nil
	}

	resourceInfos := map[string]labelInfo{}
	for _, cluster := range clusters {
		info := labelInfo{
			Name:  []string{"clusterName", "epId"},
			Value: []string{cluster.Name, cluster.EpId},
		}
		keys, values := getTags(cluster.Tags)
		info.Name = append(info.Name, keys...)
		info.Value = append(info.Value, values...)
		resourceInfos[cluster.ID] = info
	}

	allMetrics, err := listAllMetrics("SYS.MRS")
	if err != nil {
		logs.Logger.Errorf("[%s] Get all metrics of SYS.MRS error: %s", err.Error())
	}
	return resourceInfos, allMetrics
}

func getClusterInfo() ([]ResourceBaseInfo, error) {
	if getResourceFromRMS("SYS.MRS") {
		return getMRSClusterFromRMS()
	}
	return getClusterFromMRS()
}

func getMRSClusterFromRMS() ([]ResourceBaseInfo, error) {
	return getResourcesBaseInfoFromRMS("mrs", "mrs")
}

func getMRSClient() *mrs.MrsClient {
	return mrs.NewMrsClient(mrs.MrsClientBuilder().WithCredential(
		basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()).
		WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(true)).
		WithEndpoint(getEndpoint("mrs", "v1.1")).Build())
}

func getClusterFromMRS() ([]ResourceBaseInfo, error) {
	pageSize := 2000
	currentPage := 1
	pageSizeStr := fmt.Sprintf("%d", pageSize)
	var clusters []mrsmodel.Cluster
	for {
		currentPageStr := fmt.Sprintf("%d", currentPage)
		req := &mrsmodel.ListClustersRequest{PageSize: &pageSizeStr, CurrentPage: &currentPageStr}
		response, err := getMRSClient().ListClusters(req)
		if err != nil {
			logs.Logger.Errorf("Failed to get all mrs cluster : %s", err.Error())
			return nil, err
		}
		clusters = append(clusters, *response.Clusters...)
		if len(*response.Clusters) < pageSize {
			break
		}
		currentPage++
	}

	instances := make([]ResourceBaseInfo, len(clusters))
	for i, cluster := range clusters {
		instances[i].ID = *cluster.ClusterId
		instances[i].Name = *cluster.ClusterName
		instances[i].EpId = *cluster.EnterpriseProjectId
		instances[i].Tags = fmtMrsTags(cluster.Tags)
	}
	return instances, nil
}

// fmtMrsTags mrs的tags信息，返回的是“key5=value5,key1=value1”,需要转换成map
func fmtMrsTags(tagsInfo *string) map[string]string {
	if tagsInfo == nil {
		return nil
	}
	info := strings.Split(*tagsInfo, ",")
	tags := make(map[string]string, len(info))
	for _, tag := range info {
		tagInfo := strings.Split(tag, "=")
		if len(tagInfo) != 2 {
			continue
		}
		tags[tagInfo[0]] = tagInfo[1]
	}
	return tags
}
