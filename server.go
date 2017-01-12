package main

import (
	"time"
)

// ServerRequest  A server will listen on the given port
type ServerRequest struct {
	source, target                    RegistryInfo
	namespaceFilter, repositoryFilter string
	port                              int
	frequency                         time.Duration
	resourcePath                      string
}

type server struct {
	source, target *Registry
	filter         *DockerImageFilter
	port           int
}

func (request ServerRequest) newServer() server {
	return server{}
}

func (s server) start() {

}
