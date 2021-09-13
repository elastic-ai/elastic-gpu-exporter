package exporter

import (
	"k8s.io/klog"
	"nano-gpu-exporter/pkg/kubepods"
	"nano-gpu-exporter/pkg/metrics"
	"nano-gpu-exporter/pkg/nvidia"
	tree "nano-gpu-exporter/pkg/ptree"
	"nano-gpu-exporter/pkg/util"
	"time"

	v1 "k8s.io/api/core/v1"
)

const (
	CardNum0 = "0"
	CardNum1 = "1"
)

type Exporter struct {
	node       string
	gpuLabels  []string
	interval   time.Duration

	podCache   Cache
	ptree      tree.PTree
	collector  *metrics.Collector
    device     *nvidia.DeviceImpl
	watcher    kubepods.Watcher
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
	processUsage0 := make(map[int]*tree.ProcessUsage)
	processUsage1 := make(map[int]*tree.ProcessUsage)

	processUsage0, err := e.device.GetDeviceUsage(0)
	if err != nil{
		klog.Error("Cannot get processusage in GPU 0")
	}
	processUsage1, err = e.device.GetDeviceUsage(1)
	if err != nil{
		klog.Error("Cannot get processusage in GPU 1")
	}
	var cardMemUsage0, cardCoreUsage0, cardMemUsage1, cardCoreUsage1 float64

	node := e.ptree.Snapshot()

	for _, pod := range node.Pods{
		var podCore, podMem float64

		p,_ := e.podCache.GetPod(pod.UID)
		klog.Info(p.Name)
		ns := p.Namespace
		for _, container := range pod.Containers{
			var contCore, contMem float64
			for _, proc := range container.Processes{
				procUsage0, exist := processUsage0[proc.Pid]
				if exist {
					contMem  += procUsage0.GPUMemo
					contCore += procUsage0.GPUCore
					cardMemUsage0  += procUsage0.GPUMemo
					cardCoreUsage0 += procUsage0.GPUCore
				}

				procUsage1, exist := processUsage1[proc.Pid]
				if exist {
					contMem  += procUsage1.GPUMemo
					contCore += procUsage1.GPUCore
					cardMemUsage1  += procUsage1.GPUMemo
					cardCoreUsage1 += procUsage1.GPUCore
				}
			}

			podCore += contCore
			podMem += contMem
			var contName string
			for _, cont := range p.Status.ContainerStatuses {
				if cont.ContainerID == container.ID {
					contName = cont.Name
				}
			}
			e.collector.Container(ns, pod.UID, contName, contCore, contMem)
		}
		e.collector.Pod(ns, pod.UID, podCore, podMem)
	}

	if cardCoreUsage0 >= 0 || cardMemUsage0 >= 0 {
		e.collector.Card(CardNum0, cardCoreUsage0, cardMemUsage0)
	}

	if cardCoreUsage1 >= 0 || cardMemUsage1 >= 0 {
		e.collector.Card(CardNum1, cardCoreUsage1, cardMemUsage1)
	}


}

func (e *Exporter) Run(stop <-chan struct{}) {
	e.watcher.Run(stop)
	util.Loop(e.once, e.interval, stop)
}
