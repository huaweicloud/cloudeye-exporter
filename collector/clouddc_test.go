package collector

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
)

func TestCloudDCGetResourceInfo(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	logs.InitLog("")
	cloudDcInfo = serversInfo{}

	patches.ApplyFuncReturn(getIMetalInstanceFromRMS, []IMetalInfo{
		{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   "imetal-001",
				Name: "imetal-instance-1",
				EpId: "ep-001",
				Tags: map[string]string{"env": "prod"},
			},
			IMetalProperties: IMetalProperties{
				ImageName:    "ubuntu-20.04",
				ProductType:  "physical",
				PowerStatus:  "on",
				IP:           "192.168.1.100",
				HealthStatus: "healthy",
			},
		},
		{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   "imetal-002",
				Name: "imetal-instance-2",
				EpId: "ep-002",
				Tags: map[string]string{"env": "test"},
			},
			IMetalProperties: IMetalProperties{
				ImageName:    "centos-7",
				ProductType:  "physical",
				PowerStatus:  "on",
				IP:           "192.168.1.101",
				HealthStatus: "healthy",
			},
		},
	}, nil)

	patches.ApplyFuncReturn(getRackInstanceFromRMS, []RackInfo{
		{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   "rack-001",
				Name: "rack-instance-1",
				EpId: "ep-001",
				Tags: map[string]string{"location": "dc-1"},
			},
			RackProperties: RackProperties{
				Description: "Rack A",
				Location:    "Zone A",
				Power:       "1000W",
				Dc:          "DC-1",
			},
		},
	}, nil)

	// CloudDC uses dimension names "host" and "rack" for resource matching
	patches.ApplyFuncReturn(listAllMetrics, []model.MetricInfoList{
		{
			Namespace:  "SYS.CloudDC",
			MetricName: "cpu_usage",
			Dimensions: []model.MetricsDimension{
				{Name: "host", Value: "imetal-001"},
			},
		},
		{
			Namespace:  "SYS.CloudDC",
			MetricName: "memory_usage",
			Dimensions: []model.MetricsDimension{
				{Name: "host", Value: "imetal-002"},
			},
		},
		{
			Namespace:  "SYS.CloudDC",
			MetricName: "power_consumption",
			Dimensions: []model.MetricsDimension{
				{Name: "rack", Value: "rack-001"},
			},
		},
	}, nil)

	patches.ApplyFuncReturn(IsMetricInfoInWhiteList, true)
	patches.ApplyFuncReturn(GetResourceInfoExpirationTime, 5*time.Minute)

	var getter CloudDcInfo
	labels, metrics := getter.GetResourceInfo()

	assert.Equal(t, 3, len(labels)) // 2 IMetal + 1 Rack
	assert.Equal(t, 3, len(metrics))
	assert.Contains(t, labels, "imetal-001")
	assert.Contains(t, labels, "imetal-002")
	assert.Contains(t, labels, "rack-001")
	assert.Equal(t, []string{"imetal-instance-1", "ep-001", "ubuntu-20.04", "physical", "on", "192.168.1.100", "healthy"}, labels["imetal-001"].Value[:7])
	assert.Greater(t, cloudDcInfo.TTL, time.Now().Unix()-1)
}

func TestCloudDCGetResourceInfoCache(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	logs.InitLog("")
	cachedLabels := map[string]labelInfo{
		"cached-imetal-id": {
			Name:  []string{"name", "epId"},
			Value: []string{"cached-imetal", "ep-0"},
		},
	}
	cachedMetrics := []model.MetricInfoList{
		{Namespace: "SYS.CloudDC", MetricName: "cpu_usage"},
	}
	cloudDcInfo = serversInfo{
		TTL:           time.Now().Add(10 * time.Minute).Unix(),
		LabelInfo:     cachedLabels,
		FilterMetrics: cachedMetrics,
	}

	patches.ApplyFunc(getIMetalInstanceFromRMS, func() ([]IMetalInfo, error) {
		t.Fatal("getIMetalInstanceFromRMS should not be called when cache is valid")
		return nil, nil
	})
	patches.ApplyFunc(getRackInstanceFromRMS, func() ([]RackInfo, error) {
		t.Fatal("getRackInstanceFromRMS should not be called when cache is valid")
		return nil, nil
	})
	patches.ApplyFunc(listAllMetrics, func(namespace string) ([]model.MetricInfoList, error) {
		t.Fatal("listAllMetrics should not be called when cache is valid")
		return nil, nil
	})

	var getter CloudDcInfo
	labels, metrics := getter.GetResourceInfo()

	assert.Equal(t, 1, len(labels))
	assert.Equal(t, []string{"cached-imetal", "ep-0"}, labels["cached-imetal-id"].Value)
	assert.Equal(t, 1, len(metrics))
	assert.Equal(t, "cpu_usage", metrics[0].MetricName)

	cloudDcInfo = serversInfo{}
}

