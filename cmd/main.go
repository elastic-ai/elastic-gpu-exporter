package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"elasticgpu.io/elastic-gpu-exporter/pkg/exporter"
	"elasticgpu.io/elastic-gpu-exporter/pkg/util"
	"net/http"
	"strings"
	"time"
)

const Resources = "nvidia.com/gpu,tke.cloud.tencent.com/qgpu-core,tke.cloud.tencent.com/qgpu-memory,elastic-gpu/gpu-percent"

var (
	addr    = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	node      string
	resources string
	interval  int
)

func init(){
	flag.StringVar(&node, "node", "", "node name")
	flag.StringVar(&resources, "labels", Resources, "gpu resources name")
	flag.IntVar(&interval, "interval", 30, "monitor interval (second)")
	flag.Parse()
}

func main() {
	e := exporter.NewExporter(node, strings.Split(resources, ","), time.Duration(interval) * time.Second)
	go e.Run(util.NeverStop)

	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			DisableCompression: true,
		},
	))
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}