SOURCEDIR= .
SOURCES := $(shell find $(SOURCEDIR) -type f -name '*.go')
BINARY=registrysync

# H/T https://ariejan.net/2015/10/03/a-makefile-for-golang-cli-tools/
# VERSION=1.0.0
# BUILD_TIME=`date +%FT%T%z`
VERSION ?= 0.1
BUILD_TIME=$(shell date +%FT%T%z)

# LDFLAGS=-ldflags "-X github.com/jmahowald/registrysync/core.Version=${VERSION} -X github.com/jmahowald/registrysync/core.Version=${BUILD_TIME}"

.DEFAULT_GOAL: $(BINARY)

$(BINARY): $(SOURCES) vendor
	go build ${LDFLAGS} -o ${BINARY} cli.go registry.go

vendor:  
	glide install

.PHONY: install

install:
	go install ${LDFLAGS} .

.PHONY: clean
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

docs: $(DOC_SOURCES)
	cd build; go run build.go ../docs

test: $(BINARY)
	go test .



# TODO make this look into asssest directory
# BUILDDIR := src/medliumls

# build: vendor
# 	cd $(BUILDDIR) && go build .
# 	mkdir -p dist
# 	mv $(BUILDDIR)/medliumls dist/medliumls

# vendor:
# 	cd $(BUILDDIR) && glide install
