package image

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type ImageManager struct {
	Platform   string
	OS         string
	SSHCommand string
}

func getLocalOS() string, error {
	osNameCMD := exec.Command("cat", "/etc/*-release", "|", "grep", "PRETTY_NAME")
	var osNameOut bytes.Buffer
	osNameCMD.STDout = &osNameOut
	err := osNameCMD.Run()
	return string(osNameOut), err
}

func formatPrettyName(prettyName string) string {
	var distro, majorVer, minorVer string
	if strings.Contains(prettyName, "Ubuntu") {
		distro = "ubuntu"
	} else if strings.Contains(prettyName, "Debian"){
		distro = "debian"
	} else if strings.Contains(prettyName, "CentOs") {
		distro = "centos"
	} else if string.Contains(prettyName, "CoreOs") {
		distro = "coreos"
	} else {
		log.Fatal("Your OS is not currently supported, to change this, please create an issue @ https://github.com/cpg1111/kubongo/issues")
	}
	rx, rxErr := regexp.MustCompile("([0-9]+)")
	if rxErr != nil {
		panic(rxErr)
	}
	matches := rx.FindStringSubmatch(prettyName)
	majorVer = matches[0]
	minorVer = matches[1]
	return fmt.Sprintf("%s-%s-%s", distro, majorVer, minorVer)
}

func NewImageManager(platform string) *ImageManager {
	var sshCommand, imageOS string
	switch platform {
	case "gcloud":
		sshCommand := exec.Command("gcloud", "compute", "ssh")
		imageOS = os.Getenv("MONGO_INSTANCE_OS")
		break
	case "local":
		sshCommand := exec.Command("sh")
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
	finish <- i.SSHCommand.Run()
}

func (i *ImageManager) RunCMD(command string) {
	if strings.Compare(i.Platform, "GCE") == 0 {
		i.SSHCommand.Args = append(i.SSHCommand.Args, "--command", fmt.Sprintf("\"%s\"", command))
	} else if strings.Compare(i.Platform, "local") == 0 {
		i.SSHCommand.Args = append(i.SSHCommand.Args, "-C", fmt.Sprintf("\"%s\"", command))
	}
	finish := make(chan error)
	go i.run(finish)
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
