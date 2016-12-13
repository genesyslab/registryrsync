package main

import (
	"log"
	"net/http"
)

func ListenForNotifications(path, port string,
	handler NotificationEventHandler) {
	http.Handle(path, registryEventHandler(handler))
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
