package main

import (
	"fmt"
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

		Convey("We can push an image", func() {
			err = TagAndPush("alpine", regAddr, "unstable")
			if err != nil {
				fmt.Printf("Didn't tag. Pausing for you to try yourself: %+v", err)
			}
			So(err, ShouldBeNil)
			Convey("We can get back image information from the registry", func() {
				regInfo := RegistryInfo{regAddr, "", "", true}

				matches, err := GetMatchingImages(regInfo, matchEverything{})
				So(err, ShouldBeNil)
				expectedImages := []ImageIdentifier{ImageIdentifier{"alpine", "unstable"}}
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
