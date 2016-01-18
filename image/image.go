package image

import (
	"fmt"
	"os"
	"os/exec"
)

type ImageManager struct {
	Platform   string
	OS         string
	SSHCommand string
}

func NewImageManager(platform string) *ImageManager {
	var sshCommand string
	switch platform {
	case "gcloud":
		sshCommand := exec.Command("gcloud", "compute", "ssh")
	}
	imageOS := os.Getenv("MONGO_INSTANCE_OS")
	if imageOS == "" {
		imageOS = "ubuntu-14-04"
	}
	return &ImageManager{
		Platform:   platform,
		OS:         imageOS,
		SSHCommand: sshCommand,
	}
}

func run(finish chan error) {
	finish <- i.SSHCommand.Run()
}

func (i *ImageManager) RunCMD(command string) {
	i.SSHCommand.Args = append(i.SSHCommand.Args, "--command", fmt.Sprintf("\"%s\"", command))
	finish := make(chan error)
	go run(finish)
	for {
		select {
		case m1 := <-finish:
			if m1 != nil {
				log.Fatal(m1)
			} else {
				i.SSHCommand = i.SSHCommand[0:len(i.SSHCommand)]
			}
		}
	}
}

func (i *ImageManager) InstallMongo() {
	var installCMD string
	switch i.OS {
	case "ubuntu-14-04", "ubuntu-12-04", "debian-7", "debian-8":
		installCMD = "apt-get update && apt-get install mongodb"
	}
	i.RunCMD(installCMD)
}
