package hostProvider

import(
  "errors"
  "fmt"
  "os"
  "os/exec"
  "strconv"
  "strings"

  "github.com/cpg1111/kubongo/image"
)

type LocalInstance struct {
  Instance
  Process     string
  ProcessPort int
  Name        string
  IP          string
}

type LocalHost struct {
  HostProvider
  Instances []Instance
}

func NewLocal() *LocalHost {
  return &LocalHost{}
}

func(l LocalHost) GetServers(namespace string) ([]Instance, error) {
  return l.Instances, nil
}

func(l LocalHost) GetServer(project, zone, name string) (Instance, error) {
  for i := range l.Instances {
    if l.Instances[i].(LocalInstance).Name == name {
      return l.Instances[i], nil
    }
  }
  return nil, errors.New(fmt.Sprintf("Could not find %s in localhost", name))
}

func(l LocalHost) CreateServer(namespace, zone, name, machineType, sourceImage, source string) (Instance, error){
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
    Process: process,
    ProcessPort: port,
    Name: name,
    IP: "127.0.0.1",
  }
  pointerL := &l
  pointerL.Instances = append(l.Instances, newInst)
  if strings.Contains(newInst.Process, "mongo") {
    imageManager := image.NewImageManager("local")
    imageManager.InstallMongo()
  }
  return newInst, pErr
}

func killProc(proc string) error{
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

func(l LocalHost) DeleteServer(namespace, zone, name string) error {
  for i := range l.Instances {
    if l.Instances[i].(LocalInstance).Name == name {
      proc := l.Instances[i].(LocalInstance).Process
      killProc(proc)
      if i + 1 < len(l.Instances) {
        l.Instances[i] = l.Instances[i + 1]
      } else {
        l.Instances[i] = nil
      }
    }
  }
  return errors.New("Could not find instance in local instances")
}
