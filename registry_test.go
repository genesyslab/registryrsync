package main

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

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
			err = TagAndPush("alpine", regAddr.remoteName("alpine"), "unstable")
			So(err, ShouldBeNil)
			err = TagAndPush("alpine", regAddr.remoteName("mynamespace/alpine"), "0.1")
			So(err, ShouldBeNil)
			err = TagAndPush("alpine", regAddr.remoteName("alpine"), "stable")
			So(err, ShouldBeNil)
			err = TagAndPush("busybox", regAddr.remoteName("mynamespace/busybox"), "stable")
			So(err, ShouldBeNil)

			Convey("We can get back image information from the registry", func() {
				matches, err := GetMatchingImages(regInfo, matchEverything{})
				So(err, ShouldBeNil)
				expectedImages := []ImageIdentifier{
					ImageIdentifier{"alpine", "stable"},
					ImageIdentifier{"alpine", "unstable"},
					ImageIdentifier{"mynamespace/alpine", "0.1"},
					ImageIdentifier{"mynamespace/busybox", "stable"},
				}
				So(matches, ShouldResemble, expectedImages)
			})
		})

		// Convey("With a simple filtering rule", func() {

		// 	Convey("That has three images", nil)

		// 	Convey("or Three Namespaces", nil)

		// })

		// Convey("or  Three Tags", nil)

	})

}
