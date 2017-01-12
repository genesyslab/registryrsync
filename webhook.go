package main

import (
	"log"
	"net/http"
)

//ListenForNotifications starts an http server
func ListenForNotifications(path, port string,
	handler RegistryEventHandler) {
	http.Handle(path, registryEventHandler(handler))
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
