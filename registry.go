package main

import (
	"fmt"
	"log"
	"time"

	"github.com/fsouza/go-dockerclient"
)

func main() {
	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}
	imgs, err := client.ListImages(docker.ListImagesOptions{All: false})
	if err != nil {
		panic(err)
	}
	for _, img := range imgs {
		fmt.Println("ID: ", img.ID)
		fmt.Println("RepoTags: ", img.RepoTags)
		fmt.Println("Created: ", img.Created)
		fmt.Println("Size: ", img.Size)
		fmt.Println("VirtualSize: ", img.VirtualSize)
		fmt.Println("ParentId: ", img.ParentID)
	}
}

func StartRegistry() (string, func(), error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return "", nil, err
	}
	c, err := client.CreateContainer(CreateOptions())

	log.Printf("Container created %s", c)
	if err != nil {
		log.Fatal("Couldn't create even a basic container:", err)
		return "", nil, err
	}
	deferFn := func() {
		if err := client.RemoveContainer(docker.RemoveContainerOptions{
			ID:    c.ID,
			Force: true,
		}); err != nil {
			log.Println("cannot remove container: %v", err)
		}
	}

	err = client.StartContainer(c.ID, nil)
	if err != nil {
		log.Fatalf("Could not start container %s : ", c, err)
		deferFn()
		return "", nil, err
	}

	log.Printf("Container started")

	// wait for container to wake up
	if err := waitStarted(client, c.ID, 5*time.Second); err != nil {
		deferFn()
		return "", nil, err
	}
	if c, err = client.InspectContainer(c.ID); err != nil {
		deferFn()
		return "", nil, err
	}

	ports := c.NetworkSettings.Ports
	port := docker.Port("5000/tcp")
	if portInfo, ok := ports[port]; ok && len(portInfo) >= 1 {
		hostIp := portInfo[0].HostIP
		hostPort := portInfo[0].HostPort
		portAddress := hostIp + ":" + hostPort

		if err := waitReachable(portAddress, 5*time.Second); err != nil {
			deferFn()
			return "", nil, err
		}
		return portAddress, deferFn, nil
	} else {
		//Close it
		deferFn()

		return "", nil, fmt.Errorf("Coudln't find port %v in ports %v settings %v", port, ports, c.NetworkSettings)
	}
}

// dsn returns valid dsn to be used with mysql driver for the given ip.
func dsn(ip string) string {
	return fmt.Sprintf("root:@tcp(%s:3306)/mydb", ip)
}

func CreateOptions() docker.CreateContainerOptions {
	ports := make(map[docker.Port]struct{})

	port := docker.Port("5000/tcp")
	ports[port] = struct{}{}
	// ports["5000"] = struct{}{}
	opts := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        "registry:2.4",
			ExposedPorts: ports,
		},
	}

	return opts
}
