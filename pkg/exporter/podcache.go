package exporter

import (
	"sync"

	v1 "k8s.io/api/core/v1"
)

type Cache interface {
	AddPod(UID string, pod *v1.Pod)
	DelPod(UID string)
	GetPod(UID string) (*v1.Pod, bool)
}

type PodCache struct {
	cache map[string]*v1.Pod
	mu    sync.Mutex
}

func (p *PodCache) AddPod(UID string, pod *v1.Pod) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cache[UID] = pod
}

func (p *PodCache) DelPod(UID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.cache, UID)
}

func (p *PodCache) GetPod(UID string) (*v1.Pod, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	pod, ok := p.cache[UID]
	return pod, ok
}

func NewCache() Cache {
	return &PodCache{
		cache: make(map[string]*v1.Pod),
		mu:    sync.Mutex{},
	}
}

