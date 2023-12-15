package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdm/v1/model"
	"github.com/stretchr/testify/assert"
)

func TestCDMGetResourceInfo(t *testing.T) {
	sysConfig := map[string][]string{"instance_id": {"cpu_usage"}}
	patches := gomonkey.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	connectionsPage := model.ListClustersResponse{
		HttpStatusCode: 200,
		Clusters: &[]model.Clusters{
			{
				Id:   "cluster-00001",
				Name: "cluster1",
				Instances: &[]model.ClusterDetailInstance{
					{Id: "instance-00001", Name: "instance1"},
				},
			},
		},
	}
	cdmClient := getCDMClient()
	patches.ApplyMethodFunc(cdmClient, "ListClusters", func(req *model.ListClustersRequest) (*model.ListClustersResponse, error) {
		return &connectionsPage, nil
	})

	defer patches.Reset()

	var cdmgetter CDMInfo
	labels, metrics := cdmgetter.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 1, len(metrics))
}
