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
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/cpg1111/kubongo/daemonizer"
)

var waitgroup sync.WaitGroup

// Manager is a struct with data for OS images for instances
type Manager struct {
	Platform   string
	OS         string
	SSHCommand *exec.Cmd
	daemonizer *daemonizer.Daemonizer
}

func getLocalOS() string {
	return os.Getenv("GOOS")
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
		distro = prettyName
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

// NewManager returns a new Manager struct
func NewManager(platform string) *Manager {
	dzr := daemonizer.New("")
	return &Manager{
		Platform:   platform,
		OS:         runtime.GOOS,
		daemonizer: dzr,
	}
}

func (i *Manager) run(cmd string, finish chan error) {
	log.Println(cmd)
	err := i.daemonizer.Run(cmd)
	if err != nil {
		finish <- err
		return
	}
	finish <- nil
}

// RunCMD runs a Bash command on targeted image
func (i *Manager) RunCMD(command string) {
	log.Println("image:160 RUNCMD")
	waitgroup.Wait()
	waitgroup.Add(1)
	defer waitgroup.Done()
	log.Println("image:110 SPAWN")
	finish := make(chan error)
	go i.run(command, finish)
	for {
		select {
		case m1 := <-finish:
			log.Println("image:127 ", m1)
			//log.Println(command, "pid:", i.SSHCommand.Process.Pid, "is about to close")
			if m1 != nil {
				log.Fatal("res ", m1, command)
			} else {
				log.Println("image:132 CMD success")
				return
			}
		}
	}
}

// InstallMongo installs mongo on target image
func (i *Manager) InstallMongo() {
	mongoExec, execErr := os.Lstat("/usr/local/bin/mongod")
	log.Println("image:143 ", mongoExec)
	if execErr != nil {
		log.Println("image:145 ", execErr)
	} else if mongoExec == nil {
		var installCMD string
		switch i.OS {
		case "ubuntu-14-04", "ubuntu-12-04", "debian-7", "debian-8":
			installCMD = "apt-get update && apt-get install mongodb"
		case "darwin":
			log.Println("image:152 is darwin OS")
			installCMD = "/usr/local/bin/brew update && /usr/local/bin/brew install mongodb"
		}
		i.RunCMD(installCMD)
	}
}
