package collector

import (
	"errors"
	"net/http"
	"strconv"
	"unsafe"

	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/openstack"
	"github.com/huaweicloud/golangsdk/openstack/autoscaling/v1/groups"
	"github.com/huaweicloud/golangsdk/openstack/blockstorage/v2/volumes"
	"github.com/huaweicloud/golangsdk/openstack/ces/v1/metricdata"
	"github.com/huaweicloud/golangsdk/openstack/ces/v1/metrics"
	"github.com/huaweicloud/golangsdk/openstack/compute/v2/servers"
	dcs "github.com/huaweicloud/golangsdk/openstack/dcs/v1/instances"
	dms "github.com/huaweicloud/golangsdk/openstack/dms/v1/instances"
	"github.com/huaweicloud/golangsdk/openstack/dms/v1/queues"
	"github.com/huaweicloud/golangsdk/openstack/fgs/v2/function"
	"github.com/huaweicloud/golangsdk/openstack/networking/v2/extensions/lbaas_v2/listeners"
	"github.com/huaweicloud/golangsdk/openstack/networking/v2/extensions/lbaas_v2/loadbalancers"
	"github.com/huaweicloud/golangsdk/openstack/networking/v2/extensions/natgateways"
	rds "github.com/huaweicloud/golangsdk/openstack/rds/v3/instances"
	"github.com/huaweicloud/golangsdk/openstack/vpc/v1/bandwidths"
	"github.com/huaweicloud/golangsdk/openstack/vpc/v1/publicips"
)

type Config struct {
	AccessKey        string
	SecretKey        string
	DomainID         string
	DomainName       string
	EndpointType     string
	IdentityEndpoint string
	Insecure         bool
	Password         string
	Region           string
	TenantID         string
	TenantName       string
	Token            string
	Username         string
	UserID           string

	HwClient *golangsdk.ProviderClient
}

func buildClient(c *Config) error {
	if c.AccessKey != "" && c.SecretKey != "" {
		return buildClientByAKSK(c)
	} else if c.Password != "" && (c.Username != "" || c.UserID != "") {
		return buildClientByPassword(c)
	}

	return errors.New("Must config token or aksk or username password to be authorized")
}

func buildClientByPassword(c *Config) error {
	var pao, dao golangsdk.AuthOptions

	pao = golangsdk.AuthOptions{
		DomainID:   c.DomainID,
		DomainName: c.DomainName,
		TenantID:   c.TenantID,
		TenantName: c.TenantName,
	}

	dao = golangsdk.AuthOptions{
		DomainID:   c.DomainID,
		DomainName: c.DomainName,
	}

	for _, ao := range []*golangsdk.AuthOptions{&pao, &dao} {
		ao.IdentityEndpoint = c.IdentityEndpoint
		ao.Password = c.Password
		ao.Username = c.Username
		ao.UserID = c.UserID
	}

	return genClients(c, pao, dao)
}

func buildClientByAKSK(c *Config) error {
	var pao, dao golangsdk.AKSKAuthOptions

	pao = golangsdk.AKSKAuthOptions{
		ProjectName: c.TenantName,
		ProjectId:   c.TenantID,
	}

	dao = golangsdk.AKSKAuthOptions{
		DomainID: c.DomainID,
		Domain:   c.DomainName,
	}

	for _, ao := range []*golangsdk.AKSKAuthOptions{&pao, &dao} {
		ao.IdentityEndpoint = c.IdentityEndpoint
		ao.AccessKey = c.AccessKey
		ao.SecretKey = c.SecretKey
	}
	return genClients(c, pao, dao)
}

func genClients(c *Config, pao, dao golangsdk.AuthOptionsProvider) error {
	client, err := genClient(c, pao)
	if err != nil {
		return err
	}
	c.HwClient = client
	return err
}

