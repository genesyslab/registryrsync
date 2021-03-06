package main

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// Repository string
// Tag        string
//ImageIdentifier the repository and tag to identify a docker image
type ImageIdentifier struct {
	Repository string
	Tag        string
}

// DockerImageFilter Selects particular images, useed with registries
type DockerImageFilter struct {
	repoFilter, tagFilter Filter
}

// Filter generic string filter
type Filter interface {
	Matches(str string) bool
}

type pusher interface {
	Push(string) error
}
type puller interface {
	Pull(string) error
}
type tagger interface {
	Tag(string, string) error
}
type regSource struct {
	puller
	RegistryFactory
}
type regTarget struct {
	pusher
	RegistryFactory
}

// ImageHandler - knows how to pull, push and tag images
type ImageHandler struct {
	source regSource
	target regTarget
	tagger tagger
	filter DockerImageFilter
}

func (i ImageHandler) Handle(evt RegistryEvent) error {
	if i.filter.repoFilter.Matches(evt.Target.Repository) &&
		i.filter.tagFilter.Matches(evt.Target.Tag) {
		return i.PullTagPush(evt.Target.Repository, evt.Target.Tag)
	} else {
		log.Debugf("Ignoring change  %s", evt)
	}
	return nil
}

func (i ImageHandler) PullTagPush(imageName, version string) error {

	log.Infof(">>PullTagPush(%s:%s)", imageName, version)
	defer log.Infof("<<PullTagPush")
	err := i.source.Pull(fmt.Sprintf("%s:%s", imageName, version))
	if err != nil {
		log.Warnf("Couldn't pull down %s : %s", imageName, err)
		return err
	}
	if version == "" {
		log.Warnf("Pushing image %s without specific tag. Using latest", imageName)
		version = "latest"
	}
	var remoteImgName string
	remoteImgName = fmt.Sprintf("%s/%s:%s", i.target.Address(), imageName, version)
	log.Debugf("Taggin %s to %s", imageName, remoteImgName)

	err = i.tagger.Tag(imageName, remoteImgName)
	if err != nil {
		log.Warnf("Couldn't pull down %s : %s", imageName, err)
		return err
	}
	err = i.target.Push(remoteImgName)
	if err != nil {
		log.Warnf("Couldn't push %s : %s", remoteImgName, err)
		return err
	}
	return nil
}

func (i ImageHandler) RSync(filter DockerImageFilter) error {
	s, err := i.source.GetRegistry()
	if err != nil {
		log.Errorf("Couldn't connec to registry %s : %s", i.source, err)
		return err
	}
	t, err := i.target.GetRegistry()
	if err != nil {
		log.Errorf("Couldn't connec to registry %s : %s", i.target, err)
		return err
	}
	return Consolidate(s, t, filter, i)
}

type matchEverything struct{}

func (m matchEverything) Matches(name string) bool {
	return true
}

// NewNamespaceFilter constructs a regex of a number of namespaces in a registry
func NewNamespaceFilter(names ...string) Filter {
	nameRegexs := make([]string, 0, 2)
	for _, name := range names {
		nameRegexs = append(nameRegexs, fmt.Sprintf("%s/.*", name))
	}
	namespaceRegex := strings.Join(nameRegexs, "|")
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

//GetMatchingImages Finds the names and tags of all matching
func GetMatchingImages(reg Registry, filter DockerImageFilter) (RegistryTargets, error) {
	log.Debugf(">> GetMatchingImages(%+v, %+v)", reg, filter)
	matchingImages := make([]RegistryTarget, 0, 10)
	repos, err := reg.Repositories()
	if err != nil {
		fmt.Printf("cant get repositories from %s:%v. Got back %s", reg, err, repos)
		return matchingImages, err
	}
	for _, repo := range repos {
		if filter.repoFilter.Matches(repo) {
			tags, err := reg.Tags(repo)
			if err != nil {
				log.Fatal("Unable to get tags", err)
				return matchingImages, err
			}
			for _, tag := range tags {
				log.Debugf("Looking at tag %s", tag)
				if filter.tagFilter.Matches(tag) {
					matchingImages = append(matchingImages, RegistryTarget{repo, tag})
				} else {
					log.Debug("Ignoring image from repo %s with tag %s", repo, tag)
				}
			}
		} else {
			log.Debugf("Ignoring repo %s", repo)
		}
	}
	return matchingImages, nil
}

// Finds all the images that aren't in the target but are in the source
func missingImages(source, target RegistryTargets) RegistryTargets {
	diffs := make([]RegistryTarget, 0, len(source))
	vals := make(map[RegistryTarget]int)
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

//Consolidate  finds the missing images in the target from the source and fires off events for those
func Consolidate(regSource, regTarget Registry, filter DockerImageFilter, handler RegistryEventHandler) error {
	//This could easily take a while and we want to at the least log the time it took. In reality should probably
	//push a metric somewhere
	log.Infof(">>Consolidate(%s,%+v,%+v", regSource, regTarget, filter)
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
		handler.Handle(RegistryEvent{"missing", image})
	}
	return nil

}
