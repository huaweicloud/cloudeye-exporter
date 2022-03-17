package collector

import (
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	"gopkg.in/yaml.v2"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

type CloudAuth struct {
	ProjectName string `yaml:"project_name"`
	ProjectID   string `yaml:"project_id"`
	DomainName  string `yaml:"domain_name"`
	AccessKey   string `yaml:"access_key"`
	Region      string `yaml:"region"`
	SecretKey   string `yaml:"secret_key"`
	AuthURL     string `yaml:"auth_url"`
	UserName    string `yaml:"user_name"`
	Password    string `yaml:"password"`
}

type Global struct {
	Port            string `yaml:"port"`
	Prefix          string `yaml:"prefix"`
	MetricPath      string `yaml:"metric_path"`
	MaxRoutines     int    `yaml:"max_routines"`
	ScrapeBatchSize int    `yaml:"scrape_batch_size"`
}

type CloudConfig struct {
	Auth   CloudAuth `yaml:"auth"`
	Global Global    `yaml:"global"`
}

var CloudConf CloudConfig

func InitCloudConf(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &CloudConf)
	if err != nil {
		return err
	}

	SetDefaultConfigValues(&CloudConf)

	err = InitConfig()
	if err != nil {
		return err
	}

	return err
}

func SetDefaultConfigValues(config *CloudConfig) {
	if config.Global.Port == "" {
		config.Global.Port = ":8087"
	}

	if config.Global.MetricPath == "" {
		config.Global.MetricPath = "/metrics"
	}

	if config.Global.Prefix == "" {
		config.Global.Prefix = "huaweicloud"
	}

	if config.Global.MaxRoutines == 0 {
		config.Global.MaxRoutines = 20
	}

	if config.Global.ScrapeBatchSize == 0 {
		config.Global.ScrapeBatchSize = 10
	}
}

type MetricConf struct {
	Resource      string              `yaml:"resource"`
	DimMetricName map[string][]string `yaml:"dim_metric_name"`
}

var metricConf map[string]MetricConf

func InitMetricConf() error {
	metricConf = make(map[string]MetricConf)
	data, err := ioutil.ReadFile("metric.yml")
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, &metricConf)
}

func getMetricConfigMap(namespace string) map[string][]string {
	if conf, ok := metricConf[namespace]; ok {
		return conf.DimMetricName
	}
	return nil
}

func getResourceFromRMS(namespace string) bool {
	if conf, ok := metricConf[namespace]; ok {
		return conf.Resource == "RMS" || conf.Resource == "rms"
	}
	return false
}

type Config struct {
	AccessKey        string
	SecretKey        string
	DomainID         string
	DomainName       string
	EndpointType     string
	IdentityEndpoint string
	Password         string
	Region           string
	ProjectID        string
	ProjectName      string
	Token            string
	Username         string
	UserID           string
}

var conf = &Config{}

func InitConfig() error {
	conf.IdentityEndpoint = CloudConf.Auth.AuthURL
	conf.ProjectName = CloudConf.Auth.ProjectName
	conf.ProjectID = CloudConf.Auth.ProjectID
	conf.AccessKey = CloudConf.Auth.AccessKey
	conf.SecretKey = CloudConf.Auth.SecretKey
	conf.DomainName = CloudConf.Auth.DomainName
	conf.Username = CloudConf.Auth.UserName
	conf.Region = CloudConf.Auth.Region
	conf.Password = CloudConf.Auth.Password
	if conf.ProjectID == "" && conf.ProjectName == "" {
		logs.Logger.Error("Init config error: ProjectID or ProjectName must setting.")
		return errors.New("init config error: ProjectID or ProjectName must setting")
	}
	req, err := http.NewRequest("GET", conf.IdentityEndpoint, nil)
	if err != nil {
		logs.Logger.Error("Auth url is invalid.")
		return err
	}
	host = req.Host

	if conf.ProjectID == "" {
		resp, err := getProjectInfo()
		if err != nil {
			logs.Logger.Errorf("Get project info error: %s", err.Error())
			return err
		}
		if len(*resp.Projects) == 0 {
			logs.Logger.Error("project info is empty")
			return errors.New("project info is empty")
		}

		projects := *resp.Projects
		conf.ProjectID = projects[0].Id
		conf.DomainID = projects[0].DomainId
	}
	return nil
}

func getProjectInfo() (*model.KeystoneListProjectsResponse, error) {
	iamclient := v3.NewIamClient(
		v3.IamClientBuilder().
			WithEndpoint(conf.IdentityEndpoint).
			WithCredential(
				global.NewCredentialsBuilder().
					WithAk(conf.AccessKey).
					WithSk(conf.SecretKey).
					Build()).
			WithHttpConfig(config.DefaultHttpConfig().
				WithIgnoreSSLVerification(true)).
			Build())
	return iamclient.KeystoneListProjects(&model.KeystoneListProjectsRequest{Name: &conf.ProjectName})
}