func TestCloudDCGetResourceInfoIMetalError(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	logs.InitLog("")
	cloudDcInfo = serversInfo{}

	patches.ApplyFuncReturn(getIMetalInstanceFromRMS, nil, fmt.Errorf("imetal fetch error"))
	patches.ApplyFuncReturn(GetResourceInfoExpirationTime, 5*time.Minute)

	var getter CloudDcInfo
	labels, metrics := getter.GetResourceInfo()

	assert.Nil(t, labels)
	assert.Nil(t, metrics)
}

func TestCloudDCGetResourceInfoRackError(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	logs.InitLog("")
	cloudDcInfo = serversInfo{}

	patches.ApplyFuncReturn(getIMetalInstanceFromRMS, []IMetalInfo{
		{
			ResourceBaseInfo: ResourceBaseInfo{ID: "imetal-001", Name: "imetal-1", EpId: "ep-001"},
			IMetalProperties: IMetalProperties{ImageName: "ubuntu"},
		},
	}, nil)
	patches.ApplyFuncReturn(getRackInstanceFromRMS, nil, fmt.Errorf("rack fetch error"))
	patches.ApplyFuncReturn(GetResourceInfoExpirationTime, 5*time.Minute)

	var getter CloudDcInfo
	labels, metrics := getter.GetResourceInfo()

	assert.Nil(t, labels)
	assert.Nil(t, metrics)
}

func TestCloudDCGetResourceInfoMetricsError(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	logs.InitLog("")
	cloudDcInfo = serversInfo{}

	patches.ApplyFuncReturn(getIMetalInstanceFromRMS, []IMetalInfo{
		{
			ResourceBaseInfo: ResourceBaseInfo{ID: "imetal-001", Name: "imetal-1", EpId: "ep-001"},
			IMetalProperties: IMetalProperties{ImageName: "ubuntu"},
		},
	}, nil)
	patches.ApplyFuncReturn(getRackInstanceFromRMS, []RackInfo{
		{
			ResourceBaseInfo: ResourceBaseInfo{ID: "rack-001", Name: "rack-1", EpId: "ep-001"},
			RackProperties:   RackProperties{Description: "Rack A"},
		},
	}, nil)
	patches.ApplyFuncReturn(listAllMetrics, nil, fmt.Errorf("metrics fetch error"))
	patches.ApplyFuncReturn(GetResourceInfoExpirationTime, 5*time.Minute)

	var getter CloudDcInfo
	labels, metrics := getter.GetResourceInfo()

	assert.Nil(t, labels)
	assert.Nil(t, metrics)
}

func TestCloudDCGetResourceInfoEmptyInstances(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	logs.InitLog("")
	cloudDcInfo = serversInfo{}

	patches.ApplyFuncReturn(getIMetalInstanceFromRMS, []IMetalInfo{}, nil)
	patches.ApplyFuncReturn(getRackInstanceFromRMS, []RackInfo{}, nil)
	patches.ApplyFuncReturn(listAllMetrics, []model.MetricInfoList{
		{Namespace: "SYS.CloudDC", MetricName: "cpu_usage"},
	}, nil)
	patches.ApplyFuncReturn(IsMetricInfoInWhiteList, true)
	patches.ApplyFuncReturn(GetResourceInfoExpirationTime, 5*time.Minute)

	var getter CloudDcInfo
	labels, metrics := getter.GetResourceInfo()

	assert.Equal(t, 0, len(labels))
	assert.Equal(t, 0, len(metrics))
}

