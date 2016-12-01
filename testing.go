package main

import (
	"fmt"
	"log"
	"net"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

type address struct {
	HostIP string
	Port   string
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
			log.Println("Container not started %s %s", c, err)
			break
		}
		if c.State.Running {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("cannot start container %s for %v", id, maxWait)
}

//StartRegistry - starts a new registry, returning the ip port combination
//and a closing function that you should use for cleanup.  Error of course if there is a problem

func StartRegistry() (address, func(), error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return address{}, nil, err
	}
	c, err := client.CreateContainer(createOptions())

	log.Printf("Container created %+v", c)
	if err != nil {
		log.Fatal("Couldn't create even a basic container:", err)
		return address{}, nil, err
	}
	deferFn := func() {
		if err := client.RemoveContainer(docker.RemoveContainerOptions{
			ID:    c.ID,
			Force: true,
		}); err != nil {
			log.Println("cannot remove container: %s", err)
		}
	}

	err = client.StartContainer(c.ID, nil)
	if err != nil {
		log.Fatalf("Could not start container %s : ", c, err)
		deferFn()
		return address{}, nil, err
	}

	// wait for container to wake up
	if err := waitStarted(client, c.ID, 5*time.Second); err != nil {
		deferFn()
		return address{}, nil, err
	}
	if c, err = client.InspectContainer(c.ID); err != nil {
		deferFn()
		return address{}, nil, err
	}

	ports := c.NetworkSettings.Ports
	//We know that the registry listens on port 5000 generally
	//but we don't know what it will actually be mapped to when testing
	port := docker.Port("5000/tcp")
	if portInfo, ok := ports[port]; ok && len(portInfo) >= 1 {
		hostIp := portInfo[0].HostIP
		hostPort := portInfo[0].HostPort
		portAddress := hostIp + ":" + hostPort

		if err := waitReachable(portAddress, 5*time.Second); err != nil {
			deferFn()
			return address{}, nil, err
		}

		//localhost is by default I think marked as
		//an insecure registry, and since this is just for tests ..
		if hostIp == "0.0.0.0" {
			hostIp = "localhost"
		}
		return address{hostIp, hostPort}, deferFn, nil
	} else {
		//Close it
		deferFn()

		return address{}, nil, fmt.Errorf("Coudln't find port %v in ports %v settings %v", port, ports, c.NetworkSettings)
	}
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
