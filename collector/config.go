package collector

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	v3 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	"gopkg.in/yaml.v3"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

type Oidc struct {
	IdTokenFilePath string `yaml:"id_token_file_path"`
	IdpId           string `yaml:"idp_id"`
	DomainID        string `yaml:"domain_id"`
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
	OidcInfo  *Oidc  `yaml:"oidc"`
}

type Global struct {
	Port                        string `yaml:"port"`
	Prefix                      string `yaml:"prefix"`
	MetricPath                  string `yaml:"metric_path"`
	EpsInfoPath                 string `yaml:"eps_path"`
	MaxRoutines                 int    `yaml:"max_routines"`
	ScrapeBatchSize             int    `yaml:"scrape_batch_size"`
	ResourceSyncIntervalMinutes int    `yaml:"resource_sync_interval_minutes"`
	MetricInfoExpirationDays    int    `yaml:"metric_info_expiration_days"`
	MetricInfoCleanThreshold    int    `yaml:"metric_info_clean_threshold"`
	EpIds                       string `yaml:"ep_ids"`
	MetricsConfPath             string `yaml:"metrics_conf_path"`
	LogsConfPath                string `yaml:"logs_conf_path"`
	EndpointsConfPath           string `yaml:"endpoints_conf_path"`
	IgnoreSSLVerify             bool   `yaml:"ignore_ssl_verify"`
	RmsRetryTimes               int    `yaml:"rms_retry_times"`

	// 用户配置的proxy信息
	HttpSchema string `yaml:"proxy_schema"`
	HttpHost   string `yaml:"proxy_host"`
	HttpPort   int    `yaml:"proxy_port"`
	UserName   string `yaml:"proxy_username"`
	Password   string `yaml:"proxy_password"`

	// CN列表，用于校验https证书链中的DNS名称
	ClientCN string `yaml:"client_cn"`

	UnitStandardizationEnabled   bool   `yaml:"unit_standardization_enabled"`
	I18nConfigFilePath           string `yaml:"i18n_config_file_path"`
	UnitStandardizationFilePath  string `yaml:"unit_standardization_file_path"`
	MetricTimestampExportEnabled bool   `yaml:"metric_timestamp_export_enabled"`
	MetricQueryDuration          int    `yaml:"metric_query_duration"`
}

type CloudConfig struct {
	Auth   CloudAuth `yaml:"auth"`
	Global Global    `yaml:"global"`
}

var CloudConf CloudConfig
var SecurityMod bool
var HttpsEnabled bool
var ProxyEnabled bool
var TmpAK string
var TmpSK string
var TmpProxyUserName string
var TmpProxyPassword string
var AuthMode string

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

	if config.Global.MetricInfoExpirationDays <= 2 {
		config.Global.MetricInfoExpirationDays = 2
	}

	if config.Global.MetricInfoExpirationDays >= 20 {
		config.Global.MetricInfoExpirationDays = 20
	}

	if config.Global.MetricInfoCleanThreshold <= 5000 {
		config.Global.MetricInfoCleanThreshold = 5000
	}

	if config.Global.MetricInfoCleanThreshold >= 50000 {
		config.Global.MetricInfoCleanThreshold = 50000
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

	if config.Global.RmsRetryTimes <= 0 {
		config.Global.RmsRetryTimes = 1
	}

	if config.Global.RmsRetryTimes > 10 {
		config.Global.RmsRetryTimes = 10
	}

	if config.Global.I18nConfigFilePath == "" {
		config.Global.I18nConfigFilePath = "./i18n.json"
	}

	if config.Global.UnitStandardizationFilePath == "" {
		config.Global.UnitStandardizationFilePath = "./unit_standard_transform.json"
	}

	if config.Global.MetricQueryDuration < 10 || config.Global.MetricQueryDuration > 60 {
		config.Global.MetricQueryDuration = 10
	}
}

type MetricConf struct {
	Resource      string              `yaml:"resource"`
	DimMetricName map[string][]string `yaml:"dim_metric_name"`
}

var metricConf map[string]MetricConf

func InitMetricConf() error {
	metricConf = make(map[string]MetricConf)
	realPath, err := NormalizePath(CloudConf.Global.MetricsConfPath)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadFile(realPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, &metricConf)
	// 对指标去重
	if err == nil {
		for _, metricConfigs := range metricConf {
			for key, value := range metricConfigs.DimMetricName {
				// 调用去重函数
				uniqueValue := uniqueStrings(value)
				// 更新map中的值
				metricConfigs.DimMetricName[key] = uniqueValue
			}
		}
	}
	return err
}

// uniqueStrings 去重函数，返回去重后的字符串数组
func uniqueStrings(arr []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, str := range arr {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}
	return result
}

