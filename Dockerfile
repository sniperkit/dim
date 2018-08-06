############################################################################################################
#######################################################

####################################
## Builder - image arguments
####################################
ARG BUILDER_ALPINE_VERSION=${BUILDER_ALPINE_VERSION:-"3.7"}
ARG BUILDER_GOLANG_VERSION=${BUILDER_GOLANG_VERSION:-"1.10.3"}
ARG BUILDER_IMAGE_TAG=${BUILDER_IMAGE_TAG:-"${BUILDER_GOLANG_VERSION}-alpine${BUILDER_ALPINE_VERSION}"}
ARG BUILDER_IMAGE_NAME=${BUILDER_IMAGE_NAME:-"golang:${BUILDER_IMAGE_TAG}"}

####################################
## Builder
###################################
# FROM ${BUILDER_IMAGE_NAME} AS builder
FROM golang:1.10.3-alpine3.7 AS builder

ARG REPO_VCS=${REPO_VCS:-"github.com"}
ARG REPO_NAMESPACE=${REPO_NAMESPACE:-"sniperkit"}
ARG REPO_PROJECT=${REPO_PROJECT:-"dim"}
ARG REPO_URI=${REPO_URI:-"${REPO_VCS}/${REPO_NAMESPACE}/${REPO_PROJECT}"}

ARG GOLANG_TOOLS_URIS=${GOLANG_TOOLS_URIS:-"github.com/mitchellh/gox \
                                            github.com/Masterminds/glide \
                                            github.com/golang/dep/cmd/dep \
                                            github.com/mattn/gom"}

# RUN apk --no-cache --no-progress add gcc g++ make ca-certificates openssl git cmake make nano bash
RUN apk --no-cache --no-progress add gcc g++ ca-certificates openssl git make mercurial

WORKDIR /go/src/${REPO_URI}

## copy deps definitions
# COPY Gopkg.lock Gopkg.toml ./
COPY glide.lock glide.yaml ./

## install deps
# RUN clear && glide update --strip-vendor
# RUN clear && dep ensure -v
# COPY vendor vendor

## code
COPY pkg pkg
COPY plugin plugin

## binaries
COPY cmd/${REPO_PROJECT} cmd/${REPO_PROJECT}
# COPY cmd/${REPO_PROJECT}-beta ${REPO_PROJECT}-beta

## install commands
RUN for gtu in ${GOLANG_TOOLS_URIS}; do echo "[in progress]... go get -u $gtu" ; go get -u $gtu ; go install $gtu ; done

RUN if [ -f "glide.lock" ]; then \
        glide update --strip-vendor; \
    \
    elif [ -f "Gopkg.lock" ]; then \
        dep ensure -v; \
    fi \
    \
    && go install ./cmd/... \
    \
    && rm -fR $GOPATH/src \
    && rm -fR $GOPATH/pkg \
    \
    && ls -l $GOPATH/bin

############################################################################################################
#######################################################

####################################
## Dist builder - image arguments
####################################

ARG DIST_BUILDER_IMAGE_TAG=${DIST_BUILDER_IMAGE_TAG:-"wheezy"}
ARG DIST_BUILDER_IMAGE_NAME=${DIST_BUILDER_IMAGE_NAME:-"debian:${DIST_BUILDER_IMAGE_TAG}"}

# refs
# - cross-compile in golang
#  - https://virtyx.com/blog/cross-compiling-go-in-docker/
FROM debian:wheezy as xcross
# FROM ${DIST_BUILDER_IMAGE_NAME} as xcross

RUN apt-get update -y \
    && apt-get install --no-install-recommends -y -q \
             curl \
             zip \
             build-essential \
             ca-certificates \
             git mercurial bzr \
    && rm -rf /var/lib/apt/lists/*

ARG GOVERSION=${GOVERSION:-"1.10.3"}

RUN mkdir /goroot \
    && mkdir /gopath

RUN curl -s https://storage.googleapis.com/golang/go${GOVERSION}.linux-amd64.tar.gz \
           | tar xvzf - -C /goroot --strip-components=1

ENV GOPATH=${GOPATH:-"/gopath"}
ENV GOROOT=${GOROOT:-"/goroot"}
ENV PATH "$GOROOT/bin:$GOPATH/bin:$PATH"

ARG DIST_DIR=${DIST_DIR:-"/dist"}
ARG GOLANG_TOOLS_URIS=${GOLANG_TOOLS_URIS:-"github.com/mitchellh/gox \
                                            github.com/Masterminds/glide \
                                            github.com/golang/dep/cmd/dep"}

## install commands
RUN for gtu in ${GOLANG_TOOLS_URIS}; do \
        echo "[in progress]... go get -u $gtu" && \
        go get -v $gtu && \
        go install $gtu ; \
    done \
    \
    && rm -fR $GOPATH/src \
    && rm -fR $GOPATH/pkg \
    \
    && ls -l $GOPATH/bin

ARG REPO_VCS=${REPO_VCS:-"github.com"}
ARG REPO_NAMESPACE=${REPO_NAMESPACE:-"sniperkit"}
ARG REPO_PROJECT=${REPO_PROJECT:-"dim"}
ARG REPO_URI=${REPO_URI:-"${REPO_VCS}/${REPO_NAMESPACE}/${REPO_PROJECT}"}

WORKDIR ${GOPATH}/src/${REPO_URI}

## copy deps definitions
# COPY Gopkg.lock Gopkg.toml ./
COPY glide.lock glide.yaml ./

## install deps
RUN if [ -f "glide.lock" ]; then \
        glide update --strip-vendor; \
    \
    elif [ -f "Gopkg.lock" ]; then \
        dep ensure -v; \
    fi

## code
COPY pkg pkg
COPY plugin plugin
COPY cmd cmd

# re-using above deps install, it will just export without fetching them
RUN mkdir -p ${DIST_DIR} \
    && pwd \
    && ls -la \
    && if [ -f "glide.lock" ]; then \
        glide update --strip-vendor; \
    \
    elif [ -f "Gopkg.lock" ]; then \
        dep ensure -v; \
    fi \
    && gox -os="linux windows darwin" -arch="amd64 386" -output="/dist/${REPO_PROJECT}_{{.OS}}_{{.Arch}}" ./cmd/${REPO_PROJECT} \
    && ls -la /dist

# VOLUME ["/dist"]

############################################################################################################
#######################################################

####################################
## Runner - image arguments
####################################
ARG RUNNER_ALPINE_VERSION=${RUNNER_ALPINE_VERSION:-"3.8"}
ARG RUNNER_IMAGE_NAME=${RUNNER_IMAGE_NAME:-"alpine:${RUNNER_ALPINE_VERSION}"}

####################################
## Build
####################################
FROM alpine:3.8 AS runner
# FROM ${RUNNER_IMAGE_NAME} AS runner

WORKDIR /usr/bin
COPY --from=builder /go/bin ./
# Note: if dist version has to copied
# COPY --from=xcross /dist ./

RUN echo "\n---- DEBUG INFO -----\n" \
    ls -l /usr/bin/${REPO_PROJECT}* \
    echo "\nPATH: ${PATH}\n"
