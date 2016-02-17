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
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	mongo "github.com/cpg1111/kubongo/mongoInstance"

	yaml "gopkg.in/yaml.v2"
)

func printHelp() {
	log.Println("NAME")
	log.Println("Kubongoctl")
	log.Println(" ")
	log.Println("USAGE")
	log.Println("kubongoctl [options] ACTION [action arguments] TARGET | TARGET EDIT")
	log.Println(" ")
	log.Println("OPTIONS")
	log.Println("--platform-config", "./config.json", "Set the path to a json config for cloud platform, defaults to ./config.json")
	log.Println("--port", "8888", "Set the port number that kubungo's api server is listening on, defaults to 8888")
	log.Println("--host", "127.0.0.1", "Set the IP address of Kubongo's api server")
	os.Exit(0)
}

// DecodeInstanceFile will decode a JSON or YAML into an instance struct
// See https://github.com/cpg1111/kubongo/mongoInstance/ for the fields of an InstanceTemplate
func DecodeInstanceFile(filename string, file []byte) (*mongo.InstanceTemplate, error) {
	instance := &mongo.InstanceTemplate{}
	isYAML := (strings.Contains(filename, ".yaml") || strings.Contains(filename, ".yml"))
	isJSON := strings.Contains(filename, ".json")
	var err error
	if isYAML {
		err = yaml.Unmarshal(file, instance)
	} else if isJSON {
		err = json.Unmarshal(file, instance)
	} else {
		return nil, errors.New("input was not yaml or json")
	}
	return instance, err
}

// Create will send a post to the Specified endpoint to create the resource that corolates to the endpoint
func Create(url string) (*http.Response, error) {
	var instanceConf string
	if len(os.Args) > 3 {
		instanceConf = os.Args[3]
		confInfo, confErr := os.Stat(instanceConf)
		if confInfo != nil || confErr == nil {
			confBytes, readErr := ioutil.ReadFile(instanceConf)
			if readErr != nil {
				return nil, errors.New("could not open file to create instance")
			}
			inst, instErr := DecodeInstanceFile(instanceConf, confBytes)
			if instErr != nil {
				return nil, instErr
			}
			instPayload, payloadErr := json.Marshal(inst)
			if payloadErr != nil {
				return nil, payloadErr
			}
			return http.Post(url, "json", bytes.NewBuffer(instPayload))
		}
	}
	return nil, errors.New("no input given to create instance")
}

// Destroy will destroy a resource that corolates with the specified endpoint
func Destroy(url string) (*http.Response, error) {
	if len(os.Args) > 3 {
		zoneName := os.Args[3]
		if len(os.Args) > 4 {
			instanceName := os.Args[3]
			req, reqErr := http.NewRequest("DELETE", url, bytes.NewBuffer([]byte(fmt.Sprintf("{\"zone\":\"%s\", \"name\":\"%s\"}", zoneName, instanceName))))
			if reqErr != nil {
				return nil, reqErr
			}
			req.Header.Set("content-type", "json")
			client := &http.Client{}
			res, resErr := client.Do(req)
			if resErr != nil {
				return nil, resErr
			}
			return res, nil
		}
		return nil, errors.New("no instance name given")
	}
	return nil, errors.New("no instance zone given")
}

// Request will send a request to the Kubongo server
func Request(host, port, method, endpoint string) (res *http.Response, resErr error) {
	targetURL := fmt.Sprintf("http://%s:%s/%s", host, port, endpoint)
	switch method {
	case "info":
		res, resErr = http.Get(targetURL)
		break
	case "create":
		res, resErr = Create(targetURL)
		break
	case "delete":
		res, resErr = Destroy(targetURL)
		break
	default:
		res = nil
		resErr = errors.New("Invalid Method")
		break
	}
	return
}

func main() {
	var (
		platConfPath = flag.String("-platform-config", "./config.json", "Set the path to a json config for cloud platform, defaults to ./config.json")
		port         = flag.Int("-port", 8888, "Set the port number that kubungo's api server is listening on, defaults to 8888")
		host         = flag.String("-host", "127.0.0.1", "Set the IP address of Kubongo's api server")
		help         = flag.Bool("-help", false, "Prints info on Kubongoctl")
	)
	if *help {
		printHelp()
	}
	confInfo, confErr := os.Stat(*platConfPath)
	log.Println(confInfo)
	if confErr == nil {
		confBytes, readErr := ioutil.ReadFile(*platConfPath)
		if readErr != nil {
			panic(readErr)
		}
		log.Println(string(confBytes))
	}
	var method string
	if len(os.Args) > 1 {
		method = os.Args[1]
	}
	var endpoint string
	if len(os.Args) > 2 {
		endpoint = os.Args[2]
	}
	log.Println(Request(*host, fmt.Sprintf("%v", *port), method, endpoint))
}
