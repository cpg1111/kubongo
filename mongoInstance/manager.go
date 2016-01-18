package mongoInstance

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cpg1111/kubongo/hostProvider"
	"github.com/cpg1111/kubongo/metadata"
)

type Manager struct {
	Project     string
	Platform    string
	platformCtl hostProvider.HostProvider
	data        map[string]hostProvider.Instance
}

func addToInstances(instances *metadata.Instances, newServer hostProvider.Instance) {
	newInstances := metadata.AddInstance(*instances, newServer)
	instances = &newInstances
}

func (m *Manager) Create(newInstanceTmpl *InstanceTemplate, instances *metadata.Instances) ([]byte, error) {
	newServer, serverErr := m.platformCtl.CreateServer(
		m.Platform,
		newInstanceTmpl.Zone,
		newInstanceTmpl.Name,
		newInstanceTmpl.MachineType,
		newInstanceTmpl.SourceImage,
		newInstanceTmpl.Source,
	)
	if serverErr != nil {
		return nil, serverErr
	}
	addToInstances(instances, newServer)
	newServerJSON, jErr := json.Marshal(&newServer)
	return newServerJSON, jErr
}

func (m *Manager) Register(zone, name string, instances *metadata.Instances) ([]byte, error) {
	var (
		newServer hostProvider.Instance
	  serverErr error
	)
	if string.Compare(zone, "local") == 0 {
		newServer, serverErr = m.platformCtl.CreateServer(m.Platform, zone, name)
	} else{
		newServer, serverErr = m.platformCtl.GetServer(m.Platform, zone, name)
	}
	if serverErr != nil {
		return nil, serverErr
	}
	addToInstances(instances, newServer)
	newServerJSON, jErr := json.Marshal(&newServer)
	return newServerJSON, jErr
}

func (m *Manager) Remove(zone, name string) error {
	dErr := m.platformCtl.DeleteServer(m.Platform, zone, name)
	return dErr
}

func masterTmpl() *InstanceTemplate {
	zone := os.Getenv("DEFAULT_ZONE")
	if zone == "" {
		zone = "us-central1-f"
	}
	machineType := os.Getenv("MASTER_MONGO_TYPE")
	if machineType == "" {
		machineType = "n1-standard-4"
	}
	sourceImage := os.Getenv("MONGO_INSTANCE_OS")
	if sourceImage == "" {
		sourceImage = "ubuntu-14-04"
	}
	return &InstanceTemplate{
		Kind:        "Create",
		Name:        "master",
		MachineType: machineType,
		SourceImage: sourceImage,
		Source:      "",
	}
}

func (m *Manager) newMaster(rStatus, nStatus chan error, success chan []byte, instances *metadata.Instances) {
	log.Println(m.data["master"])
	master := m.data["master"].(hostProvider.GcloudInstance)
	rStatus <- m.Remove(master.Zone, master.Name)
	newInstance := masterTmpl()
	newBytes, nErr := m.Create(newInstance, instances)
	if nErr != nil {
		nStatus <- nErr
	} else {
		success <- newBytes
	}
}

func (m *Manager) Monitor(masterIP *string, instances *metadata.Instances) {
	monitor := newMonitor(masterIP)
	isHealthy := true
	healthChannel := make(chan bool)
	for isHealthy {
		go monitor.HealthCheck(healthChannel)
		isHealthy = <-healthChannel
		time.Sleep(time.Second * 3)
	}
	removeStatus := make(chan error)
	newMasterStatus := make(chan error)
	successStatus := make(chan []byte)
	go m.newMaster(removeStatus, newMasterStatus, successStatus, instances)
	for {
		select {
		case m1 := <-removeStatus:
			log.Fatal(m1)
			break
		case m2 := <-newMasterStatus:
			log.Fatal(m2)
			break
		case m3 := <-successStatus:
			log.Println("Created", string(m3))
			break
		}
	}
}

func NewManager(proj, pf string, pfctl *hostProvider.HostProvider, instances metadata.Instances) *Manager {
	return &Manager{Project: proj, Platform: pf, platformCtl: *pfctl, data: instances.ToMap()}
}
