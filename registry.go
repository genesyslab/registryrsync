package main

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/heroku/docker-registry-client/registry"
	//	"github.com/docker/distribution/manifest"
)

type RegistryInfo struct {
	address    string
	username   string
	password   string
	isInsecure bool
}

// func (r RegistryInfo) Address() string {
// 	return r.address
// }

//RegistryFactory somethign that can give us pointers to registries
type RegistryFactory interface {
	GetRegistry() (Registry, error)
	Address() string
}

func NewRegistryInfo(url, username, password string) {

}
func (r RegistryInfo) GetRegistry() (Registry, error) {
	var protocol string
	if strings.Index(r.address, "localhost") == 0 {
		protocol = "http"
	} else {
		protocol = "https"
	}

	regUrl := fmt.Sprintf("%s://%s", protocol, r.address)
	log.Infof("Connecting to registry %s:%s", regUrl)

	reg, err := registry.New(regUrl, r.username, r.password)

	if err != nil {
		//TODO should this be fatal?  maybe a warn.
		log.Errorf("Couldn't connect to registry %s:%s", regUrl, err)
		return nil, err
	}
	return reg, nil

}

type DockerImageFilter struct {
	repoFilter RepositoryFilter
	tagFilter  TagFilter
}

type RepositoryFilter interface {
	MatchesRepo(repo string) bool
}

type TagFilter interface {
	MatchesTag(tag string) bool
}

//NamespaceFilter a filter for repositories
//in a registry using a particular set of top level names.
//these must be an exact match
type NamespaceFilter struct {
	namespaces map[string]struct{}
}

func NewNamespaceFilter(names ...string) *NamespaceFilter {
	namespaceSet := make(map[string]struct{}, len(names))
	for _, name := range names {
		namespaceSet[name] = struct{}{}
	}
	return &NamespaceFilter{namespaceSet}
}

func (n *NamespaceFilter) MatchesRepo(repo string) bool {
	pathParts := strings.Split(repo, "/")
	if _, ok := n.namespaces[pathParts[0]]; ok {
		return true
	} else {
		log.Debugf("Checking if repo %v is in %v", pathParts, n)
	}
	return false
}

//RegexTagFilter structure that allows us to
//filter only on particular patterns of labels
//i.e. only things marked stable-.*
type RegexTagFilter struct {
	pattern *regexp.Regexp
}

//MatchesTag filters tags that match the regex
func (r *RegexTagFilter) MatchesTag(tag string) bool {
	return r.pattern.Match([]byte(tag))
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

//Registry abstraction of a docker registry connection
type Registry interface {
	Repositories() ([]string, error)
	Tags(string) ([]string, error)
}

//GetMatchingImages Finds the names and tags of all matching
func GetMatchingImages(reg Registry, filter DockerImageFilter) ([]ImageIdentifier, error) {
	matchingImages := make([]ImageIdentifier, 0, 10)

	repos, err := reg.Repositories()
	if err != nil {
		fmt.Printf("cant get repositories from %s:%v. Got back %s", reg, err, repos)
		return matchingImages, err
	}
	for _, repo := range repos {

		if filter.repoFilter.MatchesRepo(repo) {
			tags, err := reg.Tags(repo)
			if err != nil {
				log.Fatal("Unable to get tags", err)
				return matchingImages, err
			}
			for _, tag := range tags {
				if filter.tagFilter.MatchesTag(tag) {
					matchingImages = append(matchingImages, ImageIdentifier{repo, tag})
				}
			}
		}
	}
	return matchingImages, nil
}

func Consolidate(rs, rt RegistryFactory, filter DockerImageFilter, handler ImageHandler) error {

	regSource, err := rs.GetRegistry()
	if err != nil {
		log.Errorf("Couldn't connect to source registry %v:%v", rs, err)
	}
	regTarget, err := rt.GetRegistry()
	if err != nil {
		log.Errorf("Couldn't connect to source registry %v:%v", rt, err)
	}

	//This could easily take a while and we want to at the least log the time it took. In reality should probably
	//push a metric somewhere
	log.Infof(">>Consolidate(%v,%v,%v", regSource, regTarget, filter)
	defer log.Info("<<Consolidate")

	sourceImages, err := GetMatchingImages(regSource, filter)
	if err != nil {
		log.Errorf("Couldn't get images from source repo %v : %v", regSource, err)
		return err
	}
	targetImages, err := GetMatchingImages(regTarget, filter)
	if err != nil {
		log.Errorf("Couldn't get images from target repo %s : %s", regTarget, err)
		return err
	}
	missingImages := missingImages(sourceImages, targetImages)
	for _, image := range missingImages {
		handler.PullTagPush(image.Name, rs.Address(), rt.Address(), image.Tag)
	}

	return nil

}
