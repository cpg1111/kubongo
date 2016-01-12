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
	Id            string   `json:"id"`
	Items         []string `json:"items"`
	NextPageToken string   `json:"nextPageToken"`
}

type GcloudAccessConfig struct {
	Kind       string `json:"kind"`
	AccessType string `json:"type"`
	Name       string `json:"name"`
	NatIP      string `json:"natIP"`
}

type GcloudNetworkInterface struct {
	Network       string               `json:"network"`
	NetworkIP     string               `json:"networkIP"`
	Name          string               `json:"name"`
	AccessConfigs []GcloudAccessConfig `json:"accessConfigs"`
}

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

type GcloudServiceAccounts struct {
	Email  string   `json:"email"`
	Scopes []string `json:"scopes"`
}

type GcloudInstance struct {
	Instance
	Kind              string `json:"kind"`
	Id                uint64 `json:"id"`
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
	CanIpForward      bool                     `json:"canIpForward"`
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
	CpuPlatform string `json:"cpuPlatform"`
}

func (g GcloudInstance) GetInternalIP() string {
	for i := range g.NetworkInterfaces {
		if g.NetworkInterfaces[i].Name == "eth0" {
			return g.NetworkInterfaces[i].NetworkIP
		}
	}
	return ""
}

func NewGCEInstance() *GcloudInstance {
	return &GcloudInstance{}
}

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
	newInstances := make([]Instance, len(result.items))
	for i := range result.items {
		newInstances[i] = *NewGCEInstance()
		unmarshErr := json.Unmarshal([]byte(result.items[i]), newInstances[i])
		if unmarshErr != nil {
			return nil, unmarshErr
		}
	}
	return newInstances, nil
}

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

type InstanceTemplate struct {
	Name        string `json:"name"`
	MachineType string `json:"machineType"`
	SourceImage string `json:"sourceImage"`
	Source      string `json:"source"`
}

func (g GcloudHost) CreateServer(namespace, zone, name, machineType, sourceImage, source string) ([]byte, error) {
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
	body, bErr := ioutil.ReadAll(res.Body)
	if bErr != nil {
		return nil, bErr
	}
	return body, nil
}

func (g GcloudHost) DeleteServer(namespace, zone, name string) error {
	gcloudRoute := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/instances/%s", project, zone, name)
	req, reqErr := http.NewRequest("DELETE", gcloudRoute, nil)
	res, resErr := g.Client.Do(req)
	if resErr != nil {
		return resErr
	}
	defer res.Body.Close()
	return nil
}
