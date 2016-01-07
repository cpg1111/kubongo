package mongoInstance

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/cpg1111/kubongo/hostProvider"
	"github.com/cpg1111/kubongo/metadata"
)

type MongoHandler struct {
	http.Handler
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
		Platform:    platform,
		platformCtl: host,
		Manager:     *NewManager(platform, &host),
		Instances:   inst,
	}
}

func (m *MongoHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		m.Get(res, req)
	case "POST":
		m.Post(res, req)
	case "PUT":
		m.Put(res, req)
	case "DELETE":
		m.Delete(res, req)
	}
}

func (m *MongoHandler) Get(res http.ResponseWriter, req *http.Request) {

}

type InstanceTemplate struct {
	Kind        string `json:"kind"`
	name        string
	Zone        string `json:"zone"`
	MachineType string `json:"machineType"`
	SourceImage string `json:"sourceImage"`
	Source      string `json:"source"`
}

// mongoHandler#Post will either create or register an instance based the "kind" field in the request body
// Request Body Srtuct:
// type InstanceTemplate struct{
//     Kind string `json:"kind"` // should equal "Create" or "Regist"
//     name string
//     Zone string `json:"zone"`
//     MachineType string `json:"machineType"`
//     SourceImage string `json:"sourceImage"`
//     Source string `json:"source"`
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
		serverRes, serverErr := m.platformCtl.CreateServer(
			m.Platform,
			newInstanceTmpl.Zone,
			newInstanceTmpl.name,
			newInstanceTmpl.MachineType,
			newInstanceTmpl.SourceImage,
			newInstanceTmpl.Source,
		)
		if serverErr != nil {
			log.Fatal(serverErr)
		}
		res.Write(serverRes)
	} else {
		m.Manager.Register(newInstanceTmpl.Zone, newInstanceTmpl.name, &m.Instances)
		res.Write([]byte("{\"message\":\"201 Created\"}"))
	}
}

func (m *MongoHandler) Put(res http.ResponseWriter, req *http.Request) {

}

func (m *MongoHandler) Delete(res http.ResponseWriter, req *http.Request) {

}
