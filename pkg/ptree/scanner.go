package ptree

import (
	"bufio"
	"io/ioutil"
	"k8s.io/klog"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
)
const(
	QOSGuaranteed = "guaranteed"
	QOSBurstable  = "burstable"
	QOSBestEffort = "bestEffort"
	CgroupBase    = "/sys/fs/cgroup/memory"
	PodPrefix     = "pod"
	CGROUP_PROCS  = "cgroup.procs"
	kubeRoot      = "kubepods"
)

var (
	validShortID = regexp.MustCompile("^[a-f0-9]{64}$")
)

func IsContainerID(id string) bool {
	return validShortID.MatchString(id)
}

type Scanner interface {
	Scan(UID, QOS string) (Pod, error)
}

type ScannerImpl struct{
	pod       Pod
	container Container
}

type CgroupName []string

func (scan *ScannerImpl) Scan(UID, QOS string) (Pod, error) {
	scan.pod.Containers = make(map[string]*Container)
	scan.container.Processes = make(map[int]*Process)
	pod, err := scan.getContainers(NewPod(QOS, UID))
	if err != nil {
		klog.Errorf("Cannot scan pod: pod%s, %v", UID, err)
		return Pod{}, err
	}
	return *pod, nil
}

func (scan *ScannerImpl) getContainers(p *Pod) (*Pod, error) {
	podPath := scan.getPodPath(p.UID, p.QOS)
	basePodPath := filepath.Clean(filepath.Join(CgroupBase, podPath))
	containers, err := scan.readContainerFile(basePodPath, p)
	if err !=nil {
		klog.Errorf("Cannot read the containers in the pod: pod%s, %v", p.UID, err)
		return nil, err
	}
	return &Pod{
		UID:        p.UID,
		QOS:        p.QOS,
		Containers: containers,
	},nil
}

//getPodPath is to get the path of the pod ,such as:kubepods/besteffort/pod17eb80b0-6085-4d12-8e79-553e799d2f0b
func (scan *ScannerImpl) getPodPath(UID string, QOS string) (podPath string) {
	var parentPath CgroupName
	switch QOS {
	case QOSGuaranteed:
		parentPath = append(parentPath,kubeRoot)
	case QOSBurstable:
		parentPath = append(parentPath, kubeRoot, QOSBurstable)
	case QOSBestEffort:
		parentPath = append(parentPath, kubeRoot, QOSBestEffort)
	}
	podContainer := PodPrefix + UID
	parentPath = append(parentPath,podContainer)
	podPath = scan.transformToPath(parentPath)
	return podPath
}

func (scan *ScannerImpl) transformToPath(cgroupName CgroupName) string {
	return "/" + path.Join(cgroupName...)
}

func (scan *ScannerImpl)readContainerFile(podPath string, pod *Pod) (map[string]*Container, error) {
	fileList, err := ioutil.ReadDir(podPath)
	if err != nil {
		klog.Errorf("Can't read %s, %v", podPath, err)
		return nil, err
	}
	for _,file :=range fileList {
		containerId := file.Name()
		if IsContainerID(containerId) {
			scan.pod.AddContainer(containerId)
			scan.pod.Containers[containerId] =  &Container{
				ID: containerId,
				Parent: pod,
			}
			procPath := filepath.Join(podPath, containerId, CGROUP_PROCS)
			process, err := scan.readPidFile(procPath, scan.pod.Containers[containerId])
			if err != nil {
				klog.Errorf("Cannot read the pid in the container: %s, %v", containerId, err)
				return nil, err
			}
			scan.pod.Containers[containerId].Processes = process
			}
	}
	return scan.pod.Containers, nil
}

func (scan *ScannerImpl)readPidFile(procPath string, container *Container) (map[int]*Process, error) {
	file, err := os.Open(procPath)
	if err != nil {
		klog.Errorf("Cannot read %s, %v", procPath, err)
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if pid, err := strconv.Atoi(line); err == nil {
			scan.container.AddProcess(pid)
			scan.container.Processes[pid] =  &Process{
				Pid: pid,
				Parent: container,
			}
		}
	}
	klog.V(4).Infof("Read from %s, pids", procPath, scan.container.Processes)
	return scan.container.Processes, nil
}