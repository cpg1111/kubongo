/*
Copyright 2015 Christian Grabowski All rights reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mongoInstance

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cpg1111/kubongo/hostProvider"
	kube "github.com/cpg1111/kubongo/kubeClient"
	"github.com/cpg1111/kubongo/metadata"
)

// Manager manages mongo instances
type Manager struct {
	// Name of Project in cloud host
	Project string
	// Platform for cloud provider
	Platform string
	// Struct to Controls cloud provider actions
	platformCtl hostProvider.HostProvider
	// list of registered instances
	data *metadata.Instances
	// controller for talking to the Kubernetes api
	kubeCtl *kube.Controller
}

func addToInstances(instances *metadata.Instances, newServer hostProvider.Instance) {
	instances = metadata.AddInstance(instances, newServer)
}

// SetKubeCtl sets the kubernetes api controller, this is not done in the New() function so that the manager depends souly on mongo-side things
// and only needs a kubeClient controller for updates
func (m *Manager) SetKubeCtl(ktl *kube.Controller) {
	m.kubeCtl = ktl
}

// Create a new mongo instance
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
	m.data = instances
	newServerJSON, jErr := json.Marshal(&newServer)
	return newServerJSON, jErr
}

// Register an existing mongo instance
func (m *Manager) Register(zone, name string, instances *metadata.Instances) ([]byte, error) {
	var (
		newServer hostProvider.Instance
		serverErr error
	)
	if strings.Contains(zone, "local") {
		newServer, serverErr = m.platformCtl.CreateServer(m.Platform, zone, name, "27017", "mongo", "mongo")
	} else {
		newServer, serverErr = m.platformCtl.GetServer(m.Platform, zone, name)
	}
	if serverErr != nil {
		return nil, serverErr
	}
	addToInstances(instances, newServer)
	m.data = instances
	newServerJSON, jErr := json.Marshal(&newServer)
	return newServerJSON, jErr
}

// Remove existing mongo instance
func (m *Manager) Remove(zone, name string) error {
	dErr := m.platformCtl.DeleteServer(m.Platform, zone, name)
	newData := metadata.RemoveInstance(*m.data, m.data.ToMap()[name])
	m.data = &newData
	return dErr
}

func localMasterTmpl() *InstanceTemplate {
	return &InstanceTemplate{
		Kind:        "Create",
		Name:        "master",
		MachineType: "27017",
		SourceImage: "",
		Source:      "",
	}
}

func gcloudMasterTmpl() *InstanceTemplate {
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

func (m *Manager) newMaster(rStatus, nStatus chan error, success chan []byte, instances *metadata.Instances) error {
	log.Println("manager:132 NEWMASTER")
	uncastMaster := m.data.ToMap()["master"]
	if uncastMaster == nil {
		rStatus <- nil
	} else {
		log.Println("manager:137 REMOVE")
		master := uncastMaster.(hostProvider.LocalInstance)
		rStatus <- m.Remove(master.Zone, master.Name)
	}
	newInstance := localMasterTmpl()
	newBytes, nErr := m.Create(newInstance, instances)
	if nErr != nil {
		nStatus <- nErr
	} else {
		success <- newBytes
	}
	return nil
}

// Monitor master mongo instance
func (m *Manager) Monitor(masterIP *string, instances *metadata.Instances) error {
	monitor := newMonitor(masterIP)
	isHealthy := true
	healthChannel := make(chan bool)
	for isHealthy {
		go monitor.HealthCheck(healthChannel)
		isHealthy = <-healthChannel
		log.Println("manager:159 health:", isHealthy)
		time.Sleep(time.Second * 3)
	}
	removeStatus := make(chan error)
	newMasterStatus := make(chan error)
	successStatus := make(chan []byte)
	go m.newMaster(removeStatus, newMasterStatus, successStatus, instances)
	for {
		select {
		case m1 := <-removeStatus:
			if m1 != nil {
				return m1
			}
			break
		case m2 := <-newMasterStatus:
			if m2 != nil {
				log.Fatal(m2)
				return m2
			}
			break
		case m3 := <-successStatus:
			log.Println("manager:180 Created", string(m3))
			return m.Monitor(masterIP, instances)
		}
	}
}

// NewManager creates a new manager struct
func NewManager(proj, pf string, pfctl *hostProvider.HostProvider, instances *metadata.Instances) *Manager {
	return &Manager{Project: proj, Platform: pf, platformCtl: *pfctl, data: instances}
}
