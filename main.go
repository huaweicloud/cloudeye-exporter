package main

import (
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/huaweicloud/cloudeye-exporter/collector"
	"github.com/huaweicloud/cloudeye-exporter/logs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	clientConfig = flag.String("config", "./clouds.yml", "Path to the cloud configuration file")
	filterEnable = flag.Bool("filter-enable", false, "Enabling monitoring metric filter")
	debug        = flag.Bool("debug", false, "If debug the code.")
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
	exporter, err := collector.GetMonitoringCollector(*clientConfig, targets)
	if err != nil {
		w.WriteHeader(500)
		_, err := w.Write([]byte(err.Error()))
		if err != nil {
			logs.Logger.Errorf("Fail to write response body, error: %s", err.Error())
			return
		}
		return
	}
	registry.MustRegister(exporter)
	if err != nil {
		logs.Logger.Errorf("Fail to start to morning services: %+v, err: %s", targets, err.Error())
		return
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	flag.Parse()
	logs.InitLog(*debug)
	config, err := collector.NewCloudConfigFromFile(*clientConfig)
	if err != nil {
		logs.Logger.Fatal("New Cloud Config From File error: ", err.Error())
		return
	}
	err = collector.InitFilterConfig(*filterEnable)
	if err != nil {
		logs.Logger.Fatal("Init filter Config error: ", err.Error())
		return
	}

	http.HandleFunc(config.Global.MetricPath, handler)

	logs.Logger.Infoln("Start server at ", config.Global.Port)
	if err := http.ListenAndServe(config.Global.Port, nil); err != nil {
		logs.Logger.Errorf("Error occur when start server %s", err.Error())
		os.Exit(1)
	}
}
