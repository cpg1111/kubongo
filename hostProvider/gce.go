package hostProvider

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gcloud "google.golang.org/cloud"
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
	Network       string `json:"network"`
	NetworkIP     string `json:"networkIP"`
	Name          string `json:"name"`
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
	instance
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

func(g *GcloudInstance) GetInternalIP() string{
    for i := range g.NetworkInterfaces {
        if g.NetworkInterfaces[i].Name == "eth0" {
            return g.NetworkInterfaces[i].NetworkIP
        }
    }
}

func NewGCEInstance() *GcloudInstance {
	return &GcloudInstance{}
}

func NewGcloud(projID, jsonFile string) (*GcloudHost, error) {
	var projID string
	var client *http.Client
	if jsonFile != "" {
		jsonKey, err := ioutil.ReadFile(jsonFile)
		if err != nil {
			return nil, err
		}
		conf, err := google.JWTConfigFromJSON(jsonKey, "https://www.googleapis.com/auth/compute")
		if err != nil {
			return nil, err
		}
		client = conf.Client(oauth2.NoContext)
	} else if gce.OnGCE() {
		client = &http.Client{
			Transport: &oauth2.Transport{
				Source: google.ComputeTokenSource(""),
			},
		}
	} else {
		return nil, errors.New("Could Not Auth with Google")
	}
	newHost := &GcloudHost{
		Project:   p,
		Zones:     []string{""},
		Instances: make([]GcloudInstance, 1), // append to this
		Client:    client,
	}
	return newHost, nil
}

func (g *GcloudHost) GetServers(namespace string) ([]GcloudInstance, error) {
	gcloudRoute, rErr := gce.Get(fmt.Sprintf("/projects/%s/zones/instances"))
	if rErr != nil {
		return nil, rErr
	}
	res, resErr := g.Client.Get(gcloudRoute)
	if resErr != nil {
		return nil, rErr
	}
	defer res.Body.Close()
	body, bErr := ioutil.ReadAll(res.Body)
	if bErr != nil {
		return nil, bErr
	}
	decoder := json.NewDecoder(body)
	result := &instanceResponse{}
	decoder.Decode(result)
	return result.items, nil
}
