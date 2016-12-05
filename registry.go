package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/heroku/docker-registry-client/registry"
	//	"github.com/docker/distribution/manifest"
)

type RegistryInfo struct {
	address    address
	username   string
	password   string
	isInsecure bool
}
type address struct {
	HostIP string
	Port   string
}

func (a address) remoteName(imageName string) string {
	return fmt.Sprintf("%s:%s/%s", a.HostIP, a.Port, imageName)
}

//RegistryFactory somethign that can give us pointers to registries
type RegistryFactory interface {
	GetRegistry() (Registry, error)
	RemoteName() string
}

func (r RegistryInfo) RemoteName() string {
	return r.address.HostIP + ":" + r.address.Port + "/"
}
func (r RegistryInfo) GetRegistry() (Registry, error) {

	var protocol string
	if r.isInsecure {
		protocol = "http"
	} else {
		protocol = "https"
	}

	regUrl := fmt.Sprintf("%s://%s:%s", protocol, r.address.HostIP, r.address.Port)
	reg, err := registry.New(regUrl, r.username, r.password)

	if err != nil {
		//TODO should this be fatal?  maybe a warn.
		log.Warnf("Couldn't connect to registry %s:%s", regUrl, err)
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
func GetMatchingImages(regFactory RegistryFactory, filter DockerImageFilter) ([]ImageIdentifier, error) {
	matchingImages := make([]ImageIdentifier, 0, 10)

	reg, err := regFactory.GetRegistry()
	if err != nil {
		return matchingImages, err
	}
	repos, err := reg.Repositories()
	if err != nil {
		fmt.Printf("cant get repositories from %s:%v. Got back %s", regFactory, err, repos)
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

//TagAndPush given a basic image name, will add it to the remote repository with the given tag
func (d dockerCli) TagAndPush(imageName string, remoteAddr string, tag string) error {
	if tag != "" {
		remoteAddr = remoteAddr + ":" + tag
	}
	err := tagImage(imageName, remoteAddr)
	if err != nil {
		log.Warnf("Error tagging %s:%s", imageName, remoteAddr)
		return err
	}

	pushCmd := exec.Command("docker", "push", remoteAddr)
	data, err := pushCmd.CombinedOutput()
	if err != nil {
		log.Printf("Error pushing %s:%s  Output %s", pushCmd.Args, err, string(data))
		return err
	}
	return nil

}

//TagAndPush given a basic image name, will add it to the remote repository with the given tag
//note that is assumed that for remote registries the url
//ends in a /  (this allows us to use the same logic to pull from docker hub, although there
//are definetly smarter ways
func (d dockerCli) PullTagPush(imageName, sourceReg, targetReg, tag string) error {

	imageParts := strings.Split(imageName, ":")

	// Should look into using gorouties and channels
	pullCmd := exec.Command("docker", "pull", sourceReg+imageName)
	data, err := pullCmd.CombinedOutput()
	if err != nil {
		log.Printf("Error pulling %s:%s  Output %s", pullCmd.Args, err, string(data))
		return err
	}

	imageId := imageParts[0]
	if tag != "" {
		imageId += ":" + tag
	}
	remoteName := targetReg + imageId
	err = tagImage(imageName, remoteName)
	if err != nil {
		log.Warnf("Error tagging %s:%s", imageId, remoteName)
		return err
	}

	pushCmd := exec.Command("docker", "push", remoteName)
	data, err = pushCmd.CombinedOutput()
	if err != nil {
		log.Printf("Error pushing %s:%s  Output %s", pushCmd.Args, err, string(data))
		return err
	}

	return nil
}

type dockerCli struct{}

type ImageHandler interface {
	PullTagPush(imageName, sourceReg, targetReg, tag string) error
}

func Consolidate(regSource, regTarget RegistryFactory, filter DockerImageFilter, handler ImageHandler) error {

	//This could easily take a while and we want to at the least log the time it took. In reality should probably
	//push a metric somewhere
	log.Infof(">>Consolidate(%s,%s,%v,%v", regSource.RemoteName(), regTarget.RemoteName(), filter)
	defer log.Info("<<Consolidate")

	sourceImages, err := GetMatchingImages(regSource, filter)
	if err != nil {
		log.Errorf("Couldn't get images from source repo %s : %s", regSource.RemoteName(), err)
		return err
	}
	targetImages, err := GetMatchingImages(regTarget, filter)
	if err != nil {
		log.Errorf("Couldn't get images from target repo %s : %s", regSource.RemoteName(), err)
		return err
	}
	missingImages := missingImages(sourceImages, targetImages)
	for _, image := range missingImages {
		handler.PullTagPush(image.Name, regSource.RemoteName(), regTarget.RemoteName(), image.Tag)
	}

	return nil

}
