package main

import (
	"flag"
	"os"
	"sort"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

// type matchEverything struct{}
//
// func (m matchEverything) Matches(name string) bool {
// 	return true
// }

func TestMain(m *testing.M) {
	flag.Parse()
	log.SetLevel(log.DebugLevel)
	retCode := m.Run()
	os.Exit(retCode)
}

func TestFilteringOfARegistry(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	Convey("Given a simple but real registry", t, func() {

		hostIP, port, closer, err := startRegistry()
		So(err, ShouldBeNil)
		So(closer, ShouldNotBeNil)
		// defer closer()

		regInfo := RegistryInfo{address: hostIP + ":" + port}
		//Even though we only get here when the registry is listening on port 5000
		//it still would fail frequently but not always on a push.
		//waiting a little seems to help
		time.Sleep(1 * time.Second)
		registry, err := regInfo.GetRegistry()
		So(err, ShouldBeNil)

		Convey("We can push images and retrieve", func() {
			allimageFilter := DockerImageFilter{matchEverything{}, matchEverything{}}
			imageHandler, err := NewDockerCLIHandler(DockerHubRegistry, regInfo, allimageFilter)
			So(err, ShouldBeNil)
			err = imageHandler.PullTagPush("alpine", "latest")
			So(err, ShouldBeNil)
			err = imageHandler.tagger.Tag("alpine", "mynamespace/alpine:0.1")
			So(err, ShouldBeNil)
			err = imageHandler.target.Push("mynamespace/alpine:0.1")
			So(err, ShouldBeNil)
			log.Debug("Pushed namespaced alpine")
			err = imageHandler.tagger.Tag("alpine", "alpine:0.1")
			So(err, ShouldBeNil)
			err = imageHandler.target.Push("alpine:0.1")
			So(err, ShouldBeNil)
			err = imageHandler.source.Pull("busybox")
			So(err, ShouldBeNil)
			err = imageHandler.tagger.Tag("busybox", "mynamespace/busybox:0.1-stable")
			So(err, ShouldBeNil)
			err = imageHandler.target.Push("mynamespace/busybox:0.1-stable")
			So(err, ShouldBeNil)
			matches, err := GetMatchingImages(registry, allimageFilter)
			So(err, ShouldBeNil)
			// var expecteImages RegistryTargets
			expectedImages := RegistryTargets{
				RegistryTarget{"alpine", "0.1"},
				RegistryTarget{"alpine", "latest"},
				RegistryTarget{"mynamespace/alpine", "0.1"},
				RegistryTarget{"mynamespace/busybox", "0.1-stable"},
			}
			sort.Sort(matches)
			sort.Sort(expectedImages)
			So(matches, ShouldResemble, expectedImages)
			Convey("We can push out to another registry the requested differences", func() {
				hostIP, port, closerReg2, err := startRegistry()
				//Even though we only get here when the registry is listening on port 5000
				//it still would fail frequently but not always on a push.
				//waiting a little seems to help
				time.Sleep(1 * time.Second)
				So(err, ShouldBeNil)
				So(closer, ShouldNotBeNil)
				defer closerReg2()
				regInfo2 := RegistryInfo{address: hostIP + ":" + port}
				So(err, ShouldBeNil)
				imageHandler, err := NewDockerCLIHandler(regInfo, regInfo2, allimageFilter)
				reg1, err := regInfo.GetRegistry()
				So(err, ShouldBeNil)
				reg2, err := regInfo2.GetRegistry()
				So(err, ShouldBeNil)
				err = Consolidate(reg1, reg2,
					DockerImageFilter{NewNamespaceFilter("mynamespace"), matchEverything{}}, imageHandler)
				So(err, ShouldBeNil)
				matches, err = GetMatchingImages(registry, allimageFilter)
				So(err, ShouldBeNil)
				expectedImages = RegistryTargets{
					RegistryTarget{"mynamespace/alpine", "0.1"},
					RegistryTarget{"mynamespace/busybox", "0.1-stable"},
				}
				sort.Sort(matches)
				sort.Sort(expectedImages)
			})
		})
	})
}
