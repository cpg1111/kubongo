package main

import (
	"log"
	"net/http"

	mongo "github.com/cpg1111/kubongo/mongoInstance"
)

func main() {
	server := http.NewServeMux()
	mongoHandler := mongo.NewHandler()
	server.Handle("/instances", mongoHandler)
	log.Fatal(http.ListenAndServe(":8888", nil))
}
