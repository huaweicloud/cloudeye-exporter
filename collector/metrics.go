// Copyright 2019 HuaweiCloud.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"fmt"
	"net/http"

	"github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/openstack/ces/v1/metrics"
	"github.com/huaweicloud/golangsdk/openstack/ces/v1/metricdata"
	"github.com/huaweicloud/golangsdk/openstack"
	"github.com/huaweicloud/golangsdk/openstack/networking/v2/extensions/lbaas_v2/loadbalancers"
	"github.com/huaweicloud/golangsdk/openstack/networking/v2/extensions/lbaas_v2/listeners"
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
	err := fmt.Errorf("Must config token or aksk or username password to be authorized")

	if c.AccessKey != "" && c.SecretKey != "" {
		err = buildClientByAKSK(c)
	} else if c.Password != "" && (c.Username != "" || c.UserID != "") {
		err = buildClientByPassword(c)
	}

	if err != nil {
		return err
	}

	return nil
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


func InitConfig(config *CloudConfig)(*Config, error)  {
	auth := config.Auth
	configOptions := Config{
		IdentityEndpoint: auth.AuthURL,
		TenantName:      auth.ProjectName,
		AccessKey:       auth.AccessKey,
		SecretKey:       auth.SecretKey,
		DomainName:      auth.DomainName,
		Username:        auth.UserName,
		Region:          auth.Region,
		Password:        auth.Password,
		Insecure:        true,
	}

	err := buildClient(&configOptions)
	if err != nil {
		fmt.Println("Failed to build client: ", err)
		return nil, err
	}

	return &configOptions, err
}


func getCESClient(c *Config)(*golangsdk.ServiceClient, error)  {
	client, clientErr := openstack.NewCESV1(c.HwClient, golangsdk.EndpointOpts{
		Region: c.Region,
	})
	if clientErr != nil {
		fmt.Println("Failed to get the NewCESV1 client: ", clientErr)
		return nil, clientErr
	}

	return client, nil
}


func getELBlient(c *Config)(*golangsdk.ServiceClient, error)  {
	client, clientErr := openstack.NewNetworkV2(c.HwClient, golangsdk.EndpointOpts{
		Region: c.Region,
	})
	if clientErr != nil {
		fmt.Println("Failed to get the NewLoadBalancerV2 client: ", clientErr)
		return nil, clientErr
	}

	return client, nil
}


func getDimByDimension(num int, dimensions *[]metrics.Dimension) (string){
	dim := ""
	if len(*dimensions) > num {
		dim = (*dimensions)[num].Name + "," + (*dimensions)[num].Value
	}

	return dim
}


func getMetricData(
	c *Config,
	metric *metrics.Metric,
	dimensions *[]metrics.Dimension,
	from string,
	to string) (
	*[]metricdata.Datapoint, error){

	options := metricdata.GetOpts {
		Namespace: metric.Namespace,
		Dim0: getDimByDimension(0, dimensions),
		Dim1: getDimByDimension(1, dimensions),
		Dim2: getDimByDimension(2, dimensions),
		MetricName: metric.MetricName,
		From: from,
		To: to,
		Period: "1",
		Filter: "average",
	}

	client, err := getCESClient(c)
	if err != nil {
		fmt.Println("Failed to get client: ", err)
		return nil, err
	}

	v, err := metricdata.Get(client, options).Extract()
	if err != nil {
		fmt.Println("Failed to get metricdata: ", err)
		return nil, err
	}

	return &v.Datapoints, nil
}


func getAllMetric(client *Config, namespace string) (*[]metrics.Metric, error){
	limit := 100
	mopts := metrics.ListOpts{
		Limit: &limit,
		Start: "",
		Namespace: namespace,
	}

	c, err :=getCESClient(client)
	if err != nil {
		fmt.Println("get all metric client,%s", err)
		return nil, err
	}

	allpage, err := metrics.List(c, mopts).AllPages()
	if err != nil {
		fmt.Println("get all metric all pages error,%s", err)
		return nil, err
	}

	v, err := metrics.ExtractAllPagesMetrics(allpage)
	if err != nil {
		fmt.Println("get all metric pages error,%s", err)
		return nil, err
	}

	return &v.Metrics, nil
}


func getAllELB(client *Config) (*[]loadbalancers.LoadBalancer, error)  {
	c, err :=getELBlient(client)
	if err != nil {
		return nil, err
	}

	allPages, err := loadbalancers.List(c, loadbalancers.ListOpts{}).AllPages()
	if err != nil {
		fmt.Println("get loadbalancers all pages error,%s", err)
		return nil, err
	}

	allLoadbalancers, err := loadbalancers.ExtractLoadBalancers(allPages)
	if err != nil {
		fmt.Println("get loadbalancers pages error,%s", err)
		return nil, err
	}

	return &allLoadbalancers, nil
}


func getAllListener(client *Config) (*[]listeners.Listener, error)  {
	c, err :=getELBlient(client)
	if err != nil {
		return nil, err
	}

	allPages, err := listeners.List(c, listeners.ListOpts{}).AllPages()
	if err != nil {
		fmt.Println("get all listener all pages error,%s", err)
		return nil, err
	}

	allListeners, err := listeners.ExtractListeners(allPages)
	if err != nil {
		fmt.Println("get all listener all pages error,%s", err)
		return nil, err
	}

	return &allListeners, nil
}
