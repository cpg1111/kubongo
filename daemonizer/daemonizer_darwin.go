package daemonizer

import (
	"errors"
	"log"
	"os/exec"
)

// Daemonizer is a struct that uses launchd to daemonize the specified command
type Daemonizer struct {
	toDaemon string
	Command  *exec.Cmd
}

// New returns a new instance of Daemonizer
func New(cmd string) *Daemonizer {
	return &Daemonizer{
		toDaemon: cmd,
		Command:  exec.Command("launchctl", "start"),
	}
}

// Run runs the launchctl command to daemonize the specified command
func (d *Daemonizer) Run(cmd string) error {
	log.Println("daemonizer_darwin:24 run")
	if cmd != "" {
		d.toDaemon = cmd
	}
	if d.toDaemon == "" {
		errors.New("no command was given")
	}
	output, outErr := d.Command.Output()
	if outErr != nil {
		log.Println("OUT ERROR MOTHERFUCKER!!!")
		return outErr
	}
	runErr := d.Command.Run()
	log.Println("daemonizer_darwin:30 ", (string)(output))
	return runErr
}