func TestCloudDCGetResourceInfoCacheExpired(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	logs.InitLog("")
	cloudDcInfo = serversInfo{
		TTL:       time.Now().Add(-1 * time.Minute).Unix(),
		LabelInfo: map[string]labelInfo{"old-id": {Name: []string{"name"}, Value: []string{"old-imetal"}}},
	}

	patches.ApplyFuncReturn(getIMetalInstanceFromRMS, []IMetalInfo{
		{
			ResourceBaseInfo: ResourceBaseInfo{ID: "new-imetal-001", Name: "new-imetal", EpId: "ep-new"},
			IMetalProperties: IMetalProperties{ImageName: "centos-8", PowerStatus: "on"},
		},
	}, nil)
	patches.ApplyFuncReturn(getRackInstanceFromRMS, []RackInfo{}, nil)
	patches.ApplyFuncReturn(listAllMetrics, []model.MetricInfoList{
		{
			Namespace:  "SYS.CloudDC",
			MetricName: "cpu_usage",
			Dimensions: []model.MetricsDimension{
				{Name: "host", Value: "new-imetal-001"},
			},
		},
	}, nil)
	patches.ApplyFuncReturn(IsMetricInfoInWhiteList, true)
	patches.ApplyFuncReturn(GetResourceInfoExpirationTime, 5*time.Minute)

	var getter CloudDcInfo
	labels, metrics := getter.GetResourceInfo()

	assert.Equal(t, 1, len(labels))
	assert.Contains(t, labels, "new-imetal-001")
	assert.Equal(t, "new-imetal", labels["new-imetal-001"].Value[0])
	assert.Equal(t, "ep-new", labels["new-imetal-001"].Value[1])
	assert.Equal(t, 1, len(metrics))

	cloudDcInfo = serversInfo{}
}

func TestCloudDCGetResourceInfoMetricFilter(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	logs.InitLog("")
	cloudDcInfo = serversInfo{}

	patches.ApplyFuncReturn(getIMetalInstanceFromRMS, []IMetalInfo{
		{
			ResourceBaseInfo: ResourceBaseInfo{ID: "imetal-001", Name: "imetal-1", EpId: "ep-001"},
			IMetalProperties: IMetalProperties{ImageName: "ubuntu"},
		},
	}, nil)
	patches.ApplyFuncReturn(getRackInstanceFromRMS, []RackInfo{}, nil)

	patches.ApplyFuncReturn(listAllMetrics, []model.MetricInfoList{
		{
			Namespace:  "SYS.CloudDC",
			MetricName: "cpu_usage",
			Dimensions: []model.MetricsDimension{
				{Name: "host", Value: "imetal-001"},
			},
		},
		{
			Namespace:  "SYS.CloudDC",
			MetricName: "internal_debug",
			Dimensions: []model.MetricsDimension{
				{Name: "host", Value: "imetal-001"},
			},
		},
		{
			Namespace:  "SYS.CloudDC",
			MetricName: "memory_usage",
			Dimensions: []model.MetricsDimension{
				{Name: "host", Value: "imetal-001"},
			},
		},
	}, nil)

	patches.ApplyFunc(IsMetricInfoInWhiteList, func(metricInfo model.MetricInfoList) bool {
		return metricInfo.MetricName != "internal_debug"
	})
	patches.ApplyFuncReturn(GetResourceInfoExpirationTime, 5*time.Minute)

	var getter CloudDcInfo
	labels, metrics := getter.GetResourceInfo()

	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 2, len(metrics))
	assert.Equal(t, "cpu_usage", metrics[0].MetricName)
	assert.Equal(t, "memory_usage", metrics[1].MetricName)

	cloudDcInfo = serversInfo{}
}

func TestCloudDCGetResourceInfoWithTags(t *testing.T) {
	patches := getPatches()
	defer patches.Reset()

	logs.InitLog("")
	cloudDcInfo = serversInfo{}

	patches.ApplyFuncReturn(getIMetalInstanceFromRMS, []IMetalInfo{
		{
			ResourceBaseInfo: ResourceBaseInfo{
				ID:   "imetal-001",
				Name: "imetal-1",
				EpId: "ep-001",
				Tags: map[string]string{"env": "prod", "team": "infra"},
			},
			IMetalProperties: IMetalProperties{ImageName: "ubuntu"},
		},
	}, nil)
	patches.ApplyFuncReturn(getRackInstanceFromRMS, []RackInfo{}, nil)
	patches.ApplyFuncReturn(listAllMetrics, []model.MetricInfoList{}, nil)
	patches.ApplyFuncReturn(IsMetricInfoInWhiteList, true)
	patches.ApplyFuncReturn(GetResourceInfoExpirationTime, 5*time.Minute)

	var getter CloudDcInfo
	labels, metrics := getter.GetResourceInfo()

	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 0, len(metrics))
	// Base fields: name, epId, image_name, product_type, power_status, ip, health_status (7 fields)
	// Plus 2 tag fields
	assert.GreaterOrEqual(t, len(labels["imetal-001"].Name), 7)
	assert.Contains(t, labels["imetal-001"].Value, "imetal-1")
	assert.Contains(t, labels["imetal-001"].Value, "ep-001")

	cloudDcInfo = serversInfo{}
}
