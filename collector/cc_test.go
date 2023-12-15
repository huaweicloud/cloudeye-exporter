package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	cc "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cc/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cc/v3/model"
	"github.com/stretchr/testify/assert"
)

func TestCCGetResourceInfo(t *testing.T) {
	sysConfig := map[string][]string{CCConfigDimNames: {"network_incoming_bits_rate"}}
	patches := gomonkey.ApplyFuncReturn(getMetricConfigMap, sysConfig)

	defaultEpId := "0"
	id := "connection-0001"
	name := "connection1"
	connectionsPage := model.ListCloudConnectionsResponse{
		HttpStatusCode: 200,
		CloudConnections: &[]model.CloudConnection{
			{Id: &id, Name: &name, EnterpriseProjectId: &defaultEpId},
		},
		PageInfo: &model.PageInfo{},
	}

	ccClient := cc.CcClient{}
	patches.ApplyFuncReturn(getCCClient, &ccClient)
	patches.ApplyMethodFunc(&ccClient, "ListCloudConnections", func(req *model.ListCloudConnectionsRequest) (*model.ListCloudConnectionsResponse, error) {
		return &connectionsPage, nil
	})

	packagesPage := model.ListBandwidthPackagesResponse{
		HttpStatusCode: 200,
		BandwidthPackages: &[]model.BandwidthPackage{
			{Id: &id, Name: &name, EnterpriseProjectId: &defaultEpId},
		},
		PageInfo: &model.PageInfo{},
	}
	patches.ApplyMethodFunc(&ccClient, "ListBandwidthPackages", func(req *model.ListBandwidthPackagesRequest) (*model.ListBandwidthPackagesResponse, error) {
		return &packagesPage, nil
	})
	defer patches.Reset()

	localRegionId := "cn-test-01"
	remoteRegionId := "cn-test-02"
	bandwidthsPage := model.ListInterRegionBandwidthsResponse{
		HttpStatusCode: 200,
		InterRegionBandwidths: &[]model.InterRegionBandwidth{
			{
				Id:                 &id,
				Name:               &name,
				CloudConnectionId:  &id,
				BandwidthPackageId: &id,
				InterRegions: &[]model.InterRegion{
					{LocalRegionId: &localRegionId, RemoteRegionId: &remoteRegionId},
				}},
		},
		PageInfo: &model.PageInfo{},
	}
	patches.ApplyMethodFunc(&ccClient, "ListInterRegionBandwidths", func(req *model.ListInterRegionBandwidthsRequest) (*model.ListInterRegionBandwidthsResponse, error) {
		return &bandwidthsPage, nil
	})
	defer patches.Reset()

	var ccgetter CCInfo
	labels, metrics := ccgetter.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 1, len(metrics))
}
