package collector

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	v3 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	"gopkg.in/yaml.v3"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var MIPId = map[string]string{ // RDS和DDS的id-ip映射关系
	"62f6158bdd714ed684044140331521abin03": "10.160.56.8",
	"7ada0963b132406495912dbbc6969a06in03": "10.160.56.3",
	"56e9446e378e43a3ac3e20368b8d1182in03": "10.160.56.4",
	"f0ef6b89407a4d4abce3e61abe29881din03": "10.160.56.5",
	"cec1b10c027a41af92bf59e50a999e51in03": "10.160.56.106",
	"dcbeecadc24f4081a9a67aa9d19b2967in03": "10.160.56.6",
	"b4adc76de01e434184b7c33589eb1c82in03": "10.160.56.2",
	"91ebe0379bbc48b6b788dca21056c3b5in03": "10.160.56.108",
	"afc468c79a1842c585ddb52818c8efa2in03": "10.160.56.107",
	"21deb09156bf4f9ab7e748f90855786din01": "10.160.22.2",
	"3b29f36be92543d18c4ee5d8b034b234in03": "10.160.56.7",
	"b704051649864bae9de1494e632dc253in03": "10.160.56.9",
	"c85c01a2224f4d0cba00c7336328fec1in03": "10.160.56.109",
	"683ce4d041904ebe88730af8a643f9f7no02": "10.160.59.3",
	"2e55193102804a64b7d1e505baaff565no02": "10.160.59.4",
	"4565ca156f1e4d9982d847e4bc28f598no02": "10.160.59.2",
}

type CloudAuth struct {
	ProjectName string `yaml:"project_name"`
	ProjectID   string `yaml:"project_id"`
	DomainName  string `yaml:"domain_name"`
	// 建议您优先使用ReadMe文档中 4.1章节的方式，使用脚本将AccessKey,SecretKey解密后传入，避免在配置文件中明文配置AK SK导致信息泄露
	AccessKey string `yaml:"access_key"`
	Region    string `yaml:"region"`
	SecretKey string `yaml:"secret_key"`
	AuthURL   string `yaml:"auth_url"`
}

type Global struct {
	Port                        string `yaml:"port"`
	Prefix                      string `yaml:"prefix"`
	MetricPath                  string `yaml:"metric_path"`
	EpsInfoPath                 string `yaml:"eps_path"`
	MaxRoutines                 int    `yaml:"max_routines"`
	ScrapeBatchSize             int    `yaml:"scrape_batch_size"`
	ResourceSyncIntervalMinutes int    `yaml:"resource_sync_interval_minutes"`
	EpIds                       string `yaml:"ep_ids"`
	MetricsConfPath             string `yaml:"metrics_conf_path"`
	LogsConfPath                string `yaml:"logs_conf_path"`
	EndpointsConfPath           string `yaml:"endpoints_conf_path"`
	IgnoreSSLVerify             bool   `yaml:"ignore_ssl_verify"`
}

type CloudConfig struct {
	Auth   CloudAuth `yaml:"auth"`
	Global Global    `yaml:"global"`
}

var CloudConf CloudConfig
var SecurityMod bool
var TmpAK string
var TmpSK string

func InitCloudConf(file string) error {
	realPath, err := NormalizePath(file)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(realPath)
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

func NormalizePath(path string) (string, error) {
	relPath, err := filepath.Abs(path) // 对文件路径进行标准化
	if err != nil {
		return "", err
	}
	relPath = strings.Replace(relPath, "\\", "/", -1)
	match, err := regexp.MatchString("[!;<>&|$\n`\\\\]", relPath)
	if match || err != nil {
		return "", errors.New("match path error")
	}
	return relPath, nil
}

func SetDefaultConfigValues(config *CloudConfig) {
	if config.Global.Port == "" {
		config.Global.Port = ":8087"
	}

	if config.Global.MetricPath == "" {
		config.Global.MetricPath = "/metrics"
	}

	if config.Global.EpsInfoPath == "" {
		config.Global.EpsInfoPath = "/eps-info"
	}

	if config.Global.Prefix == "" {
		config.Global.Prefix = "huaweicloud"
	}

	if config.Global.MaxRoutines == 0 {
		config.Global.MaxRoutines = 5
	}

	if config.Global.ScrapeBatchSize == 0 {
		config.Global.ScrapeBatchSize = 300
	}

	if config.Global.ResourceSyncIntervalMinutes <= 0 {
		config.Global.ResourceSyncIntervalMinutes = 180
	}

	if config.Global.MetricsConfPath == "" {
		config.Global.MetricsConfPath = "./metric.yml"
	}

	if config.Global.LogsConfPath == "" {
		config.Global.LogsConfPath = "./logs.yml"
	}

	if config.Global.EndpointsConfPath == "" {
		config.Global.EndpointsConfPath = "./endpoints.yml"
	}
}

type MetricConf struct {
	Resource      string              `yaml:"resource"`
	DimMetricName map[string][]string `yaml:"dim_metric_name"`
}

var metricConf map[string]MetricConf

func InitMetricConf() error {
	metricConf = make(map[string]MetricConf)
	data, err := ioutil.ReadFile(CloudConf.Global.MetricsConfPath)
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
	Region           string
	ProjectID        string
	ProjectName      string
	UserID           string
}

var conf = &Config{}

func InitConfig() error {
	conf.IdentityEndpoint = CloudConf.Auth.AuthURL
	conf.ProjectName = CloudConf.Auth.ProjectName
	conf.ProjectID = CloudConf.Auth.ProjectID
	conf.DomainName = CloudConf.Auth.DomainName
	conf.Region = CloudConf.Auth.Region
	// 安全模式下，ak/sk通过用户交互获取，避免明文方式存在于存储介质中
	if SecurityMod {
		conf.AccessKey = TmpAK
		conf.SecretKey = TmpSK
	} else {
		conf.AccessKey = CloudConf.Auth.AccessKey
		conf.SecretKey = CloudConf.Auth.SecretKey
	}

	if conf.ProjectID == "" && conf.ProjectName == "" {
		fmt.Printf("Init config error: ProjectID or ProjectName must setting.")
		return errors.New("init config error: ProjectID or ProjectName must setting")
	}
	req, err := http.NewRequest("GET", conf.IdentityEndpoint, nil)
	if err != nil {
		fmt.Printf("Auth url is invalid.")
		return err
	}
	host = req.Host

	if conf.ProjectID == "" {
		resp, err := getProjectInfo()
		if err != nil {
			fmt.Printf("Get project info error: %s", err.Error())
			return err
		}
		if len(*resp.Projects) == 0 {
			fmt.Printf("Project info is empty")
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
			WithHttpConfig(config.DefaultHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
			Build())
	return iamclient.KeystoneListProjects(&model.KeystoneListProjectsRequest{Name: &conf.ProjectName})
}

var endpointConfig map[string]string

func InitEndpointConfig(path string) {
	realPath, err := NormalizePath(path)
	if err != nil {
		logs.Logger.Errorf("Normalize endpoint config err: %s", err.Error())
		return
	}

	context, err := ioutil.ReadFile(realPath)
	if err != nil {
		logs.Logger.Infof("Invalid endpoint config path, default config will be used instead")
		return
	}
	err = yaml.Unmarshal(context, &endpointConfig)
	if err != nil {
		logs.Logger.Errorf("Init endpoint config error: %s", err.Error())
	}
}
