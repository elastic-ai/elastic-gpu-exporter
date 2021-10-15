package util

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"strconv"
	"time"
)

var NeverStop = make(chan struct{})

// TODO: add recover
func Loop(f func(), duration time.Duration, stop <-chan struct{}) {
	for range time.Tick(duration) {
		select {
		case <- stop:
			return
		default:
			f()
		}
	}
}

func GetGPUCoreFromContainer(container *v1.Container) int {
	val, ok := container.Resources.Limits[ResourceGPUCore]
	if !ok {
		return 0
	}
	return int(val.Value())
}

func GetGPUMemoryFromContainer(container *v1.Container) int {
	val, ok := container.Resources.Limits[ResourceGPUMemory]
	if !ok {
		return 0
	}
	return int(val.Value())
}

func Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

