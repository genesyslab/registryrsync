package main

import (
	"fmt"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// DockerHubRegistry - empty registry that we can pull from
var DockerHubRegistry = RegistryInfo{}

// NewDockerCLIHandler creates something that can pull, tag and push to docker
// registries.  Note that if there is no address specified in the
// source it is treated at the docker hub registry
func NewDockerCLIHandler(source, target RegistryInfo, filter DockerImageFilter) (handler ImageHandler, err error) {
	s := dockerRegistryCLI{source}
	t := dockerRegistryCLI{target}
	err = s.login()
	if err != nil {
		return
	}
	err = t.login()
	if err != nil {
		return
	}
	// handler.source = source{&s, s.reg}
	handler.source = regSource{&s, s.reg}
	handler.tagger = &t
	handler.target = regTarget{&t, t.reg}
	handler.filter = filter
	return
}

type dockerRegistryCLI struct {
	reg RegistryInfo
}

func (d *dockerRegistryCLI) login() error {
	if d.reg.password != "" {
		loginCmd := exec.Command("docker", "login", d.reg.Address(),
			"-u", d.reg.username, "-p", d.reg.password)
		out, err := loginCmd.CombinedOutput()
		if err != nil {
			log.Warnf("Error logging in to %s with username %s :%s.  Output:\n%s",
				d.reg.address, d.reg.username, err, out)
			return err
		}
	} else {
		log.Infof("No credentials provided for %s ", d.reg.address)
	}
	return nil
}

type dockerCli struct {
	regInfo RegistryInfo
}

func (d *dockerRegistryCLI) Push(name string) error {

	log.Debugf(">>Push (%s) to %s", name, d.reg.address)
	defer log.Debug("<<Pull")
	targetAddr := d.reg.address
	var remoteName string
	if strings.Index(name, targetAddr) != 0 {
		remoteName = fmt.Sprintf("%s/%s", targetAddr, name)
		log.Debugf("No remote address added.  Tagging to add %s", remoteName)
		d.Tag(name, remoteName)
	} else {
		remoteName = name
	}
	pushCmd := exec.Command("docker", "push", remoteName)
	data, err := pushCmd.CombinedOutput()
	if err != nil {
		log.Warnf("Error pushing %s:%s  Output %s", pushCmd.Args, err, string(data))
		return err
	}
	return nil
}

func (d *dockerRegistryCLI) Pull(name string) error {
	log.Debugf(">>Pull (%s)", name)
	defer log.Debug("<<Pull")

	sourceAddr := d.reg.address
	var remoteName string
	if (sourceAddr != "") && (strings.Index(name, sourceAddr) != 0) {
		remoteName = fmt.Sprintf("%s/%s", sourceAddr, name)
	} else {
		remoteName = name
	}
	pullCmd := exec.Command("docker", "pull", remoteName)
	data, err := pullCmd.CombinedOutput()
	if err != nil {
		log.Warnf("Error pull %s:%s  Output %s", pullCmd.Args, err, string(data))
		return err
	}
	log.Debugf(">>Pull (%s)", name)

	return nil
}

func (d *dockerRegistryCLI) Tag(name, tag string) error {
	log.Debugf(">>Tag (%s,%s)", name, tag)
	defer log.Debug("<<Tag")
	tagCmd := exec.Command("docker", "tag", name, tag)
	data, err := tagCmd.CombinedOutput()
	if err != nil {
		log.Printf("Error tagging %s:%s  Output %s", tagCmd.Args, err, string(data))
		return err
	}
	return nil
}
