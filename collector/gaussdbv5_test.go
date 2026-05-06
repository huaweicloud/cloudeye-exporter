package collector

import (
	"fmt"
	"testing"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	"github.com/stretchr/testify/assert"
)

func TestGAUSSDBV5GetResourceInfo(t *testing.T) {
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	// Mock the metric configuration
	metricConf = map[string]MetricConf{
		"SYS.GAUSSDBV5": {
			Resource: "gaussdbv5",
			DimMetricName: map[string][]string{
				"gaussdbv5_instance_id": {
					"metric1", "metric2",
				},
				"gaussdbv5_instance_id,gaussdbv5_node_id": {
					"metric3", "metric4",
				},
				"gaussdbv5_instance_id,gaussdbv5_pdb_id": {
					"metric5", "metric6",
				},
				"gaussdbv5_instance_id,gaussdbv5_pdb_id,gaussdbv5_pdb_component_id": {
					"metric7", "metric8",
				},
			},
		},
	}

	// Mock the list of instances
	instances := []GaussdbV5Instances{
		{
			InstanceInfo: GaussdbV5Instance{
				InstanceId:          "instance-1",
				InstanceName:        "Instance 1",
				EnterpriseProjectId: "ep-1",
				Tags: []GaussdbV5Tag{
					{Key: "tag1", Value: "value1"},
					{Key: "tag2", Value: "value2"},
				},
			},
			Node: GaussdbV5Node{
				ID:               "node-1",
				Name:             "Node 1",
				Role:             "master",
				Status:           "active",
				AvailabilityZone: "az-1",
			},
			Pdb: GaussdbV5Pdb{
				PdbId:   "pdb-1",
				PdbName: "PDB 1",
			},
			Component: GaussdbV5PdbComponent{
				ComponentId:            "component-1",
				ComponentType:          "type-1",
				ComponentDistributedId: "distributed-1",
				ComponentDetail:        "detail-1",
			},
		},
	}

	// Mock the listInstances function to return the instances
	patches := gomonkey.ApplyFuncReturn(listInstances, instances, nil)
	defer patches.Reset()

	conf.AuthMode = "aksk"
	// Mock the Sync method of the gaussDBClient
	gaussDBClient := getHcClient(getEndpoint("gaussdbV5", "v1.0"))
	patches.ApplyMethod(gaussDBClient, "Sync", func(hcClient *core.HcHttpClient, req interface{}, reqDef *def.HttpRequestDef) (interface{}, error) {
		return &GaussdbV5InstancesResponse{
			Instances: instances,
			Total:     1,
		}, nil
	})

	// Call the method
	resourceInfos, filterMetrics := GAUSSDBV5Info{}.GetResourceInfo()

	// Check the results
	assert.Equal(t, 8, len(filterMetrics))
	assert.Equal(t, 4, len(resourceInfos))

	// Check the first metric
	assert.Equal(t, "SYS.GAUSSDBV5", filterMetrics[0].Namespace)
	assert.Equal(t, "metric1", filterMetrics[0].MetricName)
	assert.Equal(t, "gaussdbv5_instance_id", filterMetrics[0].Dimensions[0].Name)
	assert.Equal(t, "instance-1", filterMetrics[0].Dimensions[0].Value)

	// Check the first resource info
	resourceKey := GetResourceKeyFromMetricInfo(filterMetrics[0])
	assert.Equal(t, "name", resourceInfos[resourceKey].Name[0])
	assert.Equal(t, "Instance 1", resourceInfos[resourceKey].Value[0])
	assert.Equal(t, "epId", resourceInfos[resourceKey].Name[1])
	assert.Equal(t, "ep-1", resourceInfos[resourceKey].Value[1])
}

func TestGAUSSDBV5GetResourceInfoError(t *testing.T) {
	// Mock the listInstances function to return an error
	patches := gomonkey.ApplyFuncReturn(listInstances, nil, fmt.Errorf("mock error"))
	defer patches.Reset()

	// Call the method
	resourceInfos, filterMetrics := GAUSSDBV5Info{}.GetResourceInfo()

	// Check the results
	assert.Equal(t, 8, len(filterMetrics))
	assert.Equal(t, 4, len(resourceInfos))
}
