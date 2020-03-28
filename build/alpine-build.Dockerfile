# teambitflow/golang-build:alpine
# This image is used to build Go programs for alpine images. The purpose of this separate container
# is to mount the Go mod-cache into the container during the build, which is not possible with the 'docker build' command.
# This image is intended to be run on the build host with a volume such as: -v /tmp/go-mod-cache/alpine:/go
# docker build -t teambitflow/golang-build:alpine -f alpine-build.Dockerfile .
FROM golang:1.14.1-alpine
RUN apk --no-cache add curl bash git mercurial gcc g++ docker musl-dev
WORKDIR /build
ENV GO111MODULE=on

# Enable docker-cli experimental features
RUN mkdir ~/.docker && echo -e '{\n\t"experimental": "enabled"\n}' > ~/.docker/config.json
