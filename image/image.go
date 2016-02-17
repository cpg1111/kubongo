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

package image

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

var waitgroup sync.WaitGroup

// ImageManager is a struct with data for OS images for instances
type ImageManager struct {
	Platform   string
	OS         string
	SSHCommand *exec.Cmd
}

func getLocalOS() string {
	goos := os.Getenv("GOOS")
	osNameCMD := exec.Command("cat", "/etc/*-release", "|", "grep", "PRETTY_NAME")
	var osNameOut bytes.Buffer
	osNameCMD.Stdout = &osNameOut
	err := osNameCMD.Run()
	if err != nil && goos != "darwin" { // uses GOOS in the event it is OSX
		log.Fatal(err)
	} else if goos != "darwin" {
		return osNameOut.String()
	} else {
		return goos
	}
	return ""
}

func formatPrettyName(prettyName string) string {
	var distro, majorVer, minorVer string
	if strings.Contains(prettyName, "Ubuntu") {
		distro = "ubuntu"
	} else if strings.Contains(prettyName, "Debian") {
		distro = "debian"
	} else if strings.Contains(prettyName, "CentOs") {
		distro = "centos"
	} else if strings.Contains(prettyName, "CoreOs") {
		distro = "coreos"
	} else {
		distro = prettyName //log.Fatal("Your OS is not currently supported, to change this, please create an issue @ https://github.com/cpg1111/kubongo/issues")
	}
	rx := regexp.MustCompile("([0-9]+)")
	matches := rx.FindStringSubmatch(prettyName)
	matchesLen := len(matches)
	if matchesLen > 0 {
		majorVer = matches[0]
	}
	if matchesLen > 1 {
		minorVer = matches[1]
	}
	return fmt.Sprintf("%s-%s-%s", distro, majorVer, minorVer)
}

// NewImageManager returns a new ImageManager struct
func NewImageManager(platform string) *ImageManager {
	var (
		sshCommand *exec.Cmd
		imageOS    string
	)
	switch platform {
	case "GCE":
		sshCommand = exec.Command("gcloud", "compute", "ssh")
		imageOS = os.Getenv("MONGO_INSTANCE_OS")
		break
	case "local":
		sshCommand = exec.Command("sh")
		imageOS = formatPrettyName(getLocalOS())
	}
	if imageOS == "" {
		imageOS = "ubuntu-14-04"
	}
	return &ImageManager{
		Platform:   platform,
		OS:         imageOS,
		SSHCommand: sshCommand,
	}
}

func (i *ImageManager) run(finish chan error) {
	log.Println("Running exec")
	finish <- i.SSHCommand.Run()
	//log.Println(i.SSHCommand.Output())
}

// RunCMD runs a Bash command on targeted image
func (i *ImageManager) RunCMD(command string) {
	waitgroup.Wait()
	waitgroup.Add(1)
	defer waitgroup.Done()
	if strings.Contains(i.Platform, "GCE") {
		i.SSHCommand.Args = append(i.SSHCommand.Args, "--command", fmt.Sprintf("\"%s\"", command))
	} else if strings.Contains(i.Platform, "local") {
		i.SSHCommand.Args = append(i.SSHCommand.Args, "-C", fmt.Sprintf("\"%s\"", command))
	}
	finish := make(chan error)
	go i.run(finish)
	select {
	case m1 := <-finish:
		log.Println(command, "pid:", i.SSHCommand.Process.Pid, "is about to close")
		if m1 != nil && !strings.Contains(fmt.Sprintf("%v", m1), "127") {
			log.Fatal("res", m1)
		} else {
			log.Println("res", m1)
			i.SSHCommand = exec.Command(i.SSHCommand.Path)
			return
		}
	}
}

// InstallMongo installs mongo on target image
func (i *ImageManager) InstallMongo() {
	var installCMD string
	switch i.OS {
	case "ubuntu-14-04", "ubuntu-12-04", "debian-7", "debian-8":
		installCMD = "apt-get update && apt-get install mongodb"
	case "darwin--": // clean up "darwin" OS name
		log.Println("is darwin OS")
		installCMD = "/usr/local/bin/brew update && /usr/local/bin/brew install mongodb"
	}
	i.RunCMD(installCMD)
}
