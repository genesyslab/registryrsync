package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args
	host := args[1]
	port := args[2]

	r := RegistryInfo{address{host, port}, "", "", true}
	matchesAll := matchEverything{}
	ids, err := GetMatchingImages(r, matchesAll, matchesAll)

	fmt.Printf("Got back images:%v or error:%v", ids, err)
}
