package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/heroku/docker-registry-client/registry"
	//	"github.com/docker/distribution/manifest"

	"github.com/golang/glog"
)

type RegistryInfo struct {
	address    address
	username   string
	password   string
	isInsecure bool
}
type address struct {
	HostIP string
	Port   string
}

func (a address) remoteName(imageName string) string {
	return fmt.Sprintf("%s:%s/%s", a.HostIP, a.Port, imageName)
}

type RegistryFactory interface {
	GetRegistry() (*registry.Registry, error)
}

func (r RegistryInfo) GetRegistry() (*registry.Registry, error) {

	var protocol string
	if r.isInsecure {
		protocol = "http"
	} else {
		protocol = "https"
	}

	regUrl := fmt.Sprintf("%s://%s:%s", protocol, r.address.HostIP, r.address.Port)
	reg, err := registry.New(regUrl, r.username, r.password)

	if err != nil {
		//TODO should this be fatal?  maybe a warn.
		glog.Warningf("Couldn't connect to registry %s:%s", regUrl, err)
		return nil, err
	}
	return reg, nil

}

type ImageIdentifier struct {
	Name string
	Tag  string
}

type ImageFilter interface {
	RepositoryFilter
	TagFilter
}

type RepositoryFilter interface {
	MatchesRepo(repo string) bool
}

type TagFilter interface {
	MatchesTag(tag string) bool
}

type matchEverything struct{}

func (m matchEverything) MatchesRepo(name string) bool {
	return true
}

func (m matchEverything) MatchesTag(name string) bool {
	return true
}

func GetMatchingImages(regFactory RegistryFactory, filter ImageFilter) ([]ImageIdentifier, error) {

	matchingImages := make([]ImageIdentifier, 0, 10)

	reg, err := regFactory.GetRegistry()
	if err != nil {
		return matchingImages, err
	}
	repos, err := reg.Repositories()
	if err != nil {
		fmt.Printf("cant get repositories from %s:%v. Got back %s", regFactory, err, repos)
		return matchingImages, err
	}
	for _, repo := range repos {

		if filter.MatchesRepo(repo) {
			tags, err := reg.Tags(repo)
			if err != nil {
				log.Fatal("Unable to get tags", err)
				return matchingImages, err
			}
			for _, tag := range tags {
				if filter.MatchesTag(tag) {
					matchingImages = append(matchingImages, ImageIdentifier{repo, tag})
				}
			}
		}
	}
	return matchingImages, nil
}

func TagImage(image string, taggedName string) error {
	tagCmd := exec.Command("docker", "tag", image, taggedName)
	data, err := tagCmd.CombinedOutput()
	if err != nil {
		log.Printf("Error tagging %s:%s  Output %s", tagCmd.Args, err, string(data))
		return err
	}
	return nil
}

//  "github.com/docker/distribution/digest"
//     "github.com/docker/distribution/manifest"
//     "github.com/docker/libtrust"

func TagAndPush(imageName string, remoteAddr string, tag string) error {
	if tag != "" {
		remoteAddr = remoteAddr + ":" + tag
	}
	err := TagImage(imageName, remoteAddr)
	if err != nil {
		log.Printf("Error tagging %s:%s", imageName, remoteAddr)
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
