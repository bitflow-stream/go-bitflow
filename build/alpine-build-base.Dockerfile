# teambitflow/golang-build:alpine
# This image is used to build Go programs for alpine images. The purpose of this separate container
# is to mount the Go mod-cache into the container during the build, which is not possible with the 'docker build' command.
# docker build -f alpine-build-base.Dockerfile -t teambitflow/golang-build:alpine .
FROM golang:1.12-alpine
RUN apk --no-cache add git mercurial gcc g++
WORKDIR /build