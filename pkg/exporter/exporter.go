package exporter

import (
	"nano-gpu-exporter/pkg/kubepods"
	"nano-gpu-exporter/pkg/metrics"
	tree "nano-gpu-exporter/pkg/ptree"
	"nano-gpu-exporter/pkg/util"
	"time"

	v1 "k8s.io/api/core/v1"
)

type Exporter struct {
	node      string
	gpuLabels []string
	interval  time.Duration

	podCache  Cache
	ptree     tree.PTree
	collector *metrics.Collector

	watcher kubepods.Watcher
}

func NewExporter(node string, gpuLabels []string, interval time.Duration) *Exporter {
	ptree := tree.NewPTree(interval)
	podCache := NewCache()
	return &Exporter{
		node:      node,
		gpuLabels: gpuLabels,
		interval:  interval,

		podCache:  podCache,
		ptree:     ptree,
		collector: metrics.NewCollector(),

		watcher: kubepods.NewWatcher(&kubepods.Handler{
			AddFunc: func(pod *v1.Pod) {
				podCache.AddPod(string(pod.UID), pod)
				ptree.InterestPod(string(pod.UID), util.QoS(pod))
			},
			DelFunc: func(pod *v1.Pod) {
				podCache.DelPod(string(pod.UID))
				ptree.ForgetPod(string(pod.UID))
			},
		}, gpuLabels, node),
	}
}

func (e *Exporter) once() {

}

func (e *Exporter) Run(stop <-chan struct{}) {
	e.watcher.Run(stop)
	util.Loop(e.once, e.interval, stop)
}
