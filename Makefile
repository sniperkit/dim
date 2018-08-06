## DIM - Makefile

PWD = $(shell pwd)
SEPARATOR := "-"

APP_PREFIX := "dim"
APP_SUFFIX := ""
APP_NAME := "$(APP_PREFIX)$(SEPARATOR)$(APP_DIRNAME)$(APP_SUFFIX)"

# determine platform
ifeq (Darwin, $(findstring Darwin, $(shell uname -a)))
  PLATFORM := darwin
else
  PLATFORM := linux
endif

PLATFORM_VERSION ?= $(shell uname -r)
PLATFORM_ARCH ?= $(shell uname -m)
PLATFORM_INFO ?= $(shell uname -a)
PLATFORM_OS ?= $(shell uname -s)

APP_ROOT := $(shell pwd)
APP_DIRNAME := $(shell basename `pwd`)
APP_PKG_URI ?= $(shell pwd | sed "s\#$(GOPATH)/src/\#\#g")
APP_PKG_URI_ARRAY ?= $(shell pwd | sed "s\#$(GOPATH)/src/\#\#g" | tr "/" "\n")

GO_EXECUTABLE ?= $(shell which go)
GO_VERSION ?= $(shell $(GO_EXECUTABLE) version)

GOX_EXECUTABLE ?= $(shell which gox)
GOX_VERSION ?= "master"

GLIDE_EXECUTABLE ?= $(shell which glide)
GLIDE_VERSION ?= $(shell $(GLIDE_EXECUTABLE) --version)

GODEPS_EXECUTABLE ?= $(shell which dep)
GODEPS_VERSION ?= $(shell $(GODEPS_EXECUTABLE) version | tr -s ' ')

GIT_EXECUTABLE ?= $(shell which git)
GIT_VERSION ?= $(shell $(GIT_EXECUTABLE) version)

SHELL := /bin/bash
BINARY=dim

DIST_DIR := "$(CURDIR)/shared/dist"
BIN_DIR := "$(CURDIR)/shared/bin"
SEPARATOR := "-"
INSTALL_DIR := "/usr/local/bin/"

VET_DIR := $(shell find . -maxdepth 1 -type d | grep -Ev "(^\./\.|\./vendor|\./shared/dist|\./tests|^\.$$)" | sed  -e 's,.*,&/...,g' )
TEST_DIR := $(shell find . -maxdepth 1 -type d | grep -Ev "(^\./\.|\./vendor|\./shared/dist|\./tests|\./shared/integration|^\.$$)" | sed  -e 's,.*,&/...,g' )
DIR_SOURCES :=  $(shell find . -maxdepth 1 -type d | grep -Ev "(^\./\.|\./vendor|\./dist|\./tests|^\.$$)" | sed  -e 's,\./\(.*\),\1/...,g')
GOIMPORTS_SOURCES := $(shell find . -maxdepth 1 -type d | grep -Ev "(^\./\.|\./vendor|\./shared/dist|\./tests|^\.$$)" | sed  -e 's,\./\(.*\),\1/,g')

SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

git_tag = $(shell git describe --tags --long | sed -e 's/-/./g' | awk -F '.' '{print $$1"."$$2"."$$3+$$4}')

VERSION ?= $(shell git describe --tags)
VERSION_INCODE = $(shell perl -ne '/^var version.*"([^"]+)".*$$/ && print "v$$1\n"' main.go)
VERSION_INCHANGELOG = $(shell perl -ne '/^\# Release (\d+(\.\d+)+) / && print "$$1\n"' CHANGELOG.md | head -n1)

VCS_GIT_REMOTE_URL = $(shell git config --get remote.origin.url)
VCS_GIT_VERSION ?= $(VERSION)

.PHONY: print-% fmt default
print-%: ; @echo $*=$($*)

default: $(BINARY)

all: clean fmt lint vet test dim integration_tests docker install gox-dist

$(BINARY): $(SOURCES)
	CGO_ENABLED=0 go build -a -installsuffix cgo -o $(BINARY) -ldflags "-s -X main.Version=$(git_tag)" .

gox-dist:
	@echo "not ready yet"

distribution:
	@rm -rf ./shared/dist && mkdir -p ./shared/dist
	@docker run --rm -v "$$PWD":/go/src/github.com/sniperkit/dim -w /go/src/github.com/sniperkit/dim -e GOOS=windows -e GOARCH=amd64 golang:1.10.3-alpine go build -o ./shared/dist/$(BINARY)-windows.exe -ldflags "-s -X main.Version=$(git_tag)"
	@docker run --rm -v "$$PWD":/go/src/github.com/sniperkit/dim -w /go/src/github.com/sniperkit/dim -e GOOS=linux -e GOARCH=amd64 golang:1.10.3-alpine go build -o ./shared/dist/$(BINARY)-linux-x64 -ldflags "-s -X main.Version=$(git_tag)"
	@docker run --rm -v "$$PWD":/go/src/github.com/sniperkit/dim -w /go/src/github.com/sniperkit/dim -e GOOS=darwin -e GOARCH=amd64 golang:1.10.3-alpine go build -o ./shared/dist/$(BINARY)-darwin -ldflags "-s -X main.Version=$(git_tag)"

# add more docker targets
docker: $(BINARY)
	@docker build -t sniperkit/dim:$(git_tag) --target dist .

.PHONY: docker docker-pull docker-push docker-xcross docker-console

