package main

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

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
	d := dockerCli{}
	Convey("Given a simple but real registry", t, func() {

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
			err = d.PullTagPush("alpine", "", regInfo.RemoteName(), "stable")
			So(err, ShouldBeNil)
			err = d.PullTagPush("alpine", "", regInfo.RemoteName()+"mynamespace/", "0.1")
			So(err, ShouldBeNil)
			err = d.PullTagPush("alpine", "", regInfo.RemoteName(), "0.1")
			So(err, ShouldBeNil)
			err = d.PullTagPush("busybox", "", regInfo.RemoteName()+"mynamespace/", "0.1-stable")
			So(err, ShouldBeNil)

			Convey("We can get back image information from the registry", func() {

				tagFilter, err := NewRegexTagFilter(".*stable")
				So(err, ShouldBeNil)
				namespaceFilter := NewNamespaceFilter("mynamespace")
				So(err, ShouldBeNil)
				Convey("We can get back all the images", func() {
					matches, err := GetMatchingImages(regInfo, DockerImageFilter{matchEverything{}, matchEverything{}})
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
					matches, err := GetMatchingImages(regInfo, DockerImageFilter{matchEverything{}, tagFilter})
					So(err, ShouldBeNil)
					expectedImages := []ImageIdentifier{
						ImageIdentifier{"alpine", "stable"},
						ImageIdentifier{"mynamespace/busybox", "0.1-stable"},
					}
					So(matches, ShouldResemble, expectedImages)
				})
				Convey("we can get all the images in a given namespace", func() {
					matches, err := GetMatchingImages(regInfo, DockerImageFilter{namespaceFilter, matchEverything{}})
					So(err, ShouldBeNil)
					expectedImages := []ImageIdentifier{
						ImageIdentifier{"mynamespace/alpine", "0.1"},
						ImageIdentifier{"mynamespace/busybox", "0.1-stable"},
					}
					So(matches, ShouldResemble, expectedImages)
				})
				Convey("we can get only images in a given namespace and given tags", func() {
					matches, err := GetMatchingImages(regInfo, DockerImageFilter{namespaceFilter, tagFilter})
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

type mockRegistry struct {
	entries map[string][]string
	url     string
}

func (m mockRegistry) GetRegistry() (Registry, error) {
	return m, nil
}
func (m mockRegistry) RemoteName() string {
	return m.url
}

func (m mockRegistry) Repositories() ([]string, error) {
	repos := []string{}
	for r := range m.entries {
		repos = append(repos, r)
	}
	return repos, nil
}

func (m mockRegistry) Tags(repo string) ([]string, error) {
	return m.entries[repo], nil
}

type tagAndPushRecorder struct {
	records []tagAndPushRecord
	name    string
}

type tagAndPushRecord struct {
	imageName, sourceAddr, remoteAddr, tag string
}

func (r *tagAndPushRecorder) PullTagPush(imageName, sourceReg, targetReg, tag string) error {
	r.records = append(r.records, tagAndPushRecord{imageName, sourceReg, targetReg, tag})
	return nil
}

func regExFilter(pattern string) *RegexTagFilter {
	f, err := NewRegexTagFilter(pattern)
	if err != nil {
		panic("regex:" + pattern + " does not compute")
	}
	return f
}

func TestConsolidate(t *testing.T) {

	type args struct {
		regSource RegistryFactory
		regTarget RegistryFactory
		filter    DockerImageFilter
		handler   *tagAndPushRecorder
	}
	tests := []struct {
		name    string
		args    args
		records []tagAndPushRecord
	}{{"production filters", args{
		mockRegistry{map[string][]string{
			"production/tool1": {"0.1", "0.2"},
			"production/tool2": {"0.1", "latest"},
		}, "registry.dev.com"},
		mockRegistry{map[string][]string{
			"production/tool1": {"0.1"},
		}, "registry.production.com"},
		DockerImageFilter{NewNamespaceFilter("production"),
			regExFilter("[\\d\\.]+")},
		&tagAndPushRecorder{},
	},
		[]tagAndPushRecord{tagAndPushRecord{"production/tool1", "registry.dev.com",
			"registry.production.com", "0.2"},
			tagAndPushRecord{"production/tool2", "registry.dev.com",
				"registry.production.com", "0.1"},
		}}}
	for _, tt := range tests {
		Convey("for consolidation of:"+tt.name, t, func() {
			Consolidate(tt.args.regSource, tt.args.regTarget, tt.args.filter, tt.args.handler)
			fmt.Printf("hander %v", tt.args.handler)

			So(tt.args.handler.records, ShouldResemble, tt.records)
		})
	}
}
