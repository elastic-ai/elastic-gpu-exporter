package nvidia

import (
	"k8s.io/klog"
	process "nano-gpu-exporter/pkg/ptree"
	"time"
	"tkestack.io/nvml"
)

type Device interface {
	GetDeviceUsage(cardNum int)  (map[int]*process.ProcessUsage,error)
}

type DeviceImpl struct {
}

func (device *DeviceImpl) GetDeviceUsage(cardNum int)  (map[int]*process.ProcessUsage,error ){
	nvml.Init()
	defer nvml.Shutdown()
	dev, _ := nvml.DeviceGetHandleByIndex(uint(cardNum))
	processOnDevices, err := dev.DeviceGetComputeRunningProcesses(1024)
	if err != nil {
		klog.Warningf("Can't get processes info from device %d, error %s", uint(cardNum), err)
		return nil, err
	}
	usageMap := make(map[int]*process.ProcessUsage)
	for _, info := range processOnDevices {
		_, exit := usageMap[int(info.Pid)]
		if !exit {
			usageMap[int(info.Pid)] = new(process.ProcessUsage)
		}
		usageMap[int(info.Pid)].GPUMemo = float64(info.UsedGPUMemory >> 20)
	}
	processUtilization, err := dev.DeviceGetProcessUtilization(1024, time.Second)
	if err != nil {
		klog.Warningf("Can't get processes utilization from device %d, error %s", uint(cardNum), err)
		return nil, err
	}
	for _, info := range processUtilization {
		_, exit := usageMap[int(info.Pid)]
		if !exit {
			usageMap[int(info.Pid)] = new(process.ProcessUsage)
		}
		usageMap[int(info.Pid)].GPUCore = float64(info.SmUtil)
	}
	return usageMap, nil
}

func (device *DeviceImpl) getPidUsage(pid int) (*process.ProcessUsage, error) {
	nvml.Init()
	defer nvml.Shutdown()
	num, err := nvml.DeviceGetCount()
	if err != nil {
		return nil, err
	}
	var usedMemory float64
	var usedCore float64

	for i := 0; i < int(num); i++ {
		dev, _ := nvml.DeviceGetHandleByIndex(uint(i))
		processOnDevices, err := dev.DeviceGetComputeRunningProcesses(1024)
		if err != nil {
			klog.Warningf("Can't get processes info from device %d, error %s", uint(i), err)
			return nil, err
		}
		for _, info := range processOnDevices {
			if int(info.Pid) == pid {
				usedMemory = float64(info.UsedGPUMemory >> 20)
			}
		}
		processUtilizations, err := dev.DeviceGetProcessUtilization(1024, time.Second)
		if err != nil {
			klog.Warningf("Can't get processes utilization from device %d, error %s", uint(i), err)
			return nil, err
		}
		for _, info := range processUtilizations {
			if int(info.Pid) == pid {
				usedCore = float64(info.SmUtil)
			}
		}
		return &process.ProcessUsage{
			GPUMemo:   usedMemory,
			GPUCore:   usedCore,
		}, nil
	}
	return nil, err
}