package ptree

import (
	"bufio"
	"io/ioutil"
	"k8s.io/klog"
	"os"
	"path"
	"path/filepath"
	"strconv"
)
const(
	QOSGuaranteed = "guaranteed"
	QOSBurstable  = "burstable"
	QOSBestEffort = "bestEffort"
	CgroupBase    = "/sys/fs/cgroup/memory"
	PodPrefix     = "pod"
	CGROUP_PROCS = "cgroup.procs"
)

var (
	kubeRoot = []string{"kubepods"}
)

type Scanner interface {
	Scan(UID, QOS string) (*Pod, error)
}

type ScannerImpl struct{
	pod       *Pod
	container *Container
	process   *Process
	node      *Node
}

type CgroupName []string

func (scan *ScannerImpl) Scan(UID, QOS string) (*Pod, error){
	return scan.getContainers(NewPod(UID, QOS)), nil

}

func (scan *ScannerImpl) getContainers(p *Pod) *Pod {
	podPath := scan.getPodPath(p.UID, p.QOS)
	var containers map[string]*Container
	containers = make(map[string]*Container)

	baseDir := filepath.Clean(filepath.Join(CgroupBase, podPath))
	filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		containers = scan.readContainerFile(path, p)
		return nil
	})
	return &Pod{
		UID:        p.UID,
		QOS:        p.QOS,
		Containers: containers,
	}
}

//getPodPath is to get the path of the pod ,such as:kubepods/besteffort/pod17eb80b0-6085-4d12-8e79-553e799d2f0b
func (scan *ScannerImpl) getPodPath(UID string, QOS string) (podPath string) {
	var parentPath CgroupName
	switch QOS {
	case QOSGuaranteed:
		parentPath = kubeRoot
	case QOSBurstable:
		parentPath = append(kubeRoot, QOSBurstable)
	case QOSBestEffort:
		parentPath = append(kubeRoot, QOSBestEffort)
	}
	podContainer := PodPrefix + UID
	cgroupName := append(parentPath, podContainer)
	podPath = scan.ToCgroupPath(cgroupName)
	return podPath
}


func (scan *ScannerImpl) ToCgroupPath(cgroupName CgroupName) string {
	return "/" + path.Join(cgroupName...)
}

func (scan *ScannerImpl)readContainerFile(podPath string, pod *Pod) map[string]*Container {
	fileList, err := ioutil.ReadDir(podPath)
	if err != nil {
		klog.Errorf("can't read %s, %v", podPath, err)
		return nil
	}
	for _,file :=range fileList {
		if len(file.Name()) == 64 {
			containerId := file.Name()
			scan.pod.AddContainer(containerId)
			scan.pod.Containers[containerId] = &Container{
				ID:      file.Name(),
				Parent:  pod,
			}
			procPath := filepath.Join(podPath, containerId, CGROUP_PROCS)
			if err, processes :=scan.readPidFile(procPath, *scan.pod.Containers[containerId]); err != nil{
				scan.pod.Containers[containerId].Processes = processes
			}
		}
	}
	return scan.pod.Containers

}

func (scan *ScannerImpl)readPidFile(procPath string, container Container) (error, map[int]*Process) {
	file, err := os.Open(procPath)
	if err != nil {
		klog.Errorf("can't read %s, %v", procPath, err)
		return nil,nil
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if pid, err := strconv.Atoi(line); err == nil {
			scan.container.AddProcess(pid)
			scan.container.Processes[pid] = &Process{
				Pid:    pid,
				Parent: &container,
			}
		}
	}
	klog.V(4).Infof("Read from %s, pids", procPath, scan.container.Processes)
	return nil, scan.container.Processes
}