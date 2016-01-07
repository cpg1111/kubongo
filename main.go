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
		platform     = flag.String("-platform", "gcloud", "Set which cloud platform to use, defaults to gcloud")
		project      = flag.String("-project", "", "Set which project/organization to use, defaults to empty")
		platConfPath = flag.String("-platform-config", "./config.json", "Set the path to a json config for cloud platform, defaults to ./config.json")
		port         = flag.Int("-port", 8888, "Set the port number for kubungo's api server to listen on, defaults to 8888")
		help         = flag.Bool("-help", false, "Prints info on Kubongo")
	)
	flag.Parse()
	if *help {
		flag.PrintDefaults()
	}
	portNum := fmt.Sprintf(":%v", *port)
	server := http.NewServeMux()
	instances := metadata.New(nil)
	mongoHandler := mongo.NewHandler(*platform, *project, *platConfPath, instances)
	server.Handle("/instances", mongoHandler)
	log.Println("Kubongo Process started and is listening on port", portNum)
	log.Fatal(http.ListenAndServe(portNum, server))
}
