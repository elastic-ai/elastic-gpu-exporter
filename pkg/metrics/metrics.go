package metrics

import "github.com/prometheus/client_golang/prometheus"

type Collector struct {
	GPUCore           *prometheus.GaugeVec
	GPUCoreUtil       *prometheus.GaugeVec
	GPUMem            *prometheus.GaugeVec
	GPUMemUtil        *prometheus.GaugeVec
	PodCore           *prometheus.GaugeVec
	PodCoreUtil       *prometheus.GaugeVec
	PodCoreOccupyNode *prometheus.GaugeVec
	PodMem            *prometheus.GaugeVec
	PodMemUtil        *prometheus.GaugeVec
	PodMemOccupyNode  *prometheus.GaugeVec
	PodMemRequest     *prometheus.GaugeVec
	ContainerCore     *prometheus.GaugeVec
	ContainerCoreUtil *prometheus.GaugeVec
	ContainerMem      *prometheus.GaugeVec
	ContainerMemUtil  *prometheus.GaugeVec
}

func NewCollector() *Collector {
	return &Collector{
		GPUCore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gpu_core_usage",
				Help: "Usage of gpu core per card",
			},
			[]string{"node","card"},
		),
		GPUCoreUtil: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gpu_core_utilization",
				Help: "Utilization of gpu core per card",
			},
			[]string{"node","card"},
		),
		GPUMem: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gpu_mem_usage",
				Help: "Usage of gpu memory per card",
			},
			[]string{"node","card"},
		),
		GPUMemUtil: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gpu_mem_utilization",
				Help: "Utilization of gpu memory per card",
			},
			[]string{"node","card"},
		),
		PodCore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pod_core_usage",
				Help: "Usage of gpu core per pod",
			},
			[]string{"node","namespace", "pod"},
		),
		PodCoreUtil: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pod_core_utilization",
				Help: "Utilization of gpu core",
			},
			[]string{"node","namespace", "pod"},
		),
		PodCoreOccupyNode: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pod_core_occupy_node",
				Help: "Utilization of pod core occupied the node",
			},
			[]string{"node","namespace", "pod"},
		),
		PodMem: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pod_mem_usage",
				Help: "Usage of gpu memory per pod",
			},
			[]string{"node","namespace", "pod"},
		),
		PodMemUtil: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pod_mem_utilization",
				Help: "Utilization of pod memory",
			},
			[]string{"node","namespace", "pod"},
		),
		PodMemOccupyNode: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pod_mem_occupy_node",
				Help: "Utilization of pod memory occupied the node",
			},
			[]string{"node","namespace", "pod"},
		),
		PodMemRequest: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pod_mem_request",
				Help: "Request of pod memory",
			},
			[]string{"node","namespace", "pod"},
		),
		ContainerCore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "container_core_usage",
				Help: "Usage of gpu computing per container",
			},
			[]string{"node","namespace", "pod", "container"},
		),
		ContainerCoreUtil: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "container_core_utilization",
				Help: "Utilization of container core",
			},
			[]string{"node","namespace", "pod", "container"},
		),
		ContainerMem: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "container_mem_usage",
				Help: "Usage of gpu memory per container",
			},
			[]string{"node","namespace", "pod", "container"},
		),
		ContainerMemUtil: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "container_mem_utilization",
				Help: "Utilization of container memory",
			},
			[]string{"node","namespace", "pod", "container"},
		),
	}
}

func (c *Collector) Register() {
	prometheus.MustRegister(c.PodMem)
	prometheus.MustRegister(c.PodMemUtil)
	prometheus.MustRegister(c.PodMemOccupyNode)
	prometheus.MustRegister(c.PodMemRequest)
	prometheus.MustRegister(c.PodCore)
	prometheus.MustRegister(c.PodCoreUtil)
	prometheus.MustRegister(c.PodCoreOccupyNode)
	prometheus.MustRegister(c.GPUMem)
	prometheus.MustRegister(c.GPUMemUtil)
	prometheus.MustRegister(c.GPUCore)
	prometheus.MustRegister(c.GPUCoreUtil)
	prometheus.MustRegister(c.ContainerCore)
	prometheus.MustRegister(c.ContainerCoreUtil)
	prometheus.MustRegister(c.ContainerMem)
	prometheus.MustRegister(c.ContainerMemUtil)


}

func (c *Collector) Card(node, id string, core, mem, coreUtil, memUtil float64) {
	c.GPUCore.WithLabelValues(node, id).Set(core)
	c.GPUMem.WithLabelValues(node,id).Set(mem)
	c.GPUCoreUtil.WithLabelValues(node,id).Set(coreUtil)
	c.GPUMemUtil.WithLabelValues(node,id).Set(memUtil)



}

func (c *Collector) Pod(node, namespace, name string, core, mem, coreUtil, memUtil, memRequest, coreOccupy, memOccupy float64) {
	c.PodCore.WithLabelValues(node, namespace, name).Set(core)
	c.PodMem.WithLabelValues(node, namespace, name).Set(mem)
	c.PodMemRequest.WithLabelValues(node, namespace, name).Set(memRequest)
	c.PodMemUtil.WithLabelValues(node, namespace, name).Set(memUtil)
	c.PodCoreUtil.WithLabelValues(node, namespace, name).Set(coreUtil)
	c.PodMemOccupyNode.WithLabelValues(node, namespace, name).Set(memOccupy)
	c.PodCoreOccupyNode.WithLabelValues(node, namespace, name).Set(coreOccupy)


}

func (c *Collector) Container(node, namespace, pod, container string, core, mem, coreUtil, memUtil float64) {
	c.ContainerCore.WithLabelValues(node, namespace, pod, container).Set(core)
	c.ContainerMem.WithLabelValues(node, namespace, pod, container).Set(mem)
	c.ContainerCoreUtil.WithLabelValues(node, namespace, pod, container).Set(coreUtil)
	c.ContainerMemUtil.WithLabelValues(node, namespace, pod, container).Set(memUtil)

}


