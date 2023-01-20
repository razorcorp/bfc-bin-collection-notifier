BINARY = bfc-bin-collection-notifier
GOARCH = amd64

VERSION?=latest

# Symlink into GOPATH
GITHUB_USERNAME=razorcorp
BUILD_DIR=${GOPATH}/src/github.com/${GITHUB_USERNAME}/bfc-bin-collection-notifier
BIN_DIR=${BUILD_DIR}/bin
CURRENT_DIR=\$(shell pwd)
BUILD_DIR_LINK=\$(shell readlink ${BUILD_DIR})
BUILD_ID=$(shell git rev-parse --short HEAD)
TAG_NAME=$(shell git symbolic-ref --short HEAD | sed 's/[^/]*\///' | sed 's/[\/]/-/')
REGISTRY="registry.razorcorp.dev/${GITHUB_USERNAME}"
PUBLIC_REGISTRY="ghcr.io/${GITHUB_USERNAME}"
REGISTRY_AUTH_USER?=

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.VERSION=${VERSION}-#${BUILD_ID}"

export GO111MODULE=on
export GOPROXY=direct
export GOSUMDB=off

run:
	@go run .

linux:
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BIN_DIR}/${BINARY}-linux-${GOARCH}/${VERSION}/${BINARY}/${BINARY} .

build:
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BIN_DIR}/${BINARY} .

package:
	@cp -r resources/ ${BIN_DIR}/${BINARY}-linux-${GOARCH}/${VERSION}/${BINARY}/
	@mkdir -p ${BIN_DIR}/${VERSION}/
	@tar -cvjf ${BIN_DIR}/${VERSION}/${BINARY}-linux-${GOARCH}.tar -C ${BIN_DIR}/${BINARY}-linux-${GOARCH}/${VERSION}/ .

release: linux package docker publish clean

install:
	install ${BIN_DIR}/${BINARY} /usr/local/bin/${BINARY}

upgrade:
	@go get -u ./...

dep:
	@go list -m -u all


docker:
	@echo "Building new image"
	@docker build --tag ${REGISTRY}/${BINARY} \
		--build-arg VERSION="${VERSION}-#${BUILD_ID}" \
		--build-arg GOOS=linux \
		--build-arg GOARCH=${GOARCH} \
		--label VERSION="${VERSION}-#${BUILD_ID}" \
		--label git-commit="${BUILD_ID}" .
	@docker tag ${REGISTRY}/${BINARY}:latest ${REGISTRY}/${BINARY}:${BUILD_ID}

	@docker tag ${REGISTRY}/${BINARY}:latest ${REGISTRY}/${BINARY}:${VERSION}
	@docker tag ${REGISTRY}/${BINARY}:latest ${PUBLIC_REGISTRY}/${BINARY}:${VERSION}

auth:
ifndef REGISTRY_TOKEN
	@echo "Missing environment variable: REGISTRY_TOKEN"
	@echo "GitHub PAT token required. export REGISTRY_TOKEN="
	@exit 1
else
	@echo $${REGISTRY_TOKEN} | docker login ghcr.io -u ${REGISTRY_AUTH_USER} --password-stdin
endif


publish:
	@echo "Publishing docker image"
	@docker push ${REGISTRY}/${BINARY}:latest
	@docker push ${REGISTRY}/${BINARY}:${BUILD_ID}

	@docker push ${REGISTRY}/${BINARY}:${VERSION}
	@docker push ${PUBLIC_REGISTRY}/${BINARY}:${VERSION}
	@echo "Published"

clean:
	@rm -rf ${BIN_DIR}/${BINARY}-*-${GOARCH}
	@rm -rf ${BIN_DIR}/${BINARY}

help:
	@echo "\nUsage: make [command] [option]"
	@echo "\nCommands:"
	@echo "\t run \t\t-- Run dev environment on machine"
	@echo "\t linux \t\t-- Build application for Linux operating systems"
	@echo "\t build \t\t-- Build the application for Darwin"
	@echo "\t package \t-- Tar installation archive of binary and resources"
	@echo "\t release \t-- Release new version of the application (build, package and clean)"
	@echo "\t install \t-- Install the binary to system"
	@echo "\t upgrade \t-- Upgrade Go Mod versions"
	@echo "\t dep \t\t-- Resolve mods in the go.mod file"
	@echo "\t clean \t\t-- Clean build artefacts post package"
	@echo "\t swagger \t-- Launch swagger editor. Docker required"
	@echo "\t docker \t-- Build application for. Docker required"
	@echo "\t publish \t-- Upload Docker image(s) to package manager. Github PAT required"
	@echo "\nOptions:"
	@echo "\t BINARY \t-- Name of the application"
	@echo "\t GOARCH \t-- Architecture of the application"
	@echo "\t VERSION \t-- Application version"

.PHONY: run
