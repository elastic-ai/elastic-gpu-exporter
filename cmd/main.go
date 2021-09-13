package main

import (
	"flag"
	"nano-gpu-exporter/pkg/exporter"
	"nano-gpu-exporter/pkg/util"
	"strings"
	"time"
)

const Resources = "nvidia.com/gpu,tke.cloud.tencent.com/qgpu-core,tke.cloud.tencent.com/qgpu-core"

var (
	node      string
	resources string
	interval  int
)

func Init() {
	flag.StringVar(&node, "node", "", "node name")
	flag.StringVar(&resources, "labels", Resources, "gpu resources name")
	flag.IntVar(&interval, "interval", 30, "monitor interval (second)")
	flag.Parse()
}

func main() {
	Init()
	e := exporter.NewExporter(node, strings.Split(resources, ","), time.Duration(interval) * time.Second)
	e.Run(util.NeverStop)
}