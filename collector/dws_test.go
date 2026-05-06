package collector

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dws/v2/model"
	"github.com/stretchr/testify/assert"
)

// TestGetResourceInfo tests the GetResourceInfo method of DWSInfo
func TestDWSGetResourceInfo(t *testing.T) {
	// Mock the metric configuration
	metricConf = map[string]MetricConf{
		"SYS.DWS": {
			Resource: "dws",
			DimMetricName: map[string][]string{
				"datastore_id":    {"metric1", "metric2"},
				"dws_instance_id": {"metric3", "metric4"},
			},
		},
	}

	// Mock the clusters
	cluster := model.ClusterInfo{
		Id:                  "cluster-1",
		Name:                "test-cluster",
		EnterpriseProjectId: "ep-1",
		Tags: []model.Tags{
			{
				Key:   "tag1",
				Value: "value1",
			},
			{
				Key:   "tag2",
				Value: "value2",
			},
		},
		Nodes: []model.Nodes{
			{Id: "node-1"},
		},
	}

	clusters := []model.ClusterInfo{cluster}

	// Mock the queryDwsCluster function
	patches := gomonkey.ApplyFuncReturn(queryDwsCluster, clusters, nil)
	defer patches.Reset()

	// Call the method
	resourceInfos, filterMetrics := DWSInfo{}.GetResourceInfo()

	// Check the results
	assert.Equal(t, 4, len(filterMetrics))
	assert.Equal(t, 2, len(resourceInfos))

	// Check the first metric
	assert.Equal(t, "SYS.DWS", filterMetrics[0].Namespace)
	assert.Equal(t, "metric1", filterMetrics[0].MetricName)
	assert.Equal(t, "datastore_id", filterMetrics[0].Dimensions[0].Name)
	assert.Equal(t, "cluster-1", filterMetrics[0].Dimensions[0].Value)

	// Check the first resource info
	resourceKey := GetResourceKeyFromMetricInfo(filterMetrics[0])
	assert.Equal(t, "clusterName", resourceInfos[resourceKey].Name[0])
	assert.Equal(t, "test-cluster", resourceInfos[resourceKey].Value[0])
	assert.Equal(t, "epId", resourceInfos[resourceKey].Name[1])
	assert.Equal(t, "ep-1", resourceInfos[resourceKey].Value[1])
}

// TestGetResourceInfoError tests the GetResourceInfo method when an error occurs
func TestDWSGetResourceInfoError(t *testing.T) {
	// Mock the queryDwsCluster function to return an error
	patches := gomonkey.ApplyFuncReturn(queryDwsCluster, nil, fmt.Errorf("mock error"))
	defer patches.Reset()

	// Call the method
	resourceInfos, filterMetrics := DWSInfo{}.GetResourceInfo()

	// Check the results
	assert.Equal(t, 4, len(filterMetrics))
	assert.Equal(t, 2, len(resourceInfos))
}

// Test for queryDwsCluster
func TestQueryDwsCluster(t *testing.T) {
	conf.AuthMode = "aksk"
	conf.AccessKey = "temp_ak"
	conf.SecretKey = "temp_sk"
	// Mock the getEpIdRequestPart function to return a single epId
	epId := "ep-1"
	patches := gomonkey.ApplyFuncReturn(getEpIdRequestPart, []string{epId})
	defer patches.Reset()

	// Mock the queryDwsClusterByEpId function
	clusters := []model.ClusterInfo{
		{
			Id:                  "cluster-1",
			Name:                "test-cluster",
			EnterpriseProjectId: "ep-1",
			Tags: []model.Tags{
				{
					Key:   "tag1",
					Value: "value1",
				},
				{
					Key:   "tag2",
					Value: "value2",
				},
			},
			Nodes: []model.Nodes{
				{Id: "node-1"},
			},
		},
	}
	patches.ApplyFuncReturn(queryDwsClusterByEpId, &model.ListClustersResponse{
		Clusters: &clusters,
	}, nil)

	// Call the function
	dwsClusters, err := queryDwsCluster()

	// Assert the results
	assert.NoError(t, err)
	assert.Len(t, dwsClusters, 1)
	assert.Equal(t, "cluster-1", dwsClusters[0].Id)
	assert.Equal(t, "test-cluster", dwsClusters[0].Name)
	assert.Equal(t, "ep-1", dwsClusters[0].EnterpriseProjectId)
	assert.Equal(t, "tag1", dwsClusters[0].Tags[0].Key)
	assert.Equal(t, "value1", dwsClusters[0].Tags[0].Value)
	assert.Equal(t, "tag2", dwsClusters[0].Tags[1].Key)
	assert.Equal(t, "value2", dwsClusters[0].Tags[1].Value)
	assert.Len(t, dwsClusters[0].Nodes, 1)
	assert.Equal(t, "node-1", dwsClusters[0].Nodes[0].Id)
}

// Test for queryDwsClusterByEpId
func TestQueryDwsClusterByEpId(t *testing.T) {
	conf.AuthMode = "aksk"
	conf.AccessKey = "temp_ak"
	conf.SecretKey = "temp_sk"
	// Mock the getHcClient function
	dwsClient := getHcClient(getEndpoint("dws", "v1.0"))
	patches := gomonkey.ApplyFuncReturn(getHcClient, dwsClient)
	defer patches.Reset()

	// Mock the Sync method of the dwsClient
	clusters := []model.ClusterInfo{
		{
			Id:                  "cluster-1",
			Name:                "test-cluster",
			EnterpriseProjectId: "ep-1",
			Tags: []model.Tags{
				{
					Key:   "tag1",
					Value: "value1",
				},
				{
					Key:   "tag2",
					Value: "value2",
				},
			},
			Nodes: []model.Nodes{
				{Id: "node-1"},
			},
		},
	}
	response := &model.ListClustersResponse{
		Clusters: &clusters,
	}
	patches.ApplyMethodSeq(dwsClient, "Sync", []gomonkey.OutputCell{
		{
			Values: gomonkey.Params{response, nil},
		},
	})

	// Call the function
	dwsResponse, err := queryDwsClusterByEpId(dwsClient, "ep-1")

	// Assert the results
	assert.NoError(t, err)
	assert.NotNil(t, dwsResponse)
	assert.Len(t, *dwsResponse.Clusters, 1)
	assert.Equal(t, "cluster-1", (*dwsResponse.Clusters)[0].Id)
	assert.Equal(t, "test-cluster", (*dwsResponse.Clusters)[0].Name)
	assert.Equal(t, "ep-1", (*dwsResponse.Clusters)[0].EnterpriseProjectId)
	assert.Equal(t, "tag1", (*dwsResponse.Clusters)[0].Tags[0].Key)
	assert.Equal(t, "value1", (*dwsResponse.Clusters)[0].Tags[0].Value)
	assert.Equal(t, "tag2", (*dwsResponse.Clusters)[0].Tags[1].Key)
	assert.Equal(t, "value2", (*dwsResponse.Clusters)[0].Tags[1].Value)
	assert.Len(t, (*dwsResponse.Clusters)[0].Nodes, 1)
	assert.Equal(t, "node-1", (*dwsResponse.Clusters)[0].Nodes[0].Id)
}
