package mongoInstance

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/cpg1111/kubongo/hostProvider"
	"github.com/cpg1111/kubongo/metadata"
)

type MongoHandler struct {
	ProjectID   string
	Platform    string
	platformCtl hostProvider.HostProvider
	Manager     Manager
	Instances   metadata.Instances
}

func NewHandler(platform, projectID, confPath string, inst metadata.Instances) *MongoHandler {
	var host hostProvider.HostProvider
	var hErr error
	switch platform {
	case "gcloud":
		host = *hostProvider.NewGcloud(projectID, confPath)
	}
	if hErr != nil {
		log.Fatal(hErr)
	}
	return &MongoHandler{
		ProjectID:   projectID,
		Platform:    platform,
		platformCtl: host,
		Manager:     *NewManager(platform, projectID, &host, inst),
		Instances:   inst,
	}
}

func (m MongoHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	log.Println("HTTP request:", *req)
	switch req.Method {
	case "GET":
		m.Get(res, req)
	case "POST":
		m.Post(res, req)
	case "DELETE":
		m.Delete(res, req)
	}
}

type infoRes struct {
	Platform          string             `json:"platform"`
	ProjectName       string             `json:"projectName"`
	NumberOfInstances int                `json:"numberOfInstances"`
	Zones             []string           `json:"zones"`
	Instances         metadata.Instances `json:"instances"`
}

func (m *MongoHandler) Get(res http.ResponseWriter, req *http.Request) {
	numInsts := len(m.Instances)
	payload := &infoRes{
		Platform:          m.Platform,
		ProjectName:       m.ProjectID,
		NumberOfInstances: numInsts,
		Zones:             []string{},
		Instances:         m.Instances,
	}
	header := res.Header()
	encoder := json.NewEncoder(res)
	enErr := encoder.Encode(payload)
	if enErr != nil {
		header.Set("status", fmt.Sprintf("%v", http.StatusInternalServerError))
		res.Write([]byte("{\"error\":\"Internal Server Error\"}"))
	}
}

type InstanceTemplate struct {
	Kind        string `json:"kind" yaml:"kind"` // should equal "Create" or "Register"
	Name        string `json:"name" yaml:"name"`
	Zone        string `json:"zone" yaml:"zone"`
	MachineType string `json:"machineType" yaml:"machineType"`
	SourceImage string `json:"sourceImage" yaml:"sourceImage"`
	Source      string `json:"source" yaml:"source"`
}

// mongoHandler#Post will either create or register an instance based the "kind" field in the request body
// Request Body Srtuct:
// type InstanceTemplate struct{
//     Kind        string `json:"kind"` // should equal "Create" or "Register"
//     name        string
//     Zone        string `json:"zone"`
//     MachineType string `json:"machineType"`
//     SourceImage string `json:"sourceImage"`
//     Source      string `json:"source"`
// }
func (m *MongoHandler) Post(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	reqDecoder := json.NewDecoder(req.Body)
	newInstanceTmpl := &InstanceTemplate{}
	deErr := reqDecoder.Decode(newInstanceTmpl)
	if deErr != nil {
		log.Fatal(deErr)
	}
	if newInstanceTmpl.Kind == "Create" {
		serverRes, serverErr := m.Manager.Create(newInstanceTmpl, &m.Instances)
		if serverErr != nil {
			log.Fatal(serverErr)
		}
		res.Write(serverRes)
	} else {
		serverRes, serverErr := m.Manager.Register(newInstanceTmpl.Zone, newInstanceTmpl.Name, &m.Instances)
		if serverErr != nil {
			log.Fatal(serverRes, serverErr)
		}
		res.Write([]byte("{\"message\":\"201 CREATED\"}"))
	}
}

type DeleteData struct {
	Zone string `json:"zone"`
	Name string `json:"name"`
}

func (m *MongoHandler) Delete(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	reqDecoder := json.NewDecoder(req.Body)
	data := &DeleteData{}
	reqErr := reqDecoder.Decode(data)
	if reqErr != nil {
		log.Fatal(reqErr)
	}
	dErr := m.Manager.Remove(data.Zone, data.Name)
	if dErr != nil {
		log.Fatal(dErr)
	}
	res.Write([]byte("{\"message\":\"200 OK\"}"))
}
