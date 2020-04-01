#!/bin/bash

# Alpine based
docker build -t bitflowstream/golang-build:alpine -f alpine-build.Dockerfile .
docker build -t bitflowstream/golang-build:arm32v7 -f arm32v7-build.Dockerfile .
docker build -t bitflowstream/golang-build:arm64v8 -f arm64v8-build.Dockerfile .

# Debian based
docker build -t bitflowstream/golang-build:debian -f debian-build.Dockerfile .
docker build -t bitflowstream/golang-build:static-arm32v7 -f arm32v7-static-build.Dockerfile .
docker build -t bitflowstream/golang-build:static-arm64v8 -f arm64v8-static-build.Dockerfile .

# Push updated images
docker push bitflowstream/golang-build:alpine
docker push bitflowstream/golang-build:arm32v7
docker push bitflowstream/golang-build:arm64v8
docker push bitflowstream/golang-build:debian
docker push bitflowstream/golang-build:static-arm32v7
docker push bitflowstream/golang-build:static-arm64v8
