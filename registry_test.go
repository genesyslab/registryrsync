package main

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFilteringOfARegistry(t *testing.T) {

	Convey("Given a simple registry", t, func() {

		ip, closer, err := StartRegistry()
		So(err, ShouldBeNil)
		So(closer, ShouldNotBeNil)

		defer closer()

		fmt.Printf("ip is %s", ip)

		Convey("With a simple filtering rule", func() {

			Convey("That has three images", nil)

			Convey("or Three Namespaces", nil)

		})

		Convey("or  Three Tags", nil)

	})

}
