module nano-gpu-exporter

go 1.16

//replace tkestack.io/nvml => github.com/alex337/go-nvml v1
replace tkestack.io/nvml => github.com/tkestack/go-nvml v0.0.0-20191217064248-7363e630a33e

require (
	github.com/NVIDIA/gpu-monitoring-tools v0.0.0-20210817155834-f476d8a022cf // indirect
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/common v0.4.1
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v0.17.4
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.17.4
	//github.com/alex337/go-nvml v1.0.0
	tkestack.io/nvml v0.0.0-00010101000000-000000000000

)
