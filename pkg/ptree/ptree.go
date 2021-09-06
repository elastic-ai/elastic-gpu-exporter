package ptree

// PTree is a common interface to detect the tree such as:
// node -> pods -> containers -> processes

type PTree interface {
	Snapshot() (Node, error)
}

type Node struct {
	Pods       map[string]*Pod
	Containers map[string]*Container
	Processes  map[int]*Process
}

func NewNode() *Node {
	return &Node{
		Pods:       make(map[string]*Pod),
		Containers: make(map[string]*Container),
		Processes:  make(map[int]*Process),
	}
}

func (n *Node) GetProcessByPid(pid int) (p *Process, exist bool) {
	if process, ok := n.Processes[pid]; ok {
		return process, true
	}
	return nil, false
}

func (n *Node) AddPod(pod *Pod) {

}

func (n *Node) AddContainer(container *Container) {

}

func (n *Node) AddProcess(process *Process) {

}

type Pod struct {
	QOS        string
	UID        string
	Parent     *Node
	Containers map[string]*Container
}

type Container struct {
	ID        string
	Parent    *Pod
	Processes map[int]*Process
}

type Process struct {
	Pid    int
	Parent *Container
}
