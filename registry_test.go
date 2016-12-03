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
			err = d.TagAndPush("alpine", regAddr.remoteName("alpine"), "stable")
			So(err, ShouldBeNil)
			err = d.TagAndPush("alpine", regAddr.remoteName("mynamespace/alpine"), "0.1")
			So(err, ShouldBeNil)
			err = d.TagAndPush("alpine", regAddr.remoteName("alpine"), "0.1")
			So(err, ShouldBeNil)
			err = d.TagAndPush("busybox", regAddr.remoteName("mynamespace/busybox"), "0.1-stable")
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
}
type tagAndPushRecord struct {
	imageName, remoteAddr, tag string
}

func (r tagAndPushRecorder) TagAndPush(imageName, remoteAddr, tag string) error {
	r.records = append(r.records, tagAndPushRecord{imageName, remoteAddr, tag})
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
		regSource  RegistryFactory
		regTarget  RegistryFactory
		repoFilter RepositoryFilter
		tagFilter  TagFilter
		handler    tagAndPushRecorder
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
		}, "regisrtry.production.com"},
		NewNamespaceFilter("production"),
		regExFilter("[\\d\\.]+"),
		tagAndPushRecorder{},
	},
		[]tagAndPushRecord{tagAndPushRecord{"production/tool1", "registry.production.com", "0.2"},
			tagAndPushRecord{"production/tool2", "registry.production.com", "0.1"}}}}
	for _, tt := range tests {
		Convey("for consolidation of:"+tt.name, t, func() {
			Consolidate(tt.args.regSource, tt.args.regTarget, tt.args.repoFilter, tt.args.tagFilter, tt.args.handler)
			So(tt.args.handler.records, ShouldResemble, tt.records)
		})
	}
}
