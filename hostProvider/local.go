package hostProvider

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/cpg1111/kubongo/image"
)

// Local is for development, it is not recommended to run the local platform in production

// LocalInstance is an instance struct for the "local" platform
type LocalInstance struct {
	Instance
	Process     string
	ProcessPort int
	Name        string
	IP          string
	Zone        string
}

// LocalHost controls Instances for the "local" platform
type LocalHost struct {
	HostProvider
	Instances []Instance
}

// NewLocal returns a new LocalHost struct
func NewLocal() *LocalHost {
	return &LocalHost{}
}

// GetServers returns all local servers, i.e. registered process
func (l LocalHost) GetServers(namespace string) ([]Instance, error) {
	return l.Instances, nil
}

// GetServer returns a specific server/process
func (l LocalHost) GetServer(project, zone, name string) (Instance, error) {
	for i := range l.Instances {
		if l.Instances[i].(LocalInstance).Name == name {
			return l.Instances[i], nil
		}
	}
	return nil, fmt.Errorf("Could not find %s in localhost", name)
}

// CreateServer creates a new server
func (l LocalHost) CreateServer(namespace, zone, name, machineType, sourceImage, source string) (Instance, error) {
	var process string
	if source != "" {
		process = source
	} else if sourceImage != "" {
		process = "docker"
	} else {
		process = "mongo"
	}
	port, pErr := strconv.Atoi(machineType)
	newInst := &LocalInstance{
		Process:     process,
		ProcessPort: port,
		Name:        name,
		IP:          "127.0.0.1",
	}
	pointerL := &l
	pointerL.Instances = append(l.Instances, newInst)
	if strings.Contains(newInst.Process, "mongo") {
		imageManager := image.NewImageManager("local")
		imageManager.InstallMongo()
	}
	return newInst, pErr
}

func killProc(proc string) error {
	pid, pErr := exec.Command("pgrep", "-f", proc).Output()
	if pErr != nil {
		return pErr
	}
	pidNum, pnErr := strconv.Atoi(string(pid))
	if pnErr != nil {
		return pnErr
	}
	process, prErr := os.FindProcess(pidNum)
	if prErr != nil {
		return prErr
	}
	return process.Kill()
}

// DeleteServer will delete a registered server i.e. kill a process
func (l LocalHost) DeleteServer(namespace, zone, name string) error {
	for i := range l.Instances {
		if l.Instances[i].(LocalInstance).Name == name {
			proc := l.Instances[i].(LocalInstance).Process
			killProc(proc)
			if i+1 < len(l.Instances) {
				l.Instances[i] = l.Instances[i+1]
			} else {
				l.Instances[i] = nil
			}
		}
	}
	return errors.New("Could not find instance in local instances")
}
