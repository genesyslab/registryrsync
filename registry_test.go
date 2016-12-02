package main

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type matchEverything struct{}

func (m matchEverything) MatchesRepo(name string) bool {
	return true
}

func (m matchEverything) MatchesTag(name string) bool {
	return true
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

		Convey("We can push an image", func() {

			err = TagAndPush("alpine", regAddr, "unstable")
			if err != nil {
				fmt.Printf("Didn't tag. Pausing for you to try yourself: %+v", err)
				time.Sleep(120 * time.Second)
			}
			So(err, ShouldBeNil)
			Convey("We can get back image information from the registry", func() {
				time.Sleep(3 * time.Second)
				regInfo := RegistryInfo{regAddr, "", "", true}

				matches, err := GetMatchingImages(regInfo, matchEverything{})
				if err != nil {
					fmt.Printf("Coudldn't get repositories.  pausing to allow examination")
					time.Sleep(10 * time.Second)
					matches, err = GetMatchingImages(regInfo, matchEverything{})
					if err != nil {
						fmt.Printf("still coudn't't get repositories: %v", err)
					}

				}
				So(err, ShouldBeNil)
				expectedImages := []ImageIdentifier{ImageIdentifier{"alipine", "unstable"}}
				So(matches, ShouldEqual, expectedImages)
			})
		})

		// Convey("With a simple filtering rule", func() {

		// 	Convey("That has three images", nil)

		// 	Convey("or Three Namespaces", nil)

		// })

		// Convey("or  Three Tags", nil)

	})

}
