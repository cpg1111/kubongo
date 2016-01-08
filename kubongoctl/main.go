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

func decodeInstanceFile(filename string, file []byte) (*mongo.InstanceTemplate, error) {
	instance := &mongo.InstanceTemplate{}
	isYaml := (strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml"))
	isJson := strings.HasSuffix(filename, ".json")
	var err error
	if isYaml {
		err = yaml.Unmarshal(file, instance)
	} else if isJson {
		err = json.Unmarshal(file, instance)
	} else {
		return nil, errors.New("Input was not yaml or json")
	}
	return instance, err
}

func create(url string) (*http.Response, error) {
	var instanceConf string
	if len(os.Args) > 3 {
		instanceConf = os.Args[3]
		confInfo, confErr := os.Stat(instanceConf)
		if confInfo != nil || confErr == nil {
			confBytes, readErr := ioutil.ReadFile(instanceConf)
			if readErr != nil {
				return nil, errors.New("Could not open file to create instance")
			}
			inst, instErr := decodeInstanceFile(instanceConf, confBytes)
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
	return nil, errors.New("No input given to create instance")
}

func edit(url string) (*http.Response, error) {
	var instanceConf string
	if len(os.Args) > 3 {
		instanceConf = os.Args[3]
		confInfo, confErr := os.Stat(instanceConf)
		if confInfo != nil || confErr == nil {
			confBytes, readErr := ioutil.ReadFile(instanceConf)
			if readErr != nil {
				return nil, errors.New("Could not open file to create instance")
			}
			inst, instErr := decodeInstanceFile(instanceConf, confBytes)
			if instErr != nil {
				return nil, instErr
			}
			instPayload, payloadErr := json.Marshal(inst)
			if payloadErr != nil {
				return nil, payloadErr
			}
			req, reqErr := http.NewRequest("PUT", url, bytes.NewBuffer(instPayload))
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
		return nil, confErr
	}
	return nil, errors.New("No edit file given")
}

func destroy(url string) (*http.Response, error) {
	var instanceName string
	if len(os.Args) > 3 {
		instanceName = os.Args[3]
		req, reqErr := http.NewRequest("DELETE", url, bytes.NewBuffer([]byte(fmt.Sprintf("{\"name\":\"%s\"}", instanceName))))
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
	return nil, errors.New("No instance name given")
}

func request(host, port, method, endpoint string) (res *http.Response, resErr error) {
	targetURL := fmt.Sprintf("http://%s:%s/%s", host, port, endpoint)
	switch method {
	case "info":
		res, resErr = http.Get(targetURL)
		break
	case "create":
		res, resErr = create(targetURL)
		break
	case "edit":
		res, resErr = edit(targetURL)
		break
	case "delete":
		res, resErr = destroy(targetURL)
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
	log.Println(request(*host, fmt.Sprintf("%v", *port), method, endpoint))
}
