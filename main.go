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

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	kube "github.com/cpg1111/kubongo/kubeClient"
	"github.com/cpg1111/kubongo/metadata"
	mongo "github.com/cpg1111/kubongo/mongoInstance"
)

func main() {
	var (
		platform        = flag.String("-platform", "local", "Set which cloud platform to use, defaults to gcloud")
		project         = flag.String("-project", "", "Set which project/organization to use, defaults to empty")
		platConfPath    = flag.String("-platform-config", "./config.json", "Set the path to a json config for cloud platform, defaults to ./config.json")
		port            = flag.Int("-port", 8888, "Set the port number for kubungo's api server to listen on, defaults to 8888")
		kubeEnvVarName  = flag.String("-kube-env-var-name", "DB_CONNECT_STRING", "Set the environment variable name for mongo's service discovery in Kubernetes, defaults to \"DB_CONNECT_STRING\"")
		kubeNamespace   = flag.String("-kube-namespace", "default", "set the Kubernetes namespace to update with the mongo endpoint, defaults to \"default\"")
		initKubeMaster  = flag.String("init-kube-master", "127.0.0.1:8080", "Set the IP address and port of the Kubernetes master, defaults to 127.0.0.1:8080")
		initMongoMaster = flag.String("init-mongo-master", "127.0.0.1:27017", "Set the IP address and port of the master mongod or mongos for monitoring, default is 127.0.0.1:27017")
		masterZone      = flag.String("-master-zone", "local", "Set default zone/region for master mongo instance, default is us-central1-f")
		help            = flag.Bool("-help", false, "Prints info on Kubongo")
	)
	flag.Parse()
	if *help {
		flag.PrintDefaults()
	}
	portNum := fmt.Sprintf(":%v", *port)
	server := http.NewServeMux()
	instances := metadata.New(nil)
	mongoHandler := mongo.NewHandler(*platform, *project, *platConfPath, *instances)
	server.Handle("/instances", mongoHandler)
	log.Println("main:49 Kubongo Process started and is listening on port", *port)
	kubeClient := kube.New(*initKubeMaster, *kubeNamespace, *kubeEnvVarName)
	pingErr := kubeClient.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	mongoHandler.Manager.SetKubeCtl(kubeClient)
	log.Println("main:56 Registering", *initMongoMaster)
	mongoHandler.Manager.Register(*masterZone, "master", instances)
	log.Println("main:58 monitoring", *initMongoMaster)
	mongoHandler.Manager.Monitor(initMongoMaster, instances)
	log.Fatal(http.ListenAndServe(portNum, server))
}