type UnitTransformItem struct {
	MetricName string `json:"metric_name"`
	Unit       string `json:"unit"`
	UnitV2     string `json:"unit_v2"`
}

type UnitTransformConf struct {
	Namespace string              `json:"namespace"`
	Units     []UnitTransformItem `json:"units"`
}

var (
	i18nConfigMap        = make(map[string]string)
	unitTransformConfMap = make(map[string]UnitTransformItem)
)
var keyPatternForUnitUnifyI18n = "%s.%s.unit"
var keyPatternForUnitStandardization = "%s.%s.%s"

func InitUnitTransformConfig() error {
	if !CloudConf.Global.UnitStandardizationEnabled {
		return nil
	}
	err := InitI18nConfig()
	if err != nil {
		return err
	}

	err = InitUnitStandardizationConfig()
	if err != nil {
		return err
	}

	return nil
}

func InitI18nConfig() error {
	resultMap, err := getI18nConfigFromServer()
	if err == nil {
		i18nConfigMap = resultMap
		return nil
	}
	// 从ces服务端拉取i18n配置数据失败，降级从本地文件加载指标单位国际化信息
	logs.Logger.Errorf("Get i18n config from ces server failed: %s, i18n config will load from local files", err.Error())
	realPath, err := NormalizePath(CloudConf.Global.I18nConfigFilePath)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadFile(realPath)
	if err != nil {
		return err
	}
	i18nInfoMap := make(map[string]string)
	err = json.Unmarshal(data, &i18nInfoMap)
	if err != nil {
		return err
	}
	for key, unit := range i18nInfoMap {
		unit = strings.TrimSpace(unit)
		i18nConfigMap[key] = unit
	}
	return nil
}

type GetI18nInfoRequest struct {
}

type GetI18nResponse struct {
	HttpStatusCode int           `json:"-"`
	Body           io.ReadCloser `json:"-" type:"stream"`
}

func getReqDefForListI18n() *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().WithMethod(http.MethodGet).WithPath("/v1/{project_id}/ns-i18n").
		WithResponse(new(GetI18nResponse)).WithContentType("application/json")
	return reqDefBuilder.Build()
}

func getI18nConfigFromServer() (map[string]string, error) {
	// 拉取英文语种的i18n配置信息
	hcClient := getCESClient().HcClient.PreInvoke(map[string]string{"x-language": "en-us"})
	resp, err := hcClient.Sync(&GetI18nInfoRequest{}, getReqDefForListI18n())
	if err != nil {
		logs.Logger.Errorf("List i18 config error: %s", err.Error())
		return nil, err
	}
	realResp, ok := resp.(*GetI18nResponse)
	if !ok {
		logs.Logger.Errorf("Convert response to GetI18nResponse failed")
		return nil, fmt.Errorf("convert response to GetI18nResponse failed")
	}
	if realResp.HttpStatusCode >= 400 {
		logs.Logger.Errorf("List i18 config failed with http code: %d", realResp.HttpStatusCode)
		return nil, fmt.Errorf("list i18 config failed with http code: %d", realResp.HttpStatusCode)
	}

	defer realResp.Body.Close()
	allData, err := io.ReadAll(realResp.Body)
	if err != nil {
		logs.Logger.Errorf("Read content from i18n response body failed: %s", err.Error())
		return nil, err
	}

	allI18nInfoMap := make(map[string]interface{})
	err = json.Unmarshal(allData, &allI18nInfoMap)
	if err != nil {
		logs.Logger.Errorf("Unmarshal i18n response content to map info error: %s", err.Error())
		return nil, err
	}
	// 只取指标单位相关i18n信息
	unitI18nInfoMap := make(map[string]string)
	for key, value := range allI18nInfoMap {
		if !strings.HasSuffix(key, "unit") {
			continue
		}
		if unit, isConvertOk := value.(string); isConvertOk {
			unitI18nInfoMap[key] = strings.TrimSpace(unit)
		}
	}
	return unitI18nInfoMap, nil
}

func InitUnitStandardizationConfig() error {
	var unitTransformConfs []UnitTransformConf
	realPath, err := NormalizePath(CloudConf.Global.UnitStandardizationFilePath)
	if err != nil {
		logs.Logger.Errorf("Parse unit standardization file failed, and the conversion function from i18n unit to standard unit will be disabled: %s", err.Error())
		return err
	}
	data, err := ioutil.ReadFile(realPath)
	if err != nil {
		logs.Logger.Errorf("Read unit standardization file failed, and the conversion function from i18n unit to standard unit will be disabled: %s", err.Error())
		return err
	}
	err = json.Unmarshal(data, &unitTransformConfs)
	if err != nil {
		logs.Logger.Errorf("Parse unit standardization file failed, and the conversion function from i18n unit to standard unit will be disabled: %s", err.Error())
		return err
	}
	for _, unitTransformConf := range unitTransformConfs {
		for _, item := range unitTransformConf.Units {
			if !validateUnitTransformRule(unitTransformConf.Namespace, item) {
				logs.Logger.Errorf("Unit config error, namespace: %s, source unit: %s, unit v2: %s",
					unitTransformConf.Namespace, item.Unit, item.UnitV2)
				continue
			}
			key := fmt.Sprintf(keyPatternForUnitStandardization, unitTransformConf.Namespace, item.MetricName, item.Unit)
			unitTransformConfMap[key] = item
		}
	}
	return nil
}

