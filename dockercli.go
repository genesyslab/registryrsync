package main

import (
	"fmt"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func tagImage(image string, taggedName string) error {
	tagCmd := exec.Command("docker", "tag", image, taggedName)
	data, err := tagCmd.CombinedOutput()
	if err != nil {
		log.Printf("Error tagging %s:%s  Output %s", tagCmd.Args, err, string(data))
		return err
	}
	return nil
}

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

	var sourceImageAddr string
	if sourceReg == "" {
		sourceImageAddr = imageName
	} else {
		sourceImageAddr = fmt.Sprintf("%s/%s", sourceReg, imageName)
	}

	log.Debugf("Pulling %s", sourceImageAddr)

	// Should look into using gorouties and channels
	pullCmd := exec.Command("docker", "pull", sourceImageAddr)
	data, err := pullCmd.CombinedOutput()
	if err != nil {
		log.Printf("Error pulling %s:%s  Output %s", pullCmd.Args, err, string(data))
		return err
	}

	imageID := imageParts[0]
	if tag != "" {
		imageID += ":" + tag
	}
	remoteName := fmt.Sprintf("%s/%s", targetReg, imageID)
	err = tagImage(imageName, remoteName)
	if err != nil {
		log.Warnf("Error tagging %s - %s", remoteName, err)
		return err
	}

	log.Debugf("Pushing %s", remoteName)

	pushCmd := exec.Command("docker", "push", remoteName)
	data, err = pushCmd.CombinedOutput()
	if err != nil {
		log.Printf("Error pushing %s:%s  Output %s", pushCmd.Args, err, string(data))
		return err
	}

	return nil
}

type dockerCli struct{}