func genClient(c *Config, ao golangsdk.AuthOptionsProvider) (*golangsdk.ProviderClient, error) {
	client, err := openstack.NewClient(ao.GetIdentityEndpoint())
	if err != nil {
		return nil, err
	}

	client.HTTPClient = http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if client.AKSKAuthOptions.AccessKey != "" {
				golangsdk.ReSign(req, golangsdk.SignOptions{
					AccessKey: client.AKSKAuthOptions.AccessKey,
					SecretKey: client.AKSKAuthOptions.SecretKey,
				})
			}
			return nil
		},
	}

	err = openstack.Authenticate(client, ao)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func InitConfig(config *CloudConfig) (*Config, error) {
	auth := config.Auth
	configOptions := Config{
		IdentityEndpoint: auth.AuthURL,
		TenantName:       auth.ProjectName,
		AccessKey:        auth.AccessKey,
		SecretKey:        auth.SecretKey,
		DomainName:       auth.DomainName,
		Username:         auth.UserName,
		Region:           auth.Region,
		Password:         auth.Password,
		Insecure:         true,
	}

	err := buildClient(&configOptions)
	if err != nil {
		logs.Logger.Errorf("Failed to build client: %s", err.Error())
		return nil, err
	}

	return &configOptions, err
}

func getCESClient(c *Config) (*golangsdk.ServiceClient, error) {
	client, clientErr := openstack.NewCESV1(c.HwClient, golangsdk.EndpointOpts{
		Region: c.Region,
	})
	if clientErr != nil {
		logs.Logger.Errorf("Failed to get the NewCESV1 client: %s", clientErr.Error())
		return nil, clientErr
	}

	return client, nil
}

func getELBlient(c *Config) (*golangsdk.ServiceClient, error) {
	client, clientErr := openstack.NewNetworkV2(c.HwClient, golangsdk.EndpointOpts{
		Region: c.Region,
	})
	if clientErr != nil {
		logs.Logger.Errorf("Failed to get the NewLoadBalancerV2 client: %s", clientErr.Error())
		return nil, clientErr
	}

	return client, nil
}

func transMetric(metric metrics.Metric) metricdata.Metric {
	var m metricdata.Metric
	m.Namespace = metric.Namespace
	m.MetricName = metric.MetricName
	m.Dimensions = *(*[]metricdata.Dimension)(unsafe.Pointer(&metric.Dimensions))
	return m
}

func getBatchMetricData(c *Config, metrics *[]metricdata.Metric,
	from string, to string) (*[]metricdata.MetricData, error) {

	ifrom, err := strconv.ParseInt(from, 10, 64)
	if err != nil {
		logs.Logger.Errorf("Failed to Parse from: %s", err.Error())
		return nil, err
	}
	ito, err := strconv.ParseInt(to, 10, 64)
	if err != nil {
		logs.Logger.Errorf("Failed to Parse to: %s", err.Error())
		return nil, err
	}
	options := metricdata.BatchQueryOpts{
		Metrics: *metrics,
		From:    ifrom,
		To:      ito,
		Period:  "1",
		Filter:  "average",
	}

	client, err := getCESClient(c)
	if err != nil {
		logs.Logger.Errorf("Failed to get ces client: %s", err.Error())
		return nil, err
	}

	v, err := metricdata.BatchQuery(client, options).ExtractMetricDatas()
	if err != nil {
		logs.Logger.Errorf("Failed to get metricdata: %s", err.Error())
		return nil, err
	}

	return &v, nil
}

func getAllMetric(client *Config, namespace string) (*[]metrics.Metric, error) {
	c, err := getCESClient(client)
	if err != nil {
		logs.Logger.Errorf("Get CES client error: %s", err.Error())
		return nil, err
	}
	limit := 1000
	allpages, err := metrics.List(c, metrics.ListOpts{Namespace: namespace, Limit: &limit}).AllPages()
	if err != nil {
		logs.Logger.Errorf("Get all metric all pages error: %s", err.Error())
		return nil, err
	}

	v, err := metrics.ExtractAllPagesMetrics(allpages)
	if err != nil {
		logs.Logger.Errorf("Get all metric pages error: %s", err.Error())
		return nil, err
	}

	return &v.Metrics, nil
}

func getAllLoadBalancer(client *Config) (*[]loadbalancers.LoadBalancer, error) {
	c, err := getELBlient(client)
	if err != nil {
		return nil, err
	}

	allPages, err := loadbalancers.List(c, loadbalancers.ListOpts{}).AllPages()
	if err != nil {
		logs.Logger.Errorf("List load balancer error: %s", err.Error())
		return nil, err
	}

	allLoadbalancers, err := loadbalancers.ExtractLoadBalancers(allPages)
	if err != nil {
		logs.Logger.Errorf("Extract load balancer pages error: %s", err.Error())
		return nil, err
	}

	return &allLoadbalancers, nil
}

