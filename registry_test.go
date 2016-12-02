package main

import (
	"os"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)
	retCode := m.Run()
	os.Exit(retCode)
}

func TestFilteringOfARegistry(t *testing.T) {

	Convey("Given a simple registry", t, func() {

		regAddr, closer, err := StartRegistry()
		So(err, ShouldBeNil)
		So(closer, ShouldNotBeNil)
		defer closer()

		//Even though we only get here when the registry is listening on port 5000
		//it still would fail frequently but not always on a push.
		//waiting a little seems to help
		time.Sleep(1 * time.Second)
		regInfo := RegistryInfo{regAddr, "", "", true}

		Convey("We can push images", func() {
			err = TagAndPush("alpine", regAddr.remoteName("alpine"), "stable")
			So(err, ShouldBeNil)
			err = TagAndPush("alpine", regAddr.remoteName("mynamespace/alpine"), "0.1")
			So(err, ShouldBeNil)
			err = TagAndPush("alpine", regAddr.remoteName("alpine"), "0.1")
			So(err, ShouldBeNil)
			err = TagAndPush("busybox", regAddr.remoteName("mynamespace/busybox"), "0.1-stable")
			So(err, ShouldBeNil)

			Convey("We can get back image information from the registry", func() {

				matchesAll := matchEverything{}
				tagFilter, err := NewRegexTagFilter(".*stable")
				namespaceFilter := NewNamespaceFilter("mynamespace")
				So(err, ShouldBeNil)
				Convey("We can get back all the images", func() {
					matches, err := GetMatchingImages(regInfo, matchesAll, matchesAll)
					So(err, ShouldBeNil)
					expectedImages := []ImageIdentifier{
						ImageIdentifier{"alpine", "0.1"},
						ImageIdentifier{"alpine", "stable"},
						ImageIdentifier{"mynamespace/alpine", "0.1"},
						ImageIdentifier{"mynamespace/busybox", "0.1-stable"},
					}
					So(matches, ShouldResemble, expectedImages)
				})
				Convey("We can get only images that are marked stable", func() {
					matches, err := GetMatchingImages(regInfo, matchesAll, tagFilter)
					So(err, ShouldBeNil)
					expectedImages := []ImageIdentifier{
						ImageIdentifier{"alpine", "stable"},
						ImageIdentifier{"mynamespace/busybox", "0.1-stable"},
					}
					So(matches, ShouldResemble, expectedImages)
				})
				Convey("we can get all the images in a given namespace", func() {
					matches, err := GetMatchingImages(regInfo, namespaceFilter, matchesAll)
					So(err, ShouldBeNil)
					expectedImages := []ImageIdentifier{
						ImageIdentifier{"mynamespace/alpine", "0.1"},
						ImageIdentifier{"mynamespace/busybox", "0.1-stable"},
					}
					So(matches, ShouldResemble, expectedImages)
				})
				Convey("we can get only images in a given namespace and given tags", func() {
					matches, err := GetMatchingImages(regInfo, namespaceFilter, tagFilter)
					So(err, ShouldBeNil)
					expectedImages := []ImageIdentifier{
						ImageIdentifier{"mynamespace/busybox", "0.1-stable"},
					}
					So(matches, ShouldResemble, expectedImages)
				})
			})
		})
	})
}
