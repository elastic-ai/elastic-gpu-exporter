package util

import (
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kubectl/pkg/util/qos"
)

func QoS(pod *v1.Pod) string {
	podQoS := pod.Status.QOSClass
	if podQoS == "" {
		podQoS = qos.GetPodQOS(pod)
	}
	return strings.ToLower(string(podQoS))
}

func PodHasResource(pod *v1.Pod, set map[string]struct{}) bool {
	for _, container := range pod.Spec.Containers {
		for name, _ := range container.Resources.Limits {
			if _, ok := set[name.String()]; ok {
				return true
			}
		}
	}
	return false
}
