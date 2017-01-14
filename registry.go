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

// RegistryEvent indication that some image changed in some way
type RegistryEvent struct {
	// TODO create an enum
	Action string
	Target RegistryTarget
}

// RegistryTarget Indicates the precise image
type RegistryTarget struct {
	Repository string
	Tag        string
}

// A collection of targets
type RegistryTargets []RegistryTarget

// RegistryEvents a single notification may have many events
type RegistryEvents struct {
	Events []RegistryEvent
}

// Transforms to registry targets.  Useful for sorting
func (r RegistryEvents) getRegistryTargets() RegistryTargets {
	targets := make([]RegistryTarget, 0, len(r.Events))
	for _, evt := range r.Events {
		targets = append(targets, evt.Target)
	}
	return targets
}

// RegistryEventHandler how to process a registry event
type RegistryEventHandler interface {
	Handle(event RegistryEvent) error
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
