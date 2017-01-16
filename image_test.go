package main

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetMatchingImages(t *testing.T) {
	type args struct {
		reg    Registry
		filter DockerImageFilter
	}
	tests := []struct {
		name    string
		args    args
		want    RegistryTargets
		wantErr bool
	}{
		{
			"simple regex filter",
			args{
				mockRegistry{map[string][]string{
					"alpine":              {"0.1", "stable"},
					"mynamespace/busybox": {"0.1", "0.1-stable"}}},
				DockerImageFilter{matchEverything{}, regExFilter(".*stable")}},
			[]RegistryTarget{RegistryTarget{"alpine", "stable"}, RegistryTarget{"mynamespace/busybox", "0.1-stable"}},
			false,
		},
		{
			"filter out based off namespaces",
			args{
				mockRegistry{map[string][]string{
					"prod/image1":    {"0.1"},
					"staging/image1": {"0.1"},
					"staging/image2": {"0.1"},
					"dev/image1":     {"0.2"},
				}},
				DockerImageFilter{NewNamespaceFilter("staging", "prod"), matchEverything{}}},
			[]RegistryTarget{RegistryTarget{"prod/image1", "0.1"},
				RegistryTarget{"staging/image1", "0.1"},
				RegistryTarget{"staging/image2", "0.1"}},
			false,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMatchingImages(tt.args.reg, tt.args.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMatchingImages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			sort.Sort(got)
			sort.Sort(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMatchingImages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConsolidates(t *testing.T) {

	type args struct {
		regSource Registry
		regTarget Registry
		filter    DockerImageFilter
		handler   *eventRecorder
	}
	tests := []struct {
		name   string
		args   args
		events RegistryEvents
	}{{"production filters", args{
		mockRegistry{map[string][]string{
			"production/tool1": {"0.1", "0.2"},
			"production/tool2": {"0.1", "latest"},
		}},
		mockRegistry{map[string][]string{
			"production/tool1": {"0.1"},
		}},
		DockerImageFilter{NewNamespaceFilter("production"),
			regExFilter("[\\d\\.]+")},
		&eventRecorder{},
	},
		RegistryEvents{[]RegistryEvent{RegistryEvent{"missing", RegistryTarget{"production/tool1", "0.2"}}, RegistryEvent{"missing", RegistryTarget{"production/tool2", "0.1"}}}},
	}}
	for _, tt := range tests {
		Convey("for consolidation of:"+tt.name, t, func() {
			Consolidate(tt.args.regSource, tt.args.regTarget, tt.args.filter, tt.args.handler)
			fmt.Printf("hander %v", tt.args.handler)
			expected := tt.events.getRegistryTargets()
			actual := tt.args.handler.events.getRegistryTargets()
			sort.Sort(expected)
			sort.Sort(actual)
			So(actual, ShouldResemble, expected)
		})
	}
}

type eventRecorder struct {
	events RegistryEvents
	name   string
}

func (r *eventRecorder) Handle(evt RegistryEvent) error {
	r.events.Events = append(r.events.Events, evt)
	return nil
}

func regExFilter(pattern string) *RegexTagFilter {
	f, err := NewRegexTagFilter(pattern)
	if err != nil {
		panic("regex:" + pattern + " does not compute")
	}
	return f
}

// Add in Sort interface for registry events so tests have predictable orderings
func (n RegistryTargets) Len() int {
	return len(n)
}

func (n RegistryTargets) Less(i, j int) bool {
	record1 := n[i]
	record2 := n[j]

	result := strings.Compare(record1.Repository, record2.Repository)
	if result == 0 {
		result = strings.Compare(record1.Tag, record2.Tag)
	}
	return result < 0

}
func (n RegistryTargets) Swap(i, j int) {
	record1 := n[i]
	record2 := n[j]
	n[i] = record2
	n[j] = record1
}

type mockRegistry struct {
	entries map[string][]string
}

func (m mockRegistry) GetRegistry() (Registry, error) {
	return m, nil
}
func (m mockRegistry) Address() string {
	return "mock://"
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
