package main

import (
	"flag"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/huaweicloud/cloudeye-exporter/collector"
	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var (
	clientConfig = flag.String("config", "./clouds.yml", "Path to the cloud configuration file")
)

func handler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("services")
	if target == "" {
		http.Error(w, "'target' parameter must be specified", 400)
		return
	}

	targets := strings.Split(target, ",")
	registry := prometheus.NewRegistry()

	logs.Logger.Infof("Start to monitor services: %s", targets)
	exporter := collector.GetMonitoringCollector(targets)
	registry.MustRegister(exporter)
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	logs.Logger.Infof("End to monitor services: %s", targets)
}

func main() {
	flag.Parse()
	logs.InitLog()
	err := collector.InitCloudConf(*clientConfig)
	if err != nil {
		logs.Logger.Error("Init Cloud Config From File error: ", err.Error())
		logs.FlushLogAndExit(1)
	}
	err = collector.InitMetricConf()
	if err != nil {
		logs.Logger.Error("Init metric Config error: ", err.Error())
		logs.FlushLogAndExit(1)
	}

	http.HandleFunc(collector.CloudConf.Global.MetricPath, handler)

	logs.Logger.Info("Start server at ", collector.CloudConf.Global.Port)
	if err := http.ListenAndServe(collector.CloudConf.Global.Port, nil); err != nil {
		logs.Logger.Errorf("Error occur when start server %s", err.Error())
		logs.FlushLogAndExit(1)
	}
}
