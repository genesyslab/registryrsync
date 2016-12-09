package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	docker "github.com/fsouza/go-dockerclient"
	. "github.com/smartystreets/goconvey/convey"
)

// In the case of a single exposed port on a client will
// return the appropraite connectivity information.  This is
// needed becasue at least on OSX you CANT use the ips of the containers
// as of this writing, and docker-machine/boot2docker are no longer
// the official versions

// waitReachable waits for hostport to became reachable for the maxWait time.
func waitReachable(hostport string, maxWait time.Duration) error {
	done := time.Now().Add(maxWait)
	for time.Now().Before(done) {
		c, err := net.Dial("tcp", hostport)
		if err == nil {
			c.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("cannot connect %v for %v", hostport, maxWait)
}

// waitStarted waits for container to start for the maxWait time.
func waitStarted(client *docker.Client, id string, maxWait time.Duration) error {
	done := time.Now().Add(maxWait)
	for time.Now().Before(done) {
		c, err := client.InspectContainer(id)
		if err != nil {
			//This is to be expected so probably will remove log message later
			log.Warnf("Container not started %s %s", c, err)
			break
		}
		if c.State.Running {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("cannot start container %s for %v", id, maxWait)
}

//StartRegistry - starts a new registry, returning the ip port combination
//and a closing function that you should use for cleanup.  Error of course if there is a problem

func StartRegistry() (address, func(), error) {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return address{}, nil, err
	}
	c, err := client.CreateContainer(createOptions())

	log.Debugf("Container created %+v", c)
	if err != nil {
		log.Fatal("Couldn't create even a basic container:", err)
		return address{}, nil, err
	}
	deferFn := func() {
		if err := client.RemoveContainer(docker.RemoveContainerOptions{
			ID:    c.ID,
			Force: true,
		}); err != nil {
			log.Warnf("cannot remove container: %s", err)
		}
	}

	err = client.StartContainer(c.ID, nil)
	if err != nil {
		log.Fatalf("Could not start container %s : ", c, err)
		deferFn()
		return address{}, nil, err
	}

	// wait for container to wake up
	if err := waitStarted(client, c.ID, 5*time.Second); err != nil {
		deferFn()
		return address{}, nil, err
	}
	if c, err = client.InspectContainer(c.ID); err != nil {
		deferFn()
		return address{}, nil, err
	}

	ports := c.NetworkSettings.Ports
	//We know that the registry listens on port 5000 generally
	//but we don't know what it will actually be mapped to when testing
	port := docker.Port("5000/tcp")
	if portInfo, ok := ports[port]; ok && len(portInfo) >= 1 {
		hostIp := portInfo[0].HostIP
		hostPort := portInfo[0].HostPort
		portAddress := hostIp + ":" + hostPort

		if err := waitReachable(portAddress, 5*time.Second); err != nil {
			deferFn()
			return address{}, nil, err
		}

		//localhost is by default I think marked as
		//an insecure registry, and since this is just for tests ..
		if hostIp == "0.0.0.0" {
			hostIp = "localhost"
		}
		return address{hostIp, hostPort}, deferFn, nil
	} else {
		//Close it
		deferFn()

		return address{}, nil, fmt.Errorf("Coudln't find port %v in ports %v settings %v", port, ports, c.NetworkSettings)
	}
}

func createOptions() docker.CreateContainerOptions {
	ports := make(map[docker.Port]struct{})
	hostConfig := &docker.HostConfig{
		PublishAllPorts: true,
	}
	opts := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        "registry:2.4",
			PortSpecs:    []string{"5000"},
			ExposedPorts: ports,
		},
		HostConfig: hostConfig,
	}

	return opts
}

type matchEverything struct{}

func (m matchEverything) MatchesRepo(name string) bool {
	return true
}

func (m matchEverything) MatchesTag(name string) bool {
	return true
}

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
