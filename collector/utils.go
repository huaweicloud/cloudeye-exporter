package collector

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
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

func NewCloudConfigFromFile(file string) (*CloudConfig, error) {
	var config CloudConfig

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	SetDefaultConfigValues(&config)

	return &config, err
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

var filterConfigMap map[string]map[string][]string

func InitFilterConfig(enable bool) error {
	filterConfigMap = make(map[string]map[string][]string)
	if !enable {
		return nil
	}

	data, err := ioutil.ReadFile("metric_filter_config.yml")
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &filterConfigMap)
	if err != nil {
		return err
	}
	return nil
}

func getMetricConfigMap(namespace string) map[string][]string {
	if configMap, ok := filterConfigMap[namespace]; ok {
		return configMap
	}
	return nil
}
