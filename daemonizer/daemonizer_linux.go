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

package daemonizer

import (
	"log"
	"os"
	"os/exec"
)

// Daemonizer is a struct to have the OS' init system daemonize a commmand
type Daemonizer struct {
	toDaemon string
	Command  *exec.Cmd
	isSysv   bool
}

// New creates a new Instance of Daemonizer
func New(cmd string) *Daemonizer {
	newDaemonizer := &Daemonizer{cmd}
	hasSysv, sysvErr := os.Stat("/sbin/sysv")
	hasSystemd, systemdErr := os.Stat("/sbin/systemd")
	if (hasSysv == nil || sysvErr != nil) && (hasSystemd == nil || systemdErr != nil) {
		newDaemonizer.Command = exec.Command("initctl")
		newDaemonizer.isSysv = false
	} else if hasSystemd == nil || systemdErr != nil {
		newDaemonizer.Command = exec.Command("service")
		newDaemonizer.isSysv = true
	} else {
		newDaemonizer.Command("systemd-run")
		newDaemonizer.isSysv = false
	}
	return newDaemonizer
}

// Run runs Daemonizer's Command
func (d *Daemonizer) Run() error {
	if d.isSysv {
		d.Command.Args = append(d.Command.Args, d.toDaemon, "start")
	} else {
		d.Command.Args = append(d.Command.Args, d.toDaemon)
	}
	output, outErr := d.Command.Output()
	if outErr != nil {
		return outErr
	}
	runErr := d.Command.Run()
	log.Println("daemonizer_linx:59 ", (string)(output))
	return runErr
}
