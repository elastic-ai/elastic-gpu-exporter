package ptree

import (
	"fmt"
	"nano-gpu-exporter/pkg/util"
	"strings"
	"sync"
	"time"
	"k8s.io/klog"
)

// PTree is a common interface to detect the tree such as:
// node -> pods -> containers -> processes

type PTree interface {
	Run()
	InterestPod(UID string, QOS string)
	ForgetPod(UID string)
	Snapshot() Node
}

type PTreeImpl struct {
	interval        time.Duration
	mu              sync.Mutex
	interestingPods map[string]string
	nodeSnapshot    *Node
	scanner         Scanner
}

func (p *PTreeImpl) Run() {
	util.Loop(func() {
		if err := p.nextSnapshot(); err != nil {
			klog.Error(err.Error())
		}
	}, p.interval, util.NeverStop)
}

func (p *PTreeImpl) InterestPod(UID string, QOS string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.interestingPods[UID] = QOS
}

func (p *PTreeImpl) ForgetPod(UID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.interestingPods, UID)
}

func (p *PTreeImpl) Snapshot() *Node {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.nodeSnapshot
}

func (p *PTreeImpl) nextSnapshot() error {
	var (
		pods     = p.interesting()
		errors   = []string{}
		snapshot = NewNode()
	)

	for UID, QOS := range pods {
		if pod, err := p.scanner.Scan(UID, QOS); err != nil {
			errors = append(errors, err.Error())
		} else {
			snapshot.AddPod(&pod)
		}
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.nodeSnapshot = snapshot
	if len(errors) == 0 {
		return nil
	}
	return fmt.Errorf("%d errors: %s", len(errors), strings.Join(errors, "; "))
}

func (p *PTreeImpl) interesting() map[string]string {
	p.mu.Lock()
	defer p.mu.Unlock()
	pods := map[string]string{}
	for UID, QOS := range p.interestingPods {
		pods[UID] = QOS
	}
	return pods
}
