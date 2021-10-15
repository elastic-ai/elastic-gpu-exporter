package exporter

import (
	"fmt"
	"k8s.io/klog"
	"nano-gpu-exporter/pkg/kubepods"
	"nano-gpu-exporter/pkg/metrics"
	"nano-gpu-exporter/pkg/nvidia"
	tree "nano-gpu-exporter/pkg/ptree"
	"nano-gpu-exporter/pkg/util"
	"strconv"
	"time"
	//"github.com/alex337/go-nvml"
	"tkestack.io/nvml"
	v1 "k8s.io/api/core/v1"
)

const (
	HundredCore = 100
	GiBToMiB    = 1024
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
	collector := metrics.NewCollector()
	collector.Register()
	ptree := tree.NewPTree(interval)
	podCache := NewCache()
	return &Exporter{
		node:      node,
		gpuLabels: gpuLabels,
		interval:  interval,
		podCache:  podCache,
		ptree:     ptree,
		collector: collector,

		watcher: kubepods.NewWatcher(&kubepods.Handler{
			AddFunc: func(pod *v1.Pod) {
				podCache.AddPod(string(pod.UID), pod)
				ptree.InterestPod(string(pod.UID), util.QoS(pod))
			},
			DelFunc: func(pod *v1.Pod) {
				podCache.DelPod(string(pod.UID))
				klog.Info("pod",pod.Name)
				ptree.ForgetPod(string(pod.UID))
			},
		}, gpuLabels, node),
	}
}

func (e *Exporter) once() {
	nvml.Init()
	defer nvml.Shutdown()
	cardCount, err := nvml.DeviceGetCount()
	klog.Info("Exporter run")
	if err != nil{
		klog.Error("Cannot get DeviceGetCount by nvml")
		klog.Info(err)
	}

	cardUsages := make([]tree.CardUsage, cardCount)
	processUsages := make([]map[int]*tree.ProcessUsage, cardCount)

	for i := 0; i < int(cardCount); i++ {
		processUsages[i], err = e.device.GetDeviceUsage(i)
		if err != nil{
			klog.Errorf("Cannot get processusage in GPU %d", i)
		}
	}
	klog.Info("e.podCache:",e.podCache)
	klog.Info("e.ptree.Snapshot():",e.ptree.Snapshot())

	var totalMem uint64
	for i := 0; i < int(cardCount); i++ {
		dev, err := nvml.DeviceGetHandleByIndex(uint(i))
		if err != nil{
			klog.Error("DeviceGetHandleByIndex", err)
		}
		_, _, memTotal, err := dev.DeviceGetMemoryInfo()
		totalMem += memTotal >> 20
	}

	node := e.ptree.Snapshot()
	for _, pod := range node.Pods{
		var podCore, podMem, podCoreRequest, podMemRequest float64

		p, _ := e.podCache.GetPod(pod.UID)
		klog.Info("p.Spec.Containers-------:",p.Spec.Containers)
		for _,cont := range p.Spec.Containers{
			klog.Info("cont.id",cont.Name)
		}
        klog.Info("p.Status.ContainerStatuses-------:",p.Status.ContainerStatuses)
		containerMap := make(map[string]string)
		for _, cont := range p.Status.ContainerStatuses{
			klog.Info("cont.ContainerID",cont.ContainerID)
			klog.Info("cont.Name",cont.Name)

			containerMap[cont.ContainerID] = cont.Name
		}
		ns := p.Namespace
		klog.Info("containerMap:",containerMap)
		for _, container := range pod.Containers{

			contName, exist := containerMap[fmt.Sprintf(util.ContainerID,container.ID)]
			klog.Info("container.parent",container.Parent)
			if !exist {
				continue
			}
			var contCore, contMem float64
			for _, proc := range container.Processes{
				for i := 0; i < int(cardCount); i++ {
					procUsage, exist := processUsages[i][proc.Pid]
					if exist {
						contMem  += procUsage.GPUMem
						contCore += procUsage.GPUCore
						//klog.Info("contCore----------------:", contCore)
						cardUsages[i].Mem  += procUsage.GPUMem
						cardUsages[i].Core += procUsage.GPUCore
					}
				}
			}
			//contMem /= float64(1024)
			podCore += contCore
			podMem += contMem
			var memRequest, coreRequest float64
			for _,cont := range p.Spec.Containers{
				if contName == cont.Name {
					memRequest = float64(util.GetGPUMemoryFromContainer(&cont))
					coreRequest = float64(util.GetGPUCoreFromContainer(&cont))
				}
			}
			podCoreRequest += coreRequest
			podMemRequest += memRequest

			var contCoreUtil float64
			var contMemUtil float64
			if contCore != 0 && coreRequest != 0{
				contCoreUtil = contCore / coreRequest
			}
			if contMem != 0 && memRequest != 0{
				contMemUtil = (contMem / GiBToMiB) / memRequest
			}
			klog.Info("contMem:",contMem)
			klog.Info("memRequest:",memRequest)
			klog.Info("contCore:",contCore)
			klog.Info("coreRequest:",coreRequest)

			e.collector.Container(e.node, ns, pod.UID, contName, contCore, contMem, util.Decimal(contCoreUtil * 100), util.Decimal(contMemUtil * 100))
		}
		var podMemUtil, podCoreUtil float64
		if podMemRequest != 0 && podMem != 0 {
			podMemUtil = (podMem / GiBToMiB) / podMemRequest
		}
		if podCoreRequest != 0 && podCore != 0 {
			podCoreUtil = podCore / podCoreRequest
		}

		e.collector.Pod(e.node, ns, pod.UID, podCore, podMem, util.Decimal(podCoreUtil * 100), util.Decimal(podMemUtil * 100), podMemRequest * GiBToMiB, util.Decimal(podCore / float64(cardCount * HundredCore) * 100), util.Decimal(podMem / float64(totalMem) * 100))
	}
	for i := 0; i < int(cardCount); i++ {
		dev, err := nvml.DeviceGetHandleByIndex(uint(i))
		_, memUsed, memTotal, err := dev.DeviceGetMemoryInfo()
		//util1, _ := dev.DeviceGetAverageGPUUsage(time.Second)
		//klog.Info("util1----------",util1)


		utilization, err := dev.DeviceGetUtilizationRates()
		//klog.Info("util:",utilization.GPU)

		if err != nil {
			klog.Error("DeviceGetMemoryInfo", err)
		}

		if cardUsages[i].Mem >= 0 || cardUsages[i].Core >= 0 {
			e.collector.Card(e.node, strconv.Itoa(i), cardUsages[i].Core, float64(memUsed >> 20), util.Decimal(float64(utilization.GPU )), util.Decimal(float64(memUsed >> 20) / float64(memTotal >> 20) * 100))
		}
	}
}

func (e *Exporter) Run(stop <-chan struct{}) {
	go e.ptree.Run(stop)
	e.watcher.Run(stop)
	util.Loop(e.once, e.interval, stop)
}