docker: docker-pull
	@docker build -t sniperkit/dim:go1.10.3-alpine3.7-dev .

docker-all: docker-console docker-xcross

docker-runner:
	@docker build -t sniperkit/dim:go1.10.3-alpine3.7-prod --target=runner .

docker-console:
	@docker build -t sniperkit/dim:go1.10.3-alpine3.7-console -f Dockerfile.console .

docker-xcross: docker
	@docker build -t sniperkit/dim:go1.10.3-debian-wheezy-dist --target=xcross .
	@docker run -ti --rm -e CGO_ENABLED=0 \
	-v $(CURDIR):/gopath/src/github.com/sniperkit/dim \
	-w /gopath/src/github.com/sniperkit/dim \
	sniperkit/dim:go1.10.3-debian-wheezy-dist \
	gox \
	-osarch="darwin/amd64 darwin/386 linux/amd64 linux/386 windows/amd64 windows/386" \
	-output "dist/{{.Dir}}_{{.OS}}_{{.Arch}}"

docker-xcross-standalone:
	@docker build -t sniperkit/dim:go1.10.3-debian-wheezy-standalone -f Dockerfile.xcross .

docker-push:
	@docker push sniperkit/dim:go1.10.3-debian-wheezy-dist
	@docker push sniperkit/dim:go1.10.3-alpine3.7-dev

# $(shell date -v-1d +%Y-%m-%d)
build: build-$(PLATFORM) ## build local executable of the default dim cli version
default: build-$(PLATFORM)

.PHONY: run run-linux run-darwin run-darwin-pro run-windows-pro run-windows-pro
run: run-$(PLATFORM)

.PHONY: build-linux build-darwin build-darwin-pro build-windows build-windows-pro
build-dist: build-linux build-darwin build-darwin-pro build-windows build-windows-pro ## build dim for all platforms

.PHONY: install-linux install-darwin
install: install-$(PLATFORM) ## install dim for your local platform

run-linux: ## run dim for linux (64bits)
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go run ./cmd/dim/*.go

build-linux: ## build dim for linux (64bits)
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BIN_DIR)/dim -v github.com/sniperkit/dim/cmd/dim
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(DIST_DIR)/dim_linux_amd64 -v github.com/sniperkit/dim/cmd/dim

install-linux:
	@go install github.com/sniperkit/dim/cmd/dim
	@dim version

build-darwin: ## build dim for MacOSX (64bits)
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o $(BIN_DIR)/dim -v github.com/sniperkit/dim/cmd/dim
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o $(DIST_DIR)/dim_linux_amd64 -v github.com/sniperkit/dim/cmd/dim

run-darwin: ## run dim for MacOSX (64bits)
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go run -race ./cmd/dim/*.go

install-darwin: ## install dim for MacOSX (64bits)
	@go install github.com/sniperkit/dim/cmd/dim
	@dim version

build-windows: ## build dim for Windows (64bits)
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o $(DIST_DIR)/dim_windows_amd64.exe -v github.com/sniperkit/dim/cmd/dim

run-windows: ## run dim for Windows (64bits)
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go run ./cmd/dim/*.go

# todo: add more detailled build info ./pkg/version
install:
	@go clean -i
	@go install -ldflags "-s -X main.Version=$(git_tag)" github.com/sniperkit/dim/cmd/dim

.PHONY: clean install vet lint fmt

clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

test: fmt
	@go test ${TEST_DIR}

vet: fmt
	@go vet ${VET_DIR}
	@go vet cmd/dim/main.go

lint: fmt
	@for d in $(DIR_SOURCES); do golint $$d; done
	@golint main.go

fmt:
	@goimports -w ${GOIMPORTS_SOURCES}
	@go fmt ${VET_DIR}


completion:
	@go run main.go autocomplete
	@sudo mv dim_compl /etc/bash_completion.d/dim_compl
	@@echo "run source ~/.bashrc to refresh completion"

dim_integration_tests: $(BINARY)
	@dim lift -d --build
	@go test ./shared/integration/...
	@dim stop && dim rm -fv

integration_tests: $(BINARY)
	@docker-compose up -d --build
	@go test ./shared/integration/...
	@docker-compose stop && docker-compose rm -fv

current_version:
	@echo $(git_tag)

version_bump:
	git pull --tags
	n=$$(git describe --tags --long | sed -e 's/-/./g' | awk -F '.' '{print $$4}'); \
	maj=$$(git log --format=oneline -n $$n | grep "#major"); \
	min=$$(git log --format=oneline -n $$n | grep "#minor"); \
	if [ -n "$$maj" ]; then \
		TAG=$(shell git describe --tags --long | sed -e 's/-/./g' | awk -F '.' '{print $$1+1".0.0"}'); \
	elif [ -n "$$min" ]; then \
		TAG=$(shell git describe --tags --long | sed -e 's/-/./g' | awk -F '.' '{print $$1"."$$2+1".0"}'); \
	else \
		TAG=$(shell git describe --tags --long | sed -e 's/-/./g' | awk -F '.' '{print $$1"."$$2"."$$3+$$4+1}'); \
	fi; \
	git tag -a -m "Automatic version bump" $$TAG
	git push --tags

.PHONY: help
help: ## display available makefile targets for this project
	@echo "\033[36mMAKEFILE TARGETS:\033[0m"
	@echo "- PLATFORM: $(PLATFORM)"
	@echo "- PWD: $(PWD)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "- \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)