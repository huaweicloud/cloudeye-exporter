package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/huaweicloud/huaweicloud-exporter/collector"
	"github.com/prometheus/common/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)


var (
	clientConfig = flag.String("clientConfig", "./clouds.yml", "Path to the cloud configuration file")
	debug = flag.Bool("debug", false, "If debug the code.")
)


func main() {
	flag.Parse()
	config, err := collector.NewCloudConfigFromFile(*clientConfig)
	if err != nil {
		log.Fatal(err)
		return
	}

	collector.SetDefaultConfigValues(config)
	client, err := collector.InitClient(config)
	if err != nil {
		log.Fatal(err)
		return
	}

	reg := prometheus.NewPedanticRegistry()
	for _, service := range config.InfoMetrics {
		exporter, err := collector.GetMonitoringCollector(client, config.Global.Prefix, service, *debug)
		if err != nil {
			log.Errorf("Fail to start to morning service: %s, err: %s", service, err)
			continue
		}
		reg.MustRegister(exporter)
	}

	gatherers := prometheus.Gatherers{reg}

	h := promhttp.HandlerFor(gatherers,
		promhttp.HandlerOpts{
			ErrorLog:      log.NewErrorLogger(),
			ErrorHandling: promhttp.ContinueOnError,
		})
	http.HandleFunc(config.Global.MetricPath, func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})

	log.Infoln("Start server at ", config.Global.Port)
	if err := http.ListenAndServe(config.Global.Port, nil); err != nil {
		log.Error("Error occur when start server %v", err)
		os.Exit(1)
	}
}
