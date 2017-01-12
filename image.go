package main

import (
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
)

//ImageIdentifier the repository and tag to identify a docker image
type ImageIdentifier struct {
	Name string
	Tag  string
}

// DockerImageFilter Selects particular images, useed with registries
type DockerImageFilter struct {
	repoFilter, tagFilter Filter
}

// Filter generic string filter
type Filter interface {
	Matches(str string) bool
}

//NamespaceFilter a filter for repositories
//in a registry using a particular set of top level names.
//these must be an exact match
type NamespaceFilter struct {
	namespaces map[string]struct{}
}

// NewNamespaceFilter constructs a regex of a number of namespaces in a registry
func NewNamespaceFilter(names ...string) Filter {
	namespaceRegex := strings.Join(names, "|")
	filter, err := NewRegexTagFilter(namespaceRegex)
	if err != nil {
		log.Errorf("Can't transform %s into regex:%s", namespaceRegex, err)
	}
	return filter
}

//RegexTagFilter structure that allows us to
//filter only on particular patterns of labels
//i.e. only things marked stable-.*
type RegexTagFilter struct {
	pattern *regexp.Regexp
}

//Matches filters tags that match the regex
func (r *RegexTagFilter) Matches(str string) bool {
	return r.pattern.Match([]byte(str))
}

//NewRegexTagFilter used to create filter for all versions of
//a given
func NewRegexTagFilter(regex string) (*RegexTagFilter, error) {
	pattern, err := regexp.Compile(regex)
	if err != nil {
		log.Warnf("Couldn't compile to regex %s:%v", regex, err)
		return nil, err
	}
	return &RegexTagFilter{pattern}, nil
}

// Finds all the images that aren't in the target but are in the source
func missingImages(source, target []ImageIdentifier) []ImageIdentifier {
	diffs := make([]ImageIdentifier, 0, len(source))
	vals := make(map[ImageIdentifier]int)
	for _, t := range target {
		vals[t] = 1
	}
	for _, s := range source {
		if _, ok := vals[s]; !ok {
			diffs = append(diffs, s)
		}
	}
	return diffs
}
