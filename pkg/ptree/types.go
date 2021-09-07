package ptree

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

func (n *Node) addPod(pod *Pod) {
	for _, c := range pod.Containers {
		n.addContainer(c)
	}
	n.Pods[pod.UID] = pod
}

func (n *Node) addContainer(container *Container) {
	for _, p := range container.Processes {
		n.addProcess(p)
	}
	n.Containers[container.ID] = container
}

func (n *Node) addProcess(process *Process) {
	n.Processes[process.Pid] = process
}

type Pod struct {
	QOS        string
	UID        string
	Parent     *Node
	Containers map[string]*Container
}

func (c Container) AddContainer(ID string) *Container {

}

type Container struct {
	ID        string
	Parent    *Pod
	Processes map[int]*Process
}

func (c *Container) AddProcess(pid int) {
	c.Processes[pid] = &Process{
		Pid:    pid,
		Parent: c,
	}
}

type Process struct {
	Pid    int
	Parent *Container
}