func getAllListener(client *Config) (*[]listeners.Listener, error) {
	c, err := getELBlient(client)
	if err != nil {
		return nil, err
	}

	allPages, err := listeners.List(c, listeners.ListOpts{}).AllPages()
	if err != nil {
		logs.Logger.Errorf("List listener all pages error: %s", err.Error())
		return nil, err
	}

	allListeners, err := listeners.ExtractListeners(allPages)
	if err != nil {
		logs.Logger.Errorf("Extract listener pages error: %s", err.Error())
		return nil, err
	}

	return &allListeners, nil
}

func getAllNat(c *Config) (*[]natgateways.NatGateway, error) {
	client, err := openstack.NewNatV2(c.HwClient, golangsdk.EndpointOpts{
		Region: c.Region,
	})
	if err != nil {
		return nil, err
	}

	allPages, err := natgateways.List(client, natgateways.ListOpts{}).AllPages()
	if err != nil {
		logs.Logger.Errorf("List nat gateways error: %s", err.Error())
		return nil, err
	}

	allNatGateways, err := natgateways.ExtractNatGateways(allPages)
	if err != nil {
		logs.Logger.Errorf("Extract nat gateway pages error: %s", err.Error())
		return nil, err
	}

	return &allNatGateways, nil
}

func getAllRds(c *Config) (*rds.ListRdsResponse, error) {
	client, err := openstack.NewRDSV3(c.HwClient, golangsdk.EndpointOpts{
		Region: c.Region,
	})
	if err != nil {
		logs.Logger.Errorf("Unable to get NewRDSV3 client: %s", err.Error())
		return nil, err
	}

	allPages, err := rds.List(client, rds.ListRdsInstanceOpts{}).AllPages()
	if err != nil {
		logs.Logger.Errorf("List rds error: %s", err.Error())
		return nil, err
	}

	allRds, err := rds.ExtractRdsInstances(allPages)
	if err != nil {
		logs.Logger.Errorf("Extract rds pages error: %s", err.Error())
		return nil, err
	}

	return &allRds, nil
}

func getAllDcs(c *Config) (*dcs.ListDcsResponse, error) {
	client, err := openstack.NewDCSServiceV1(c.HwClient, golangsdk.EndpointOpts{
		Region: c.Region,
	})
	if err != nil {
		logs.Logger.Errorf("Failed to NewDCSServiceV1, error: %s", err.Error())
		return nil, err
	}

	allPages, err := dcs.List(client, dcs.ListDcsInstanceOpts{}).AllPages()
	if err != nil {
		logs.Logger.Errorf("List dcs error: %s", err.Error())
		return nil, err
	}

	allDcs, err := dcs.ExtractDcsInstances(allPages)
	if err != nil {
		logs.Logger.Errorf("Extract dcs pages error: %s", err.Error())
		return nil, err
	}

	return &allDcs, nil
}

func getAllDms(c *Config) (*dms.ListDmsResponse, error) {
	client, err := openstack.NewDMSServiceV1(c.HwClient, golangsdk.EndpointOpts{
		Region: c.Region,
	})
	if err != nil {
		logs.Logger.Errorf("Failed to NewDMSServiceV1, error: %s", err.Error())
		return nil, err
	}

	allPages, err := dms.List(client, dms.ListDmsInstanceOpts{}).AllPages()
	if err != nil {
		logs.Logger.Errorf("List dms instances error: %s", err.Error())
		return nil, err
	}

	allDms, err := dms.ExtractDmsInstances(allPages)
	if err != nil {
		logs.Logger.Errorf("Extract dms instances pages error: %s", err.Error())
		return nil, err
	}

	return &allDms, nil
}

func getAllDmsQueue(c *Config) (*[]queues.Queue, error) {
	client, err := openstack.NewDMSServiceV1(c.HwClient, golangsdk.EndpointOpts{
		Region: c.Region,
	})
	if err != nil {
		return nil, err
	}

	allPages, err := queues.List(client, false).AllPages()
	if err != nil {
		logs.Logger.Errorf("List dms queues error: %s", err.Error())
		return nil, err
	}

	allQueues, err := queues.ExtractQueues(allPages)
	if err != nil {
		logs.Logger.Errorf("Extract dms queues pages error: %s", err.Error())
		return nil, err
	}

	return &allQueues, nil
}

