package main

import (
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
)

//TagAndPush given a basic image name, will add it to the remote repository with the given tag
func (d dockerCli) TagAndPush(imageName string, remoteAddr string, tag string) error {
	if tag != "" {
		remoteAddr = remoteAddr + ":" + tag
	}
	err := tagImage(imageName, remoteAddr)
	if err != nil {
		log.Warnf("Error tagging %s:%s", imageName, remoteAddr)
		return err
	}

	pushCmd := exec.Command("docker", "push", remoteAddr)
	data, err := pushCmd.CombinedOutput()
	if err != nil {
		log.Printf("Error pushing %s:%s  Output %s", pushCmd.Args, err, string(data))
		return err
	}
	return nil

}

//TagAndPush given a basic image name, will add it to the remote repository with the given tag
//note that is assumed that for remote registries the url
//ends in a /  (this allows us to use the same logic to pull from docker hub, although there
//are definetly smarter ways
func (d dockerCli) PullTagPush(imageName, sourceReg, targetReg, tag string) error {

	imageParts := strings.Split(imageName, ":")

	// Should look into using gorouties and channels
	pullCmd := exec.Command("docker", "pull", sourceReg+imageName)
	data, err := pullCmd.CombinedOutput()
	if err != nil {
		log.Printf("Error pulling %s:%s  Output %s", pullCmd.Args, err, string(data))
		return err
	}

	imageId := imageParts[0]
	if tag != "" {
		imageId += ":" + tag
	}
	remoteName := targetReg + imageId
	err = tagImage(imageName, remoteName)
	if err != nil {
		log.Warnf("Error tagging %s:%s", imageId, remoteName)
		return err
	}

	pushCmd := exec.Command("docker", "push", remoteName)
	data, err = pushCmd.CombinedOutput()
	if err != nil {
		log.Printf("Error pushing %s:%s  Output %s", pushCmd.Args, err, string(data))
		return err
	}

	return nil
}

type dockerCli struct{}

type ImageHandler interface {
	PullTagPush(imageName, sourceReg, targetReg, tag string) error
}
