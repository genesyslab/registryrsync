package main

import (
	"fmt"
	"os"
	"os/exec"

	"log"

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

func TagAndPush(imageName string, addr address, tag string) error {
	remoteAddr := fmt.Sprintf("%s:%s/%s", addr.HostIP, addr.Port, imageName)

	if tag != "" {
		remoteAddr = remoteAddr + ":" + tag
	}
	tagCmd := exec.Command("docker", "tag", imageName, remoteAddr)
	tagCmd.Stdout = os.Stdout
	err := tagCmd.Run()
	if err != nil {
		log.Printf("Error tagging %s", tagCmd.Args)
		return err
	}

	pushCmd := exec.Command("docker", "push", remoteAddr)
	pushCmd.Stdout = os.Stdout
	err = pushCmd.Run()
	if err != nil {
		log.Printf("Error pushging %s %s", pushCmd.Args, err)
		return err
	}
	return nil

}