func getAllPublicIp(c *Config) (*[]publicips.PublicIP, error) {
	client, err := openstack.NewVPCV1(c.HwClient, golangsdk.EndpointOpts{
		Region: c.Region,
	})
	if err != nil {
		return nil, err
	}

	allPages, err := publicips.List(client, publicips.ListOpts{
		Limit: 1000,
	}).AllPages()

	if err != nil {
		logs.Logger.Errorf("List public ips error: %s", err.Error())
		return nil, err
	}
	publicipList, err1 := publicips.ExtractPublicIPs(allPages)

	if err1 != nil {
		logs.Logger.Errorf("Extract public ips pages error: %s", err.Error())
		return nil, err
	}

	return &publicipList, nil
}

func getAllBandwidth(c *Config) (*[]bandwidths.BandWidth, error) {
	client, err := openstack.NewVPCV1(c.HwClient, golangsdk.EndpointOpts{
		Region: c.Region,
	})
	if err != nil {
		return nil, err
	}

	allPages, err := bandwidths.List(client, bandwidths.ListOpts{
		Limit: 1000,
	}).AllPages()
	if err != nil {
		logs.Logger.Errorf("List bandwidths error: %s", err.Error())
		return nil, err
	}

	result, err := bandwidths.ExtractBandWidths(allPages)
	if err != nil {
		logs.Logger.Errorf("Extract bandwidths all pages error: %s", err.Error())
		return nil, err
	}

	return &result, nil
}

func getAllVolume(c *Config) (*[]volumes.Volume, error) {
	client, err := openstack.NewBlockStorageV2(c.HwClient, golangsdk.EndpointOpts{
		Region: c.Region,
	})
	if err != nil {
		return nil, err
	}

	allPages, err := volumes.List(client, volumes.ListOpts{
		Limit: 1000,
	}).AllPages()
	if err != nil {
		logs.Logger.Errorf("List volumes error: %s", err.Error())
		return nil, err
	}

	result, err := volumes.ExtractVolumes(allPages)
	if err != nil {
		logs.Logger.Errorf("Extract volumes all pages error: %s", err.Error())
		return nil, err
	}

	return &result, nil
}

func getAllServer(c *Config) (*[]servers.Server, error) {
	client, err := openstack.NewComputeV2(c.HwClient, golangsdk.EndpointOpts{
		Region: c.Region,
	})
	if err != nil {
		return nil, err
	}

	allPages, err := servers.List(client, servers.ListOpts{
		Limit: 1000,
	}).AllPages()
	if err != nil {
		logs.Logger.Errorf("List servers error: %s", err.Error())
		return nil, err
	}

	result, err := servers.ExtractServers(allPages)
	if err != nil {
		logs.Logger.Errorf("Extract servers all pages error: %s", err.Error())
		return nil, err
	}

	return &result, nil
}

func getAllGroup(c *Config) (*[]groups.Group, error) {
	client, err := openstack.NewAutoScalingV1(c.HwClient, golangsdk.EndpointOpts{
		Region: c.Region,
	})
	if err != nil {
		return nil, err
	}

	allPages, err := groups.List(client, groups.ListOpts{}).AllPages()
	if err != nil {
		logs.Logger.Errorf("List groups error: %s", err.Error())
		return nil, err
	}

	result, err := (allPages.(groups.GroupPage)).Extract()
	if err != nil {
		logs.Logger.Errorf("Extract groups all pages error: %s", err.Error())
		return nil, err
	}

	return &result, nil
}

func getAllFunction(c *Config) (*function.FunctionList, error) {
	client, err := openstack.NewFGSV2(c.HwClient, golangsdk.EndpointOpts{
		Region: c.Region,
	})
	if err != nil {
		return nil, err
	}

	allPages, err := function.List(client, function.ListOpts{}).AllPages()
	if err != nil {
		logs.Logger.Errorf("List function error: %s", err.Error())
		return nil, err
	}

	result, err := function.ExtractList(allPages)
	if err != nil {
		logs.Logger.Errorf("Extract function all pages error: %s", err.Error())
		return nil, err
	}

	return &result, nil
}
