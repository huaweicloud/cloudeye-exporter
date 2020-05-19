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

package main

import (
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/huaweicloud/cloudeye-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

var (
	clientConfig = flag.String("config", "./clouds.yml", "Path to the cloud configuration file")
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

	log.Infof("Start to monitor services: %s", targets)
	exporter, err := collector.GetMonitoringCollector(*clientConfig, targets, *debug)
	registry.MustRegister(exporter)
	if err != nil {
		log.Errorf("Fail to start to morning services: %s, err: %s", targets, err)
		return
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	flag.Parse()
	config, err := collector.NewCloudConfigFromFile(*clientConfig)
	if err != nil {
		log.Fatal(err)
		return
	}

	http.HandleFunc(config.Global.MetricPath, handler)

	log.Infoln("Start server at ", config.Global.Port)
	if err := http.ListenAndServe(config.Global.Port, nil); err != nil {
		log.Error("Error occur when start server %v", err)
		os.Exit(1)
	}
}
