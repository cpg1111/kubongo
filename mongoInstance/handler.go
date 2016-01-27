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

package mongoInstance

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/cpg1111/kubongo/hostProvider"
	"github.com/cpg1111/kubongo/metadata"
)

// MongoHandler handles http request for mongo instances
type MongoHandler struct {
	ProjectID   string
	Platform    string
	platformCtl hostProvider.HostProvider
	Manager     Manager
	Instances   metadata.Instances
}

// NewHandler creates a new mongo handler struct
func NewHandler(platform, projectID, confPath string, inst metadata.Instances) *MongoHandler {
	var host hostProvider.HostProvider
	var hErr error
	log.Println(platform == "local")
	switch platform {
	case "GCE":
		host = *hostProvider.NewGcloud(projectID, confPath)
		hErr = nil
	case "local":
		host = *hostProvider.NewLocal()
		hErr = nil
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

// ServeHTTP serves http for mongo instance
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

// Get for GET method on /instances
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

// InstanceTemplate is a struct for creating instances
type InstanceTemplate struct {
	Kind        string `json:"kind" yaml:"kind"` // should equal "Create" or "Register"
	Name        string `json:"name" yaml:"name"`
	Zone        string `json:"zone" yaml:"zone"`
	MachineType string `json:"machineType" yaml:"machineType"`
	SourceImage string `json:"sourceImage" yaml:"sourceImage"`
	Source      string `json:"source" yaml:"source"`
}

// Post will either create or register an instance based the "kind" field in the request body
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

// DeleteData is req data to delete instance
type DeleteData struct {
	Zone string `json:"zone"`
	Name string `json:"name"`
}

// Delete will delete instances
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
