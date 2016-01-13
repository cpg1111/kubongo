package mongoInstance

import (
	"encoding/json"
	"time"

	"github.com/cpg1111/kubongo/hostProvider"
	"github.com/cpg1111/kubongo/metadata"
)

type Manager struct {
	Project     string
	Platform    string
	platformCtl hostProvider.HostProvider
}

func (m *Manager) Create(newInstanceTmpl *InstanceTemplate) ([]byte, error) {
	newServer, serverErr := m.platformCtl.CreateServer(
		m.Platform,
		newInstanceTmpl.Zone,
		newInstanceTmpl.Name,
		newInstanceTmpl.MachineType,
		newInstanceTmpl.SourceImage,
		newInstanceTmpl.Source,
	)
	return newServer, serverErr
}

func (m *Manager) Register(zone, name string, instances *metadata.Instances) ([]byte, error) {
	newServer, serverErr := m.platformCtl.GetServer(m.Platform, zone, name)
	if serverErr != nil {
		return nil, serverErr
	}
	newInstances := metadata.AddInstance(*instances, newServer)
	instances = &newInstances
	newServerJSON, jErr := json.Marshal(&newServer)
	return newServerJSON, jErr
}

func (m *Manager) Remove(zone, name string) error {
	dErr := m.platformCtl.DeleteServer(m.Platform, zone, name)
	return dErr
}

func (m *Manager) Monitor(masterIP *string) {
	monitor := newMonitor(masterIP)
	isHealthy := true
	healthChannel := make(chan bool)
	for isHealthy {
		go monitor.HealthCheck(healthChannel)
		isHealthy = <-healthChannel
		time.Sleep(time.Second * 3)
	}
}

func NewManager(proj, pf string, pfctl *hostProvider.HostProvider) *Manager {
	return &Manager{Project: proj, Platform: pf, platformCtl: *pfctl}
}
