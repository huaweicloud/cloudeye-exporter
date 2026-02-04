package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/huaweicloud/cloudeye-exporter/collector"
	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var (
	clientConfig = flag.String("config", "./clouds.yml", "Path to the cloud configuration file")
	// 安全模式，从用户交互输入获取ak/sk，避免明文ak/sk敏感信息存储在配置文件中
	securityMod = flag.Bool("s", false, "Get ak sk from command line")
	// 以https协议启动cloudeye-exporter，需要从用户交互输入获取ca证书路径，服务端https证书路径，服务端私钥路径以及私钥密码
	httpsEnabled = flag.Bool("k", false, "Start the cloudeye exporter service in https mode")
	getVersion   = flag.Bool("v", false, "Get version from command line")
	proxyEnabled = flag.Bool("p", false, "Start the cloudeye exporter service and start the proxy")
	authMode     = flag.String("auth_mode", collector.AuthModePermanentAkSk, "Specify the authentication mode when cloudeye exporter query resources and metrics")

	ak, sk, proxyUserName, proxyPassword string
)

func handler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("services")
	if target == "" {
		http.Error(w, "'target' parameter must be specified", 400)
		return
	}

	targets := strings.Split(target, ",")
	if len(targets) > collector.MaxNamespacesCount {
		http.Error(w, "namespaces not allowed to exceed 1000", 400)
		return
	}
	registry := prometheus.NewRegistry()
	logs.Logger.Infof("Start to monitor services: %s", targets)
	exporter := collector.GetMonitoringCollector(targets)
	registry.MustRegister(exporter)
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	logs.Logger.Infof("End to monitor services: %s", targets)
}

func epHandler(w http.ResponseWriter, r *http.Request) {
	epsInfo, err := collector.GetEPSInfo()
	if err != nil {
		http.Error(w, fmt.Sprintf("get eps info error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	_, err = w.Write([]byte(epsInfo))
	if err != nil {
		logs.Logger.Errorf("Response to caller error: %s", err.Error())
	}
}

func getAkSkFromCommandLine() {
	if *securityMod {
		collector.SecurityMod = *securityMod
		// 用户交互输入ak/sk，避免明文配置敏感信息
		fmt.Print("Please input ak&sk split with space(eg: {example_ak example_sk}): \n")
		_, err := fmt.Scanln(&ak, &sk)
		if err != nil {
			fmt.Printf("Read ak sk error: %s", err.Error())
			return
		}
		collector.TmpAK = ak
		collector.TmpSK = sk
	}
}

func getProxyInfoFromCommandLine() {
	if *proxyEnabled {
		collector.ProxyEnabled = *proxyEnabled
		// 用户交互输入代理userName/password，避免明文配置敏感信息
		fmt.Print("Please input proxy userName&proxy password split with space(eg: {example_proxy_user_name example_proxy_password}): \n")
		_, err := fmt.Scanln(&proxyUserName, &proxyPassword)
		if err != nil {
			fmt.Printf("Read proxy info error: %s", err.Error())
			return
		}
		collector.TmpProxyUserName = proxyUserName
		collector.TmpProxyPassword = proxyPassword
	}
}

func getHttpsEnabledFromCommandLine() {
	if *httpsEnabled {
		collector.HttpsEnabled = *httpsEnabled
	}
}

func getAuthModeFromCommandLine() {
	collector.AuthMode = *authMode
}

func getVersionFunc() {
	if *getVersion {
		fmt.Printf("Cloudeye-exporter version: %s", collector.Version)
		os.Exit(0)
	}
}

func main() {
	flag.Parse()
	getVersionFunc()
	getAkSkFromCommandLine()
	getHttpsEnabledFromCommandLine()
	getProxyInfoFromCommandLine()
	getAuthModeFromCommandLine()
	initConf()
	initCache()

	http.HandleFunc(collector.CloudConf.Global.MetricPath, handler)
	http.HandleFunc(collector.CloudConf.Global.EpsInfoPath, epHandler)
	collector.StartServer()
}

func initConf() {
	err := collector.InitCloudConf(*clientConfig)
	if err != nil {
		fmt.Printf("Init Cloud Config From File error: %s", err.Error())
		os.Exit(1)
	}

	logs.InitLog(collector.CloudConf.Global.LogsConfPath)
	err = collector.InitMetricConf()
	if err != nil {
		logs.Logger.Errorf("Init metric Config error: %s", err.Error())
		logs.FlushLogAndExit(1)
	}

	err = collector.InitUnitTransformConfig()
	if err != nil {
		logs.Logger.Errorf("Failed to init unit transform config: %s", err.Error())
	}

	collector.InitEndpointConfig(collector.CloudConf.Global.EndpointsConfPath)
}

func initCache() {
	collector.GetAgentDimensionRefresher().RefreshAgentDimensionWithInterval()
}
