package collector

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
	"github.com/stretchr/testify/assert"
)

// TestGetAllRecordSets tests the getAllRecordSets function
func TestGetAllRecordSets(t *testing.T) {
	outputs := []gomonkey.OutputCell{
		{
			Values: gomonkey.Params{&model.ListRecordSetsWithLineResponse{
				Recordsets: &[]model.QueryRecordSetWithLineAndTagsResp{
					{
						Id:   strPtr("recordset-1"),
						Name: strPtr("example.com"),
						Type: strPtr("A"),
					},
				},
				HttpStatusCode: 200,
			}, nil},
		},
		{
			Values: gomonkey.Params{&model.ListRecordSetsWithLineResponse{
				Recordsets:     &[]model.QueryRecordSetWithLineAndTagsResp{},
				HttpStatusCode: 200,
			}, nil},
		},
	}
	conf.AuthMode = "aksk"
	conf.AccessKey = "temp_ak"
	conf.SecretKey = "temp_sk"
	// Apply the monkey patch
	dnsClient := getDnsClient()
	patches := gomonkey.ApplyMethodSeq(dnsClient, "ListRecordSetsWithLine", outputs)
	defer patches.Reset()
	// Call the function
	recordSets, err := getAllRecordSets("public")
	// Assert the results
	assert.NoError(t, err)
	assert.Len(t, recordSets, 1)
	assert.Equal(t, "recordset-1", *recordSets[0].Id)
	assert.Equal(t, "example.com", *recordSets[0].Name)
	assert.Equal(t, "A", *recordSets[0].Type)
}

func TestDNSGetResourceInfo(t *testing.T) {
	// Mock the metric configuration
	metricConf = map[string]MetricConf{
		"SYS.DNS": {
			Resource: "dns",
			DimMetricName: map[string][]string{
				"dns_recordset_id": {"metric1", "metric2"},
				"dns_zone_id":      {"metric3", "metric4"},
			},
		},
	}

	value1Str := "value1"
	value2Str := "value2"
	tags1 := []model.Tag{{Key: "tag1", Value: &value1Str}, {Key: "tag2", Value: &value2Str}}
	// Mock the record sets
	recordSets := []model.QueryRecordSetWithLineAndTagsResp{
		{
			Id:   strPtr("recordset-1"),
			Name: strPtr("example.com"),
			Type: strPtr("A"),
			Tags: &tags1,
		},
	}

	value3Str := "value3"
	value4Str := "value4"
	tags2 := []model.Tag{{Key: "tag3", Value: &value3Str}, {Key: "tag4", Value: &value4Str}}
	// Mock the zones
	zones := []model.PublicZoneResp{
		{
			Id:                  strPtr("zone-1"),
			Name:                strPtr("example.com"),
			ZoneType:            strPtr("public"),
			EnterpriseProjectId: strPtr("ep-1"),
			Tags:                &tags2,
		},
	}

	// Mock the getAllRecordSets and getAllPublicZones functions
	patches := gomonkey.ApplyFuncReturn(getAllRecordSets, recordSets, nil)
	defer patches.Reset()

	patches.ApplyFuncReturn(getAllPublicZones, zones, nil)

	// Call the method
	resourceInfos, filterMetrics := DNSInfo{}.GetResourceInfo()

	// Check the results
	assert.Equal(t, 6, len(filterMetrics))
	assert.Equal(t, 2, len(resourceInfos))

	// Check the first metric
	assert.Equal(t, "SYS.DNS", filterMetrics[0].Namespace)
	assert.Equal(t, "metric1", filterMetrics[0].MetricName)
	assert.Equal(t, "dns_recordset_id", filterMetrics[0].Dimensions[0].Name)
	assert.Equal(t, "recordset-1", filterMetrics[0].Dimensions[0].Value)

	// Check the first resource info
	resourceKey := GetResourceKeyFromMetricInfo(filterMetrics[0])
	assert.Equal(t, "name", resourceInfos[resourceKey].Name[0])
	assert.Equal(t, "example.com", resourceInfos[resourceKey].Value[0])
	assert.Equal(t, "type", resourceInfos[resourceKey].Name[1])
	assert.Equal(t, "A", resourceInfos[resourceKey].Value[1])
}

