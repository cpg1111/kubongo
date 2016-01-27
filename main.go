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

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/cpg1111/kubongo/metadata"
	mongo "github.com/cpg1111/kubongo/mongoInstance"
)

func main() {
	var (
		platform     = flag.String("-platform", "local", "Set which cloud platform to use, defaults to gcloud")
		project      = flag.String("-project", "", "Set which project/organization to use, defaults to empty")
		platConfPath = flag.String("-platform-config", "./config.json", "Set the path to a json config for cloud platform, defaults to ./config.json")
		port         = flag.Int("-port", 8888, "Set the port number for kubungo's api server to listen on, defaults to 8888")
		initMaster   = flag.String("init-master", "127.0.0.1:28017", "Set the IP address and port of the master or mongos for monitoring, default is 127.0.0.1:28017")
		masterZone   = flag.String("-master-zone", "local", "Set default zone/region for master mongo instance, default is us-central1-f")
		help         = flag.Bool("-help", false, "Prints info on Kubongo")
	)
	flag.Parse()
	if *help {
		flag.PrintDefaults()
	}
	portNum := fmt.Sprintf(":%v", *port)
	server := http.NewServeMux()
	instances := metadata.New(nil)
	go func() {
		for {
			log.Println(instances.ToMap()["master"])
		}
	}()
	mongoHandler := mongo.NewHandler(*platform, *project, *platConfPath, *instances)
	server.Handle("/instances", mongoHandler)
	log.Println("Kubongo Process started and is listening on port", *port)
	log.Println("Registering", *initMaster)
	mongoHandler.Manager.Register(*masterZone, "master", instances)
	log.Println("monitoring", *initMaster)
	mongoHandler.Manager.Monitor(initMaster, instances)
	log.Fatal(http.ListenAndServe(portNum, server))
}
