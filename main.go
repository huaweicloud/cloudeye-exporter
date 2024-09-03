package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/huaweicloud/cloudeye-exporter/collector"
	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var (
	clientConfig = flag.String("config", "./clouds.yml", "Path to the cloud configuration file")
	// 安全模式，从用户交互输入获取ak/sk，避免明文ak/sk敏感信息存储在配置文件中
	securityMod = flag.Bool("s", false, "Get ak sk from command line")
	getVersion  = flag.Bool("v", false, "Get version from command line")
	ak, sk      string
)

func addIdIpMappingMetrics() *prometheus.GaugeVec {
	gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "huaweicloud_instance_info",
		Help: "华为云RDS，DDS实例 内网ip--实例id 映射表",
	},
		[]string{"instance_id", "ip"},
	)

	i := 1
	for id, ip := range collector.MIPId {
		gaugeVec.With(prometheus.Labels{"instance_id": id, "ip": ip}).Set(float64(i))
		i++
	}

	return gaugeVec
}

func handler(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")
	target := r.URL.Query().Get("services")
	if target == "" && service == "" {
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

	if service == "INFO.ID-IP" { // service 参数的值为 INFO 返回id-ip映射表
		infoExporter := addIdIpMappingMetrics()
		registry.MustRegister(infoExporter)
	}

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
		fmt.Print("Please input ak&sk split with space: (eg: {example_ak example_sk})")
		_, err := fmt.Scanln(&ak, &sk)
		if err != nil {
			fmt.Printf("Read ak sk error: %s", err.Error())
			return
		}
		collector.TmpAK = ak
		collector.TmpSK = sk
	}
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
	initConf()

	http.HandleFunc(collector.CloudConf.Global.MetricPath, handler)
	http.HandleFunc(collector.CloudConf.Global.EpsInfoPath, epHandler)
	server := &http.Server{
		Addr:         collector.CloudConf.Global.Port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	logs.Logger.Info("Start server at ", collector.CloudConf.Global.Port)
	if err := server.ListenAndServe(); err != nil {
		logs.Logger.Errorf("Error occur when start server %s", err.Error())
		logs.FlushLogAndExit(1)
	}
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
	collector.InitEndpointConfig(collector.CloudConf.Global.EndpointsConfPath)
}
