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
	kind          string
	selfLink      string
	id            string
	items         []string
	nextPageToken string
}

type GcloudAccessConfig struct {
	kind       string
	accessType string `json:"type"`
	name       string
	natIP      string
}

type GcloudNetworkInterface struct {
	Network       string               `json:"network"`
	NetworkIP     string               `json:"networkIP"`
	Name          string               `json:"name"`
	AccessConfigs []GcloudAccessConfig `json:"accessConfigs"`
}

type GcloudDisk struct {
	kind             string
	index            int
	DiskType         string `json:"type"`
	mode             string
	source           string
	deviceName       string
	boot             bool
	initializeParams struct {
		diskName    string
		sourceImage string
		diskSizeGb  uint64
		diskType    string
	}
	autoDelete    bool
	licenses      []string
	DiskInterface string `json:"interface"`
}

type GcloudServiceAccounts struct {
	email  string
	scopes []string
}

type GcloudInstance struct {
	Instance
	kind              string
	Id                uint64 `json:"id"`
	creationTimestamp string
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
		kind        string `json:"kind"`
		Fingerprint []byte `json:"fingerPrint"`
		Items       []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"items"`
	} `json:"metaData"`
	serviceAccounts []GcloudServiceAccounts
	selfLink        string
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
