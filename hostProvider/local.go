package hostProvider

import(
  "errors"
  "fmt"
  "os"
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

func NewLocal() *LocalHostProvider {
  return &LocalHost{}
}

func(l LocalHostProvider) GetServers(namespace string) ([]Instance, error) {
  return l.Instances, nil
}

func(l LocalHostProvider) GetServer(project, zone, name string) (Instance, error) {
  for i := range l.Instances {
    if l.Instances[i].Name == name {
      return l.Instances[i]
    }
  }
  return nil, errors.New(fmt.Sprintf("Could not find %s in localhost", name))
}

func(l LocalHostProvider) CreateServer(namespace, zone, name, machineType, sourceImage, source string) (Instance, error){
  var process string
  if source != "" {
    process = source
  } else if sourceImage != "" {
    process = "docker"
  } else {
    process = "mongo"
  }
  newInst := &LocalInstance{
    Process: process,
    ProcessPort: machineType,
    Name: name,
    IP: "127.0.0.1",
  }
  pointerL := &l
  pointerL.Instances = append(l.Instances, newInst)
  if strings.Compare(l.Process, "mongo") == 0 {
    imageManager := image.NewImageManager("local")
    imageManager.installMongo()
  }
}

func(l LocalHostProvider) DeleteServer(namespace, zone, name string) error {
  for i := range l.Instances {
    if l.Instances[i].Name == name {
      proc := l.Instances[i].Process
      //TODO kill mongo proc
      if i + 1 < len(l.Instances) {
        l.Instances[i] = l.Instances[i + 1]
      } else {
        l.Instances[i] = nil
      }
    }
  }
}