/*
*

	{
	  "namespace": "SYS.NAT",
	  "units": [
	    # 单个UnitTransformItem结构
	    {
	      "metric_name": "*", #不能为空，值取具体指标名或者代表所有指标(*代表所有指标名)
	      "unit": "Count",   #I18N单位（支持配置空值，此时metric_name不能为*，代表当前特定指标需要补齐单位，并在unit_v2中说明服务补充的单位信息）
	      "unit_v2": "count" #指标单位含义阐述字段，表述本指标单位的标准化信息
	    }
	  ]
	}
*/
func validateUnitTransformRule(namespace string, transformRule UnitTransformItem) bool {
	if transformRule.MetricName == "" || transformRule.UnitV2 == "" {
		logs.Logger.Errorf("Unit config error, namespace: %s, source unit: %s, unit v2: %s",
			namespace, transformRule.MetricName, transformRule.Unit, transformRule.UnitV2)
		return false
	}

	if transformRule.MetricName == "*" && transformRule.Unit == "" {
		logs.Logger.Errorf("Source unit couldn't be empty when metric_name is *, namespace: %s, source unit: %s, unit v2: %s",
			namespace, transformRule.Unit, transformRule.UnitV2)
		return false
	}
	return true
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
	AuthMode         string
	IdTokenFilePath  string
	IdpId            string
}

var conf = &Config{}
var authCredentialMap = map[string]func(serviceType string) auth.ICredential{
	AuthModePermanentAkSk: func(serviceType string) auth.ICredential {
		var credential auth.ICredential
		switch serviceType {
		case GlobalServiceType:
			credential = global.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithDomainId(conf.DomainID).Build()
		case RegionServiceType:
			credential = basic.NewCredentialsBuilder().WithAk(conf.AccessKey).WithSk(conf.SecretKey).WithProjectId(conf.ProjectID).Build()
		default:
			logs.Logger.Errorf("Invalid service type: %s", serviceType)
		}
		return credential
	},
	AuthModeOidcToken: func(serviceType string) auth.ICredential {
		var credential auth.ICredential
		switch serviceType {
		case GlobalServiceType:
			credential = global.NewCredentialsBuilder().WithIdpId(conf.IdpId).WithIdTokenFile(conf.IdTokenFilePath).WithDomainId(conf.DomainID).WithIamEndpointOverride(conf.IdentityEndpoint).Build()
		case RegionServiceType:
			credential = basic.NewCredentialsBuilder().WithIdpId(conf.IdpId).WithIdTokenFile(conf.IdTokenFilePath).WithProjectId(conf.ProjectID).WithIamEndpointOverride(conf.IdentityEndpoint).Build()
		default:
			logs.Logger.Errorf("Invalid service type: %s", serviceType)
		}
		return credential
	},
}

func InitConfig() error {
	conf.IdentityEndpoint = CloudConf.Auth.AuthURL
	conf.ProjectName = CloudConf.Auth.ProjectName
	conf.ProjectID = CloudConf.Auth.ProjectID
	conf.DomainName = CloudConf.Auth.DomainName
	conf.Region = CloudConf.Auth.Region
	if AuthMode == AuthModeOidcToken {
		conf.AuthMode = AuthModeOidcToken
		conf.DomainID = CloudConf.Auth.OidcInfo.DomainID
		conf.IdpId = CloudConf.Auth.OidcInfo.IdpId
		conf.IdTokenFilePath = CloudConf.Auth.OidcInfo.IdTokenFilePath
	} else if AuthMode == AuthModePermanentAkSk {
		conf.AuthMode = AuthModePermanentAkSk
		// 安全模式下，ak/sk通过用户交互获取，避免明文方式存在于存储介质中
		if SecurityMod {
			conf.AccessKey = TmpAK
			conf.SecretKey = TmpSK
		} else {
			conf.AccessKey = CloudConf.Auth.AccessKey
			conf.SecretKey = CloudConf.Auth.SecretKey
		}
	} else {
		fmt.Printf("Auth mode is invalid: %s", AuthMode)
		return errors.New("Auth mode is invalid")
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
			WithCredential(authCredentialMap[conf.AuthMode](GlobalServiceType)).
			WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
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
