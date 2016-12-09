SOURCEDIR= .
SOURCES := $(shell find $(SOURCEDIR) -type f -name '*.go')
BINARY=registryrsync
ROOT_SOURCES := $(wildcard *.go)'
DOCKER_TAG=genesyslab/$(BINARY)

export PROJ_GO_SRC ?=/go/src/github.com/genesyslab/registryrsync/

.DEFAULT_GOAL: $(BINARY)

$(BINARY): $(SOURCES) vendor
	go build ${LDFLAGS} -o ${BINARY} $(ROOT_SOURCES)

vendor:
	glide install

.PHONY: install

install:
	go install ${LDFLAGS} .

.PHONY: clean
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
	if [ -f ${LINUXBINARY} ] ; then rm ${LINUXBINARY} ; fi

docs: $(DOC_SOURCES)
	cd build; go run build.go ../docs

test: $(BINARY)
	go test .



	# This section is used to create a relatively minimal docker image
	# Thx - https://developer.atlassian.com/blog/2015/07/osx-static-golang-binaries-with-docker/


$(LINUXBINARY): $(SOURCES) vendor
	$(MAKE) buildlinux

buildgo:
	CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o $(LINUXBINARY) $(PROJ_GO_SRC)

buildlinux: vendor
	docker build -t build-$(BINARY) -f ./Dockerfile.build .
	docker run -t build-$(BINARY) /bin/true
	# docker cp `docker ps -q -n=1`:/$(LINUXBINARY) .
	# chmod 755 ./$(LINUXBINARY)

builddocker:  $(LINUXBINARY)
	docker build --rm=true --tag=joshmahowald/formterra -f Dockerfile.static .
