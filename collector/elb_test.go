package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/elb/v3/model"
	"github.com/stretchr/testify/assert"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

func TestELBInfo_GetResourceInfo(t *testing.T) {
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	conf.AuthMode = "aksk"
	metricConf = map[string]MetricConf{
		"SYS.ELB": {
			DimMetricName: map[string][]string{"lbaas_instance_id": {"mc_l7_http_2xx"}, "lbaas_instance_id,lbaas_listener_id": {"mc_l7_http_2xx"},
				"lbaas_instance_id,lbaas_pool_id": {"m18_l7_upstream_2xx"}, "lbaas_instance_id,available_zone": {"mc_l7_http_2xx"}, "lbaas_instance_id,ip_address": {"upstream_tls_negotiation_error"}},
		},
	}
	patches := getPatches()
	logs.InitLog("")
	tags := []model.Tag{
		{
			Key:   new(string),
			Value: new(string),
		},
	}
	*tags[0].Key = "test_tag"
	*tags[0].Value = "test_tag_value"
	lbOutputs := []gomonkey.OutputCell{
		{
			Values: gomonkey.Params{&model.ListLoadBalancersResponse{
				Loadbalancers: &[]model.LoadBalancer{
					{Name: "test_lb", EnterpriseProjectId: "test_epId", VipAddress: "127.0.0.1", Ipv6VipAddress: "test_ipv6_addr", Provider: "test_provider",
						Listeners: []model.ListenerRef{{Id: "test_listener"}}, Pools: []model.PoolRef{{Id: "test_pool"}}, AvailabilityZoneList: []string{"cn-north-7"}, Tags: tags},
				},
			}, nil},
		},
		{
			Values: gomonkey.Params{&model.ListLoadBalancersResponse{
				Loadbalancers: &[]model.LoadBalancer{},
			}, nil},
		},
	}
	patches.ApplyMethodSeq(getELBClient(), "ListLoadBalancers", lbOutputs)

	listenerOutputs := []gomonkey.OutputCell{
		{
			Values: gomonkey.Params{&model.ListListenersResponse{
				Listeners: &[]model.Listener{
					{Id: "test_listener", Name: "test_listener", Protocol: "http", ProtocolPort: 8087},
				},
			}, nil},
		},
		{
			Values: gomonkey.Params{&model.ListListenersResponse{
				Listeners: &[]model.Listener{},
			}, nil},
		},
	}
	patches.ApplyMethodSeq(getELBClient(), "ListListeners", listenerOutputs)

	poolOutputs := []gomonkey.OutputCell{
		{
			Values: gomonkey.Params{&model.ListPoolsResponse{
				Pools: &[]model.Pool{
					{Id: "test_pool", Name: "test_pool", Protocol: "http"},
				},
			}, nil},
		},
		{
			Values: gomonkey.Params{&model.ListPoolsResponse{
				Pools: &[]model.Pool{},
			}, nil},
		},
	}
	patches.ApplyMethodSeq(getELBClient(), "ListPools", poolOutputs)

	zoneOutputs := []gomonkey.OutputCell{
		{
			Values: gomonkey.Params{&model.ListAvailabilityZonesResponse{
				AvailabilityZones: &[][]model.AvailabilityZone{
					[]model.AvailabilityZone{
						{Code: "cn-north-7", State: "ACTIVE", Protocol: []string{"l4", "l7"}, PublicBorderGroup: "center", Category: 0},
					},
				},
			}, nil},
		},
		{
			Values: gomonkey.Params{&model.ListAvailabilityZonesResponse{
				AvailabilityZones: &[][]model.AvailabilityZone{},
			}, nil},
		},
	}
	patches.ApplyMethodSeq(getELBClient(), "ListAvailabilityZones", zoneOutputs)
	var elbGetter ELBInfo
	labelInfos, metricInfos := elbGetter.GetResourceInfo()
	assert.NotEmpty(t, labelInfos)
	assert.NotEmpty(t, metricInfos)
	expectLabelInfos := map[string]labelInfo{}
	expectLabelInfos[""] = labelInfo{
		Name:  []string{"name", "epId", "vip_address", "provider", "test_tag", "public_ip_version4", "public_ip_version6"},
		Value: []string{"test_lb", "test_epId", "127.0.0.1", "test_provider", "test_tag_value", "", ""},
	}
	expectLabelInfos[".test_listener"] = labelInfo{
		Name:  []string{"name", "epId", "vip_address", "provider", "test_tag", "listener_name", "port", "protocol", "public_ip_version4", "public_ip_version6"},
		Value: []string{"test_lb", "test_epId", "127.0.0.1", "test_provider", "test_tag_value", "test_listener", "8087", "http", "", ""},
	}
	expectLabelInfos[".test_pool"] = labelInfo{
		Name:  []string{"name", "epId", "vip_address", "provider", "test_tag", "pool_name", "pool_protocol", "public_ip_version4", "public_ip_version6"},
		Value: []string{"test_lb", "test_epId", "127.0.0.1", "test_provider", "test_tag_value", "test_pool", "http", "", ""},
	}
	expectLabelInfos["127.0.0.1."] = labelInfo{
		Name:  []string{"name", "epId", "vip_address", "provider", "test_tag", "public_ip_version4", "public_ip_version6"},
		Value: []string{"test_lb", "test_epId", "127.0.0.1", "test_provider", "test_tag_value", "", ""},
	}
	expectLabelInfos["cn-north-7."] = labelInfo{
		Name:  []string{"name", "epId", "vip_address", "provider", "test_tag", "state", "public_border_group", "category", "protocol", "public_ip_version4", "public_ip_version6"},
		Value: []string{"test_lb", "test_epId", "127.0.0.1", "test_provider", "test_tag_value", "ACTIVE", "center", "0", "l4,l7", "", ""},
	}
	expectLabelInfos["test_ipv6_addr."] = labelInfo{
		Name:  []string{"name", "epId", "vip_address", "provider", "test_tag", "public_ip_version4", "public_ip_version6"},
		Value: []string{"test_lb", "test_epId", "127.0.0.1", "test_provider", "test_tag_value", "", ""},
	}
	for key, labelInfo := range labelInfos {
		assert.NotEmpty(t, expectLabelInfos[key])
		assert.True(t, labelInfoEqual(labelInfo, expectLabelInfos[key]))
	}

}

func labelInfoEqual(a, b labelInfo) bool {
	if len(a.Name) != len(b.Name) {
		return false
	}
	if len(a.Value) != len(b.Value) {
		return false
	}
	aMap := map[string]string{}
	for index, name := range a.Name {
		aMap[name] = a.Name[index]
	}
	bMap := map[string]string{}
	for index, name := range b.Name {
		bMap[name] = b.Name[index]
	}
	for name, value := range aMap {
		if _, ok := bMap[name]; !ok {
			return false
		}
		if value != bMap[name] {
			return false
		}
	}
	return true
}
