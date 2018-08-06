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
FROM sniperkit/golang:dev-1.10.3-alpine3.7 AS builder

ARG REPO_VCS=${REPO_VCS:-"github.com"}
ARG REPO_NAMESPACE=${REPO_NAMESPACE:-"sniperkit"}
ARG REPO_PROJECT=${REPO_PROJECT:-"dim"}
ARG REPO_URI=${REPO_URI:-"${REPO_VCS}/${REPO_NAMESPACE}/${REPO_PROJECT}"}

## add apk build dependencies
# RUN apk --no-cache --no-progress add gcc g++ make ca-certificates openssl git cmake make nano bash

WORKDIR /go/src/${REPO_URI}

## deps
COPY Gopkg.lock Gopkg.toml ./
COPY vendor vendor

## pkg
COPY pkg pkg
COPY plugin plugin

## executables
COPY cmd/${REPO_PROJECT} ${REPO_PROJECT}

## install commands
RUN go install ./... \
    && ls -la /go/bin

############################################################################################################
############################################################################################################

####################################
## Builder - image arguments
####################################
ARG RUNNER_ALPINE_VERSION=${RUNNER_ALPINE_VERSION:-"3.7"}
ARG RUNNER_IMAGE_NAME=${RUNNER_IMAGE_NAME:-"alpine:${RUNNER_ALPINE_VERSION}"}

####################################
## Build
####################################
FROM alpine:3.7 AS dist
WORKDIR /usr/bin
COPY --from=builder /go/bin ./
RUN echo "\n---- DEBUG INFO -----\n" \
    ls -l /usr/bin/dim* \
    echo "\nPATH: ${PATH}\n"
