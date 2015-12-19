package hostProvider

import (
    "errors"
    "fmt"
    "io/ioutil"
    "encoding/json"
    "log"
    "net"
    "net/http"

    "golang.org/x/net/context"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    gcloud "github.com/GoogleCloudPlatform/gcloud-golang"
    gce "github.com/GoogleCloudPlatform/gcloud-golang/compute/metadata"
)

type GcloudHost struct {
    HostProvider
    Project string
    Zones []string
    Instances []*GcloudInstance{}
    Client *http.Client
}

type instanceResponse struct {
    kind string
    selfLink string
    id string
    items []string
    nextPageToken string
}

type GcloudAccessConfig struct {
      kind string
      accessType string `json:"type"`
      name string
      natIP string
}

type GcloudNetworkInterface struct {
    network string
    networkIP string
    name string
    accessConfigs []GcloudAccessConfig{}
}

type GcloudDisk struct {
    kind string
    index int
    DiskType string `json:"type"`
    mode string
    source string
    deviceName string
    boot bool
    initializeParams struct {
        diskName string
        sourceImage string
        diskSizeGb uint64
        diskType string
    }
    autoDelete bool
    licenses []string
    DiskInterface string `json:"interface"`
}

type GcloudServiceAccounts struct {
    email string
    scopes []string
}

type GcloudInstance struct {
    kind string
    Id uint64 `json:"id"`
    creationTimestamp string
    Zone string `json:"zone"`
    Status string `json:"status"`
    StatusMessage string `json:"statusMessage"`
    Name string `json:"name"`
    Description string `json:"description"`
    Tags struct { `json:"tags"`
        Items []string `json:"items"`
        Fingerprint []byte `json:"fingerprint"`
    }
    MachineType string `json:"machineType"`
    CanIpForward bool `json:"canIpForward"`
    NetworkInterfaces []GcloudNetworkInterface{} `json:"networkInterfaces"`
    Disks []GcloudDisk{} `json:"disks"`
    Metadata struct { `json:"metaData"`
        kind string `json:"kind"`
        Fingerprint []byte `json:"fingerPrint"`
        Items []struct{ `json:"items"`
            Key string `json:"key"`
            Value string `json:"value"`
        }
    }
    serviceAccounts []GcloudServiceAccounts{}
    selfLink string
    scheduling struct {
        onHostMaintenance string
        automaticRestart bool
        preemptible bool
    }
    cpuPlatform string
}

func NewGCEInstance() *GcloudInstance{
    return &GcloudInstance{}
}

func NewGcloud(projID, jsonFile string) *GcloudHost, error{
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
        Project: p,
        Zones: []string{""},
        Instances: make([]GcloudInstance, 1), // append to this
        Client: client,
    }
    return newHost, nil
}

func(g *GcloudHost) GetServers(namespace string) []GcloudInstance, error{
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
