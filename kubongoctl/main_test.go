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
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

type testFiles struct {
	JsonFile   *os.File
	YamlFile   *os.File
	JsonString string
	YamlString string
	JsonLen    int
	YamlLen    int
}

func writeFiles(name string, t *testing.T) *testFiles {
	testJsonFile, jFileErr := ioutil.TempFile(os.TempDir(), fmt.Sprintf("j%s.json", name))
	if jFileErr != nil {
		t.Error(jFileErr)
	}
	testYamlFile, yFileErr := ioutil.TempFile(os.TempDir(), fmt.Sprintf("y%s.yaml", name))
	if yFileErr != nil {
		t.Error(yFileErr)
	}
	jsonString := `
    {
        "kind" : "Register",
        "name" : "test",
        "zone" : "us-central1-f",
        "machineType" : "zones/us-central1-f/machineTypes/n1-standard-1",
        "sourceImage" : "",
        "source" : ""
    }

    `
	yamlString := `
    kind: "Register"
    name: "test"
    zone: "us-central1-f"
    machineType: "zones/us-central1-f/machineTypes/n1-standard-1"
    sourceImage: ""
    source: ""

    `
	j, jErr := testJsonFile.WriteString(jsonString)
	y, yErr := testYamlFile.WriteString(yamlString)
	if jErr != nil {
		t.Error(jErr)
	}
	if yErr != nil {
		t.Error(yErr)
	}
	return &testFiles{
		JsonFile:   testJsonFile,
		YamlFile:   testYamlFile,
		JsonString: jsonString,
		YamlString: yamlString,
		JsonLen:    j,
		YamlLen:    y,
	}
}

func TestDecodeFileInstanceFile(t *testing.T) {
	testFiles := writeFiles("TestFile", t)
	defer os.Remove(testFiles.JsonFile.Name())
	defer os.Remove(testFiles.YamlFile.Name())
	jData, jErr := ioutil.ReadFile(testFiles.JsonFile.Name())
	yData, yErr := ioutil.ReadFile(testFiles.YamlFile.Name())
	t.Log(testFiles.JsonFile.Name(), testFiles.JsonLen, testFiles.YamlLen, string(jData), string(yData))
	if jErr != nil {
		t.Error(jErr)
	}
	if yErr != nil {
		t.Error(yErr)
	}
	if string(jData) != testFiles.JsonString {
		t.Error("read JSON File Does Not Match Input JSON File")
	}
	if string(yData) != testFiles.YamlString {
		t.Error("read YAML File Does Not Match Input YAML File")
	}
	jInstance, instErr := DecodeInstanceFile(testFiles.JsonFile.Name(), jData)
	if instErr != nil || jInstance.Name != "test" {
		t.Log(instErr)
		t.Error("instance Does Not Match Input JSON File")
	}
	yInstance, instErr := DecodeInstanceFile(testFiles.YamlFile.Name(), yData)
	if instErr != nil || yInstance.Name != "test" {
		t.Log(instErr)
		t.Error("instance Does Not Match Input Yaml File")
	}
}
