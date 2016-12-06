package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type RecorderHandler struct {
	events []RegistryEvent
}

func (r *RecorderHandler) Handle(event RegistryEvent) error {
	r.events = append(r.events, event)
	return nil
}

func Test_registryEventHandler(t *testing.T) {
	type args struct {
		requestBody string
		handler     *RecorderHandler
	}
	tests := []struct {
		name string
		args args
		want []RegistryEvent
	}{{
		"Full event notification",
		args{`{
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
	}`,
			&RecorderHandler{}},
		[]RegistryEvent{RegistryEvent{"pull", RegistryTarget{"helllo-world", "latest"}}},
	}}
	Convey("We can parse out", t, func() {
		for _, tt := range tests {
			Convey(tt.name, func() {
				t.Run(tt.name, func(t *testing.T) {
					eventHandler := registryEventHandler(tt.args.handler)
					req, _ := http.NewRequest("POST", "/", strings.NewReader(tt.args.requestBody))
					w := httptest.NewRecorder()
					eventHandler.ServeHTTP(w, req)
					//TODO this isn't really data driven. i.e. I can't return non 200 with
					// this approach
					if w.Code != http.StatusOK {
						t.Errorf("Home page didn't return %v", http.StatusOK)
					}
					ShouldResemble(tt.args.handler.events, tt.want)
					// if got := tt.args.handler.events; !reflect.DeepEqual(got, tt.want) {
					// 	t.Errorf("didn't handle events() = %v, want %v", got, tt.want)
					// }
				})
			})
		}
	})
}
