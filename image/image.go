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

// ImageManager is a struct with data for OS images for instances
type ImageManager struct {
	Platform   string
	OS         string
	SSHCommand *exec.Cmd
}

func getLocalOS() string {
	osNameCMD := exec.Command("cat", "/etc/*-release", "|", "grep", "PRETTY_NAME")
	var osNameOut bytes.Buffer
	osNameCMD.Stdout = &osNameOut
	err := osNameCMD.Run()
	if err != nil {
		log.Fatal(err)
	}
	return osNameOut.String()
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
		log.Fatal("Your OS is not currently supported, to change this, please create an issue @ https://github.com/cpg1111/kubongo/issues")
	}
	rx := regexp.MustCompile("([0-9]+)")
	matches := rx.FindStringSubmatch(prettyName)
	majorVer = matches[0]
	minorVer = matches[1]
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
	finish <- i.SSHCommand.Run()
}

// RunCMD runs a Bash command on targeted image
func (i *ImageManager) RunCMD(command string) {
	originalLen := len(i.SSHCommand.Args)
	if strings.Contains(i.Platform, "GCE") {
		i.SSHCommand.Args = append(i.SSHCommand.Args, "--command", fmt.Sprintf("\"%s\"", command))
	} else if strings.Contains(i.Platform, "local") {
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
				i.SSHCommand.Args = i.SSHCommand.Args[0:originalLen]
			}
		}
	}
}

// InstallMongo installs mongo on target image
func (i *ImageManager) InstallMongo() {
	var installCMD string
	switch i.OS {
	case "ubuntu-14-04", "ubuntu-12-04", "debian-7", "debian-8":
		installCMD = "apt-get update && apt-get install mongodb"
	}
	i.RunCMD(installCMD)
}
