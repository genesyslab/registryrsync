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

		Convey("We can push an image", func() {

			err = TagAndPush("alpine", regAddr, "joshtest")
			if err != nil {
				fmt.Printf("Didn't tag. Pausing for you to try yourself: %+v", err)
				time.Sleep(120 * time.Second)
			}
			So(err, ShouldBeNil)
		})

		// Convey("With a simple filtering rule", func() {

		// 	Convey("That has three images", nil)

		// 	Convey("or Three Namespaces", nil)

		// })

		// Convey("or  Three Tags", nil)

	})

}
