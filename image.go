package main

import (
	"log"
	"os/exec"
)

type ImageIdentifier struct {
	Name string
	Tag  string
}

// Finds all the images that aren't in the target but are in the source
func missingImages(source, target []ImageIdentifier) []ImageIdentifier {
	diffs := make([]ImageIdentifier, 0, len(source))
	vals := make(map[ImageIdentifier]int)
	for _, t := range target {
		vals[t] = 1
	}
	for _, s := range source {
		if _, ok := vals[s]; !ok {
			diffs = append(diffs, s)
		}
	}
	return diffs
}

func tagImage(image string, taggedName string) error {
	tagCmd := exec.Command("docker", "tag", image, taggedName)
	data, err := tagCmd.CombinedOutput()
	if err != nil {
		log.Printf("Error tagging %s:%s  Output %s", tagCmd.Args, err, string(data))
		return err
	}
	return nil
}
