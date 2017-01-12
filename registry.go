package main

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/heroku/docker-registry-client/registry"
	//	"github.com/docker/distribution/manifest"
)

// RegistryInfo connection information to speak with a docker registry
type RegistryInfo struct {
	address    string
	username   string
	password   string
	isInsecure bool
}

//RegistryFactory somethign that can give us pointers to registries
type RegistryFactory interface {
	GetRegistry() (Registry, error)
	Address() string
}

//Registry abstraction of a docker registry connection
type Registry interface {
	Repositories() ([]string, error)
	Tags(string) ([]string, error)
}

// GetRegistry gets an actual registry with repositories and tags
func (r RegistryInfo) GetRegistry() (Registry, error) {
	var protocol string
	if strings.Index(r.address, "localhost") == 0 {
		protocol = "http"
	} else {
		protocol = "https"
	}

	regURL := fmt.Sprintf("%s://%s", protocol, r.address)
	log.Infof("Connecting to registry %s:%s", regURL)

	reg, err := registry.New(regURL, r.username, r.password)

	if err != nil {
		//TODO should this be fatal?  maybe a warn.
		log.Errorf("Couldn't connect to registry %s:%s", regURL, err)
		return nil, err
	}
	return reg, nil

}

//GetMatchingImages Finds the names and tags of all matching
func GetMatchingImages(reg Registry, filter DockerImageFilter) ([]ImageIdentifier, error) {
	matchingImages := make([]ImageIdentifier, 0, 10)

	repos, err := reg.Repositories()
	if err != nil {
		fmt.Printf("cant get repositories from %s:%v. Got back %s", reg, err, repos)
		return matchingImages, err
	}
	for _, repo := range repos {
		if filter.repoFilter.Matches(repo) {
			tags, err := reg.Tags(repo)
			if err != nil {
				log.Fatal("Unable to get tags", err)
				return matchingImages, err
			}
			for _, tag := range tags {
				if filter.tagFilter.Matches(tag) {
					matchingImages = append(matchingImages, ImageIdentifier{repo, tag})
				}
			}
		}
	}
	return matchingImages, nil
}

//Consolidate  finds the missing images in the target from the source and fires off events for those
func Consolidate(rs, rt RegistryFactory, filter DockerImageFilter, handler RegistryEventHandler) error {
	regSource, err := rs.GetRegistry()
	if err != nil {
		log.Errorf("Couldn't connect to source registry %v:%v", rs, err)
	}
	regTarget, err := rt.GetRegistry()
	if err != nil {
		log.Errorf("Couldn't connect to source registry %v:%v", rt, err)
	}
	//This could easily take a while and we want to at the least log the time it took. In reality should probably
	//push a metric somewhere
	log.Infof(">>Consolidate(%v,%v,%v", regSource, regTarget, filter)
	defer log.Info("<<Consolidate")
	sourceImages, err := GetMatchingImages(regSource, filter)
	if err != nil {
		log.Errorf("Couldn't get images from source repo %v : %v", regSource, err)
		return err
	}
	targetImages, err := GetMatchingImages(regTarget, filter)
	if err != nil {
		log.Errorf("Couldn't get images from target repo %s : %s", regTarget, err)
		return err
	}
	missingImages := missingImages(sourceImages, targetImages)
	for _, image := range missingImages {
		handler.Handle(RegistryEvent{"missing", RegistryTarget{image.Name, image.Tag}})
	}
	return nil

}
