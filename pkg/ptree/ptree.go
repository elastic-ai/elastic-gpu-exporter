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
	Run(stop <-chan struct{})
	InterestPod(UID, QOS string)
	ForgetPod(UID string)
	Snapshot() *Node
	LastUpdate() time.Time
}

type PTreeImpl struct {
	interval        time.Duration
	mu              sync.Mutex
	interestingPods map[string]string
	nodeSnapshot    *Node
	lastUpdate      time.Time
	scanner         Scanner
}

func NewPTree(interval time.Duration) *PTreeImpl {
	return &PTreeImpl{
		interval:        interval,
		mu:              sync.Mutex{},
		interestingPods: make(map[string]string),
		nodeSnapshot:    NewNode(),
		lastUpdate:      time.Now(),
		scanner:         NewScanner(),
	}
}

func (p *PTreeImpl) Run(stop <-chan struct{}) {
	util.Loop(func() {
		if err := p.nextSnapshot(); err != nil {
			klog.Error(err.Error())
		}
	}, p.interval, stop)
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

func (p *PTreeImpl) LastUpdate() time.Time {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lastUpdate
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
			snapshot.addPod(&pod)
		}
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.nodeSnapshot = snapshot
	p.lastUpdate = time.Now()
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
