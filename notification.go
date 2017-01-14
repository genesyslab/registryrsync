package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

func registryEventHandler(handler RegistryEventHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Infof("Got new request")
		if r.Body == nil {
			http.Error(w, "Please send a request body", 400)
			return
		}
		log.Debugf("Processing Notification event")

		var events RegistryEvents
		err := json.NewDecoder(r.Body).Decode(&events)
		if err != nil {
			log.Warnf("Couldn't decode")

			http.Error(w, err.Error(), 400)
			return
		}
		log.Debugf("Got back events %v", events)

		for _, event := range events.Events {
			handler.Handle(event)
		}
		fmt.Fprintf(w, "Events processed")
	}
}

// func GetRegistryEvents(reader io.Reader) (notification *RegistryNotification, err error) {
// 	err = json.NewDecoder(reader).Decode(&notification)
// 	if err != nil {
// 		log.Warnf("Couldn't parse into a notification:%s", reader)
// 	}
// 	return
// }

/**
 *{
   "events": [
      {
         "id": "320678d8-ca14-430f-8bb6-4ca139cd83f7",
         "timestamp": "2016-03-09T14:44:26.402973972-08:00",
         "action": "pull",
         "target": {
            "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
            "size": 708,
            "digest": "sha256:fea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf",
            "length": 708,
            "repository": "hello-world",
            "url": "http://192.168.100.227:5000/v2/hello-world/manifests/sha256:fea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf",
            "tag": "latest"
         },
         "request": {
            "id": "6df24a34-0959-4923-81ca-14f09767db19",
            "addr": "192.168.64.11:42961",
            "host": "192.168.100.227:5000",
            "method": "GET",
            "useragent": "curl/7.38.0"
         },
         "actor": {},
         "source": {
            "addr": "xtal.local:5000",
            "instanceID": "a53db899-3b4b-4a62-a067-8dd013beaca4"
         }
      }
   ]
}
*/
