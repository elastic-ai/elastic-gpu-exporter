package metrics

import "github.com/prometheus/client_golang/prometheus"

type Collector struct {
	GPUCore       *prometheus.GaugeVec
	GPUMemo       *prometheus.GaugeVec
	PodCore       *prometheus.GaugeVec
	PodMemo       *prometheus.GaugeVec
	ContainerCore *prometheus.GaugeVec
	ContainerMemo *prometheus.GaugeVec
}

func NewCollector() *Collector {
	return &Collector{
		GPUCore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gpu_core_usage",
				Help: "Usage of gpu core per card",
			},
			[]string{"card"},
		),
		GPUMemo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gpu_memo_usage",
				Help: "Usage of gpu memory per card",
			},
			[]string{"card"},
		),
		PodCore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pod_core_usage",
				Help: "Usage of gpu core per pod",
			},
			[]string{"namespace", "pod"},
		),
		PodMemo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "pod_memo_usage",
				Help: "Usage of gpu memory per pod",
			},
			[]string{"namespace", "pod"},
		),
		ContainerCore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "container_core_usage",
				Help: "Usage of gpu computing per container",
			},
			[]string{"namespace", "pod", "container"},
		),
		ContainerMemo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "container_memo_usage",
				Help: "Usage of gpu memory per container",
			},
			[]string{"namespace", "pod", "container"},
		),
	}
}

func (c *Collector) Card(id string, core, memo float64) {
	c.GPUCore.WithLabelValues(id).Set(core)
	c.GPUMemo.WithLabelValues(id).Set(memo)
}

func (c *Collector) Pod(namespace, name string, core, memo float64) {
	c.PodCore.WithLabelValues(namespace, name).Set(core)
	c.PodMemo.WithLabelValues(namespace, name).Set(memo)
}

func (c *Collector) Container(namespace, pod, container string, core, memo float64) {
	c.ContainerCore.WithLabelValues(namespace, pod, container).Set(core)
	c.ContainerMemo.WithLabelValues(namespace, pod, container).Set(memo)
}
