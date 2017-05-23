package main

import (
	"fmt"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	docker "github.com/fsouza/go-dockerclient"
)

//startRegistry - starts a new registry, returning the ip port combination
//and a closing function that you should use for cleanup.
//Error of course if there is a problem
func startRegistry() (hostIP, port string, deferFn func(), err error) {
	client, err := docker.NewClientFromEnv()
	c, err := client.CreateContainer(createOptions())
	log.Debugf("Container created %+v", c)
	if err != nil {
		log.Fatal("Couldn't create even a basic container:", err)
		return
	}
	deferFn = func() {
		if err = client.RemoveContainer(docker.RemoveContainerOptions{
			ID:    c.ID,
			Force: true,
		}); err != nil {
			log.Warnf("cannot remove container: %s", err)
		}
	}
	err = client.StartContainer(c.ID, nil)
	if err != nil {
		log.Fatalf("Could not start container %v : %s", c, err)
		deferFn()
		return
	}
	// wait for container to wake up
	if err = waitStarted(client, c.ID, 5*time.Second); err != nil {
		deferFn()
		return
	}
	if c, err = client.InspectContainer(c.ID); err != nil {
		deferFn()
		return
	}
	ports := c.NetworkSettings.Ports
	//We know that the registry listens on port 5000 generally
	//but we don't know what it will actually be mapped to when testing
	p := docker.Port("5000/tcp")
	if portInfo, ok := ports[p]; ok && len(portInfo) >= 1 {
		hostIP = portInfo[0].HostIP
		port = portInfo[0].HostPort
		//localhost is by default I think marked as
		//an insecure registry, and since this is just for tests ..
		if hostIP == "0.0.0.0" {
			hostIP = "localhost"
		}
		if err = waitReachable(hostIP+":"+port, 5*time.Second); err != nil {
			deferFn()
			return
		}
		return
	}
	deferFn()
	err = fmt.Errorf("Coudln't find port %v in ports %v settings %v", port, ports, c.NetworkSettings)
	return
}

// In the case of a single exposed port on a client will
// return the appropraite connectivity information.  This is
// needed becasue at least on OSX you CANT use the ips of the containers
// as of this writing, and docker-machine/boot2docker are no longer
// the official versions

// waitReachable waits for hostport to became reachable for the maxWait time.
func waitReachable(hostport string, maxWait time.Duration) error {
	done := time.Now().Add(maxWait)
	for time.Now().Before(done) {
		c, err := net.Dial("tcp", hostport)
		if err == nil {
			c.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("cannot connect %v for %v", hostport, maxWait)
}

// waitStarted waits for container to start for the maxWait time.
func waitStarted(client *docker.Client, id string, maxWait time.Duration) error {
	done := time.Now().Add(maxWait)
	for time.Now().Before(done) {
		c, err := client.InspectContainer(id)
		if err != nil {
			//This is to be expected so probably will remove log message later
			log.Warnf("Container not started %s %s", c, err)
			break
		}
		if c.State.Running {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("cannot start container %s for %v", id, maxWait)
}

func createOptions() docker.CreateContainerOptions {
	ports := make(map[docker.Port]struct{})
	hostConfig := &docker.HostConfig{
		PublishAllPorts: true,
	}
	opts := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        "registry:2.4",
			PortSpecs:    []string{"5000"},
			ExposedPorts: ports,
		},
		HostConfig: hostConfig,
	}

	return opts
}
