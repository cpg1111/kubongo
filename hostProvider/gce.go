/*
Copyright 2014 Christian Grabowski All rights reserved.
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

package hostProvider

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gce "google.golang.org/cloud/compute/metadata"
)

// GcloudHost is the HostProvider struct for gcloud, used to control instances on GCE
type GcloudHost struct {
	HostProvider
	Project   string
	Zones     []string
	Instances []*GcloudInstance
	Client    *http.Client
}

type instanceResponse struct {
	Kind          string   `json:"kind"`
	SelfLink      string   `json:"selfLink"`
	ID            string   `json:"id"`
	Items         []string `json:"items"`
	NextPageToken string   `json:"nextPageToken"`
}

// GcloudAccessConfig is a nested struct for network access data
type GcloudAccessConfig struct {
	Kind       string `json:"kind"`
	AccessType string `json:"type"`
	Name       string `json:"name"`
	NatIP      string `json:"natIP"`
}

// GcloudNetworkInterface is a nested struct for network access data
type GcloudNetworkInterface struct {
	Network       string               `json:"network"`
	NetworkIP     string               `json:"networkIP"`
	Name          string               `json:"name"`
	AccessConfigs []GcloudAccessConfig `json:"accessConfigs"`
}

// GcloudDisk is a struct for data about an instances disk(s)
type GcloudDisk struct {
	Kind             string `json:"kind"`
	Index            int    `json:"index"`
	DiskType         string `json:"type"`
	Mode             string `json:"mode"`
	Source           string `json:"source"`
	DeviceName       string `json:"deviceName"`
	Boot             bool   `json:"boot"`
	InitializeParams struct {
		DiskName    string `json:"diskName"`
		SourceImage string `json:"sourceImage"`
		DiskSizeGb  uint64 `json:"diskSizeGb"`
		DiskType    string `json:"diskType"`
	} `json:"initializeParams"`
	AutoDelete    bool     `json:"autoDelete"`
	Licenses      []string `json:"licenses"`
	DiskInterface string   `json:"interface"`
}

// GcloudServiceAccounts is a struct for GCE service account data
type GcloudServiceAccounts struct {
	Email  string   `json:"email"`
	Scopes []string `json:"scopes"`
}

// GcloudInstance is the outer most struct for GCE instance data
type GcloudInstance struct {
	Instance
	Kind              string `json:"kind"`
	ID                uint64 `json:"id"`
	CreationTimestamp string `json:"creationTimestamp"`
	Zone              string `json:"zone"`
	Status            string `json:"status"`
	StatusMessage     string `json:"statusMessage"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Tags              struct {
		Items       []string `json:"items"`
		Fingerprint []byte   `json:"fingerprint"`
	} `json:"tags"`
	MachineType       string                   `json:"machineType"`
	CanIPForward      bool                     `json:"canIpForward"`
	NetworkInterfaces []GcloudNetworkInterface `json:"networkInterfaces"`
	Disks             []GcloudDisk             `json:"disks"`
	Metadata          struct {
		Kind        string `json:"kind"`
		Fingerprint []byte `json:"fingerPrint"`
		Items       []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"items"`
	} `json:"metaData"`
	ServiceAccounts []GcloudServiceAccounts `json:"serviceAccounts"`
	SelfLink        string                  `json:"selfLink"`
	Scheduling      struct {
		OnHostMaintenance string `json:"onHostMaintenance"`
		AutomaticRestart  bool   `json:"automaticRestart"`
		Preemptible       bool   `json:"preemptible"`
	} `json:"scheduling"`
	CPUPlatform string `json:"cpuPlatform"`
}

// GetInternalIP returns the internal IP of its instance
func (g GcloudInstance) GetInternalIP() string {
	for i := range g.NetworkInterfaces {
		if g.NetworkInterfaces[i].Name == "eth0" {
			return g.NetworkInterfaces[i].NetworkIP
		}
	}
	return ""
}

// NewGCEInstance returns a new GceInstance struct
func NewGCEInstance() *GcloudInstance {
	return &GcloudInstance{}
}

// NewGcloud returns a new GCE HostProvider
func NewGcloud(p, jsonFile string) *GcloudHost {
	var client *http.Client
	if jsonFile != "" {
		jsonKey, err := ioutil.ReadFile(jsonFile)
		if err != nil {
			log.Fatal(err)
		}
		conf, err := google.JWTConfigFromJSON(jsonKey, "https://www.googleapis.com/auth/compute")
		if err != nil {
			log.Fatal(err)
		}
		client = conf.Client(oauth2.NoContext)
	} else if gce.OnGCE() {
		client = &http.Client{
			Transport: &oauth2.Transport{
				Source: google.ComputeTokenSource(""),
			},
		}
	} else {
		log.Fatal(errors.New("Could Not Auth with Google"))
	}
	newHost := &GcloudHost{
		Project:   p,
		Zones:     []string{""},
		Instances: make([]*GcloudInstance, 1), // append to this
		Client:    client,
	}
	return newHost
}

// GetServers returns a slice of servers in GCE
func (g GcloudHost) GetServers(namespace string) ([]Instance, error) {
	gcloudRoute := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/instances", namespace)
	res, resErr := g.Client.Get(gcloudRoute)
	if resErr != nil {
		return nil, resErr
	}
	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)
	result := &instanceResponse{}
	decodeErr := decoder.Decode(result)
	if decodeErr != nil {
		return nil, decodeErr
	}
	newInstances := make([]Instance, len(result.Items))
	for i := range result.Items {
		newInstances[i] = *NewGCEInstance()
		unmarshErr := json.Unmarshal([]byte(result.Items[i]), newInstances[i])
		if unmarshErr != nil {
			return nil, unmarshErr
		}
	}
	return newInstances, nil
}

// GetServer returns a specific server on GCE
func (g GcloudHost) GetServer(project, zone, name string) (Instance, error) {
	gcloudRoute := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/instances/%s", project, zone, name)
	res, resErr := g.Client.Get(gcloudRoute)
	if resErr != nil {
		return nil, resErr
	}
	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)
	result := &GcloudInstance{}
	decodeErr := decoder.Decode(result)
	if decodeErr != nil {
		return nil, decodeErr
	}
	return result, nil
}

// InstanceTemplate is a struct for request data to create a server
type InstanceTemplate struct {
	Name        string `json:"name"`
	MachineType string `json:"machineType"`
	SourceImage string `json:"sourceImage"`
	Source      string `json:"source"`
}

// CreateServer will send a POST to the GCE api to create an instance
func (g GcloudHost) CreateServer(namespace, zone, name, machineType, sourceImage, source string) (Instance, error) {
	gcloudRoute := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/instances", namespace, zone)
	newInstance := &InstanceTemplate{
		Name:        name,
		MachineType: machineType,
		SourceImage: sourceImage,
		Source:      source,
	}
	reqBytes, bErr := json.Marshal(newInstance)
	if bErr != nil {
		return nil, bErr
	}
	reqPayload := bytes.NewBuffer(reqBytes)
	res, resErr := g.Client.Post(gcloudRoute, "application/JSON", reqPayload)
	if resErr != nil {
		return nil, resErr
	}
	defer res.Body.Close()
	decoder := json.NewDecoder(res.Body)
	result := &GcloudInstance{}
	decodeErr := decoder.Decode(result)
	if decodeErr != nil {
		return nil, decodeErr
	}
	return result, nil
}

// DeleteServer will send GCE a DELETE to delete a specific instance
func (g GcloudHost) DeleteServer(namespace, zone, name string) error {
	gcloudRoute := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/instances/%s", namespace, zone, name)
	req, reqErr := http.NewRequest("DELETE", gcloudRoute, nil)
	if reqErr != nil {
		return reqErr
	}
	res, resErr := g.Client.Do(req)
	if resErr != nil {
		return resErr
	}
	defer res.Body.Close()
	return nil
}
