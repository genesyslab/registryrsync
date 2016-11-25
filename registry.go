package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
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

// StartMysql starts new docker container with MySQL running.
//
// It returns dsn, defer function and error if any.
func StartMysql() (string, func(), error) {
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

	// VM IP is the IP of DockerMachine VM, if running (used in non-Linux OSes)
	vm_ip := strings.TrimSpace(DockerMachineIP())
	// var nonLinux bool = (vm_ip != "")

	// err = client.StartContainer(c.ID, StartOptions(nonLinux))
	err = client.StartContainer(c.ID, nil)
	if err != nil {
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

	// determine IP address for MySQL
	ip := ""
	if vm_ip != "" {
		ip = vm_ip
	} else if c.NetworkSettings != nil {
		ip = strings.TrimSpace(c.NetworkSettings.IPAddress)
	}

	// wait MySQL to wake up
	if err := waitReachable(ip+":3306", 5*time.Second); err != nil {
		deferFn()
		return "", nil, err
	}

	return dsn(ip), deferFn, nil
}

// dsn returns valid dsn to be used with mysql driver for the given ip.
func dsn(ip string) string {
	return fmt.Sprintf("root:@tcp(%s:3306)/mydb", ip)
}

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
			break
		}
		if c.State.Running {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("cannot start container %s for %v", id, maxWait)
}

func CreateOptions() docker.CreateContainerOptions {
	ports := make(map[docker.Port]struct{})
	ports["3306"] = struct{}{}
	opts := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        "mariadb",
			ExposedPorts: ports,
		},
	}

	return opts
}

func DockerMachineIP() string {
	// Docker-machine is a modern solution for docker in MacOS X.
	// Try to detect it, with fallback to boot2docker
	var dockerMachine bool
	machine := os.Getenv("DOCKER_MACHINE_NAME")
	if machine != "" {
		dockerMachine = true
	}

	var buf bytes.Buffer

	var cmd *exec.Cmd
	if dockerMachine {
		cmd = exec.Command("docker-machine", "ip", machine)
	} else {
		cmd = exec.Command("boot2docker", "ip")
	}
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		// ignore error, as it's perfectly OK on Linux
		return ""
	}

	return buf.String()
}

func StartOptions(bindPorts bool) *docker.HostConfig {
	port_binds := make(map[docker.Port][]docker.PortBinding)
	if bindPorts {
		port_binds["3306"] = []docker.PortBinding{
			docker.PortBinding{HostPort: "3306"},
		}
	}
	conf := docker.HostConfig{
		PortBindings: port_binds,
	}

	return &conf
}
