package mongoInstance

import (
	"log"

	"github.com/cpg1111/kubongo/hostProvider"
	"github.com/cpg1111/kubongo/metadata"
)

type Manager struct {
	Platform    string
	platformCtl hostProvider.HostProvider
}

func (m *Manager) Register(zone, name string, instances *metadata.Instances) {
	newServer, serverErr := m.platformCtl.GetServer(m.Platform, zone, name)
	if serverErr != nil {
		log.Fatal(serverErr)
	}
	newInstances := metadata.AddInstance(*instances, newServer)
	instances = &newInstances
}

func NewManager(pf string, pfctl *hostProvider.HostProvider) *Manager {
	return &Manager{Platform: pf, platformCtl: *pfctl}
}
