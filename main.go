package main

import (
	"flag"
	"log"
	"net/http"

	mongo "github.com/cpg1111/kubongo/mongoInstance"
)

func main() {
	var (
		platform     = flag.String("-platform", "gcloud", "Set which cloud platform to use, defaults to gcloud")
		project      = flag.String("-project", "", "Set which project/organization to use, defaults to empty")
		platConfPath = flag.String("-platform-config", "./config.json", "Set the path to a json config for cloud platform, defaults to ./config.json")
	)
	server := http.NewServeMux()
	mongoHandler := mongo.NewHandler(*platform, *project, *platConfPath)
	server.Handle("/instances", mongoHandler)
	log.Fatal(http.ListenAndServe(":8888", nil))
}
