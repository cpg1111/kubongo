package main

import (
	"net/http"

	"github.com/cpg1111/kubongo/mongoInstance"
)

func main() {
	server := http.NewServeMux()
	server.Handle()
}