func TestDNSGetResourceInfoError(t *testing.T) {
	// Mock the getAllRecordSets function to return an error
	patches := gomonkey.ApplyFuncReturn(getAllRecordSets, nil, fmt.Errorf("mock error"))
	defer patches.Reset()

	// Call the method
	resourceInfos, filterMetrics := DNSInfo{}.GetResourceInfo()

	// Check the results
	assert.Equal(t, 6, len(filterMetrics))
	assert.Equal(t, 2, len(resourceInfos))
}

// Helper function to create a string pointer
func strPtr(s string) *string {
	return &s
}

// Test for getAllPublicZones
func TestGetAllPublicZones(t *testing.T) {
	// Mock the getAllPublicZonesByEpId function
	epId := "ep-1"
	zones := []model.PublicZoneResp{
		{
			Id:                  strPtr("zone-1"),
			Name:                strPtr("example.com"),
			ZoneType:            strPtr("public"),
			EnterpriseProjectId: strPtr(epId),
		},
	}

	// Mock the getDnsEpIdRequestPart function to return a single epId
	patches := gomonkey.ApplyFuncReturn(getDnsEpIdRequestPart, []string{epId})
	defer patches.Reset()

	// Mock the getAllPublicZonesByEpId function
	patches.ApplyFuncReturn(getAllPublicZonesByEpId, zones, nil)

	// Call the function
	publicZones, err := getAllPublicZones()

	// Assert the results
	assert.NoError(t, err)
	assert.Len(t, publicZones, 1)
	assert.Equal(t, "zone-1", *publicZones[0].Id)
	assert.Equal(t, "example.com", *publicZones[0].Name)
	assert.Equal(t, "public", *publicZones[0].ZoneType)
	assert.Equal(t, epId, *publicZones[0].EnterpriseProjectId)
}

// Test for getAllPublicZonesByEpId
func TestGetAllPublicZonesByEpId(t *testing.T) {
	conf.AuthMode = "aksk"
	conf.AccessKey = "temp_ak"
	conf.SecretKey = "temp_sk"
	// Mock the getDnsClient function
	dnsClient := getDnsClient()
	patches := gomonkey.ApplyFuncReturn(getDnsClient, dnsClient)
	defer patches.Reset()

	// Mock the ListPublicZones method
	epId := "ep-1"
	response := &model.ListPublicZonesResponse{
		Zones: &[]model.PublicZoneResp{
			{
				Id:                  strPtr("zone-1"),
				Name:                strPtr("example.com"),
				ZoneType:            strPtr("public"),
				EnterpriseProjectId: strPtr(epId),
			},
		},
		HttpStatusCode: 200,
	}

	// Apply the monkey patch
	patches.ApplyMethodSeq(dnsClient, "ListPublicZones", []gomonkey.OutputCell{
		{
			Values: gomonkey.Params{response, nil},
		},
		{
			Values: gomonkey.Params{&model.ListPublicZonesResponse{
				Zones:          &[]model.PublicZoneResp{},
				HttpStatusCode: 200,
			}, nil},
		},
	})

	// Call the function
	publicZones, err := getAllPublicZonesByEpId(epId)

	// Assert the results
	assert.NoError(t, err)
	assert.Len(t, publicZones, 1)
	assert.Equal(t, "zone-1", *publicZones[0].Id)
	assert.Equal(t, "example.com", *publicZones[0].Name)
	assert.Equal(t, "public", *publicZones[0].ZoneType)
	assert.Equal(t, epId, *publicZones[0].EnterpriseProjectId)
}
