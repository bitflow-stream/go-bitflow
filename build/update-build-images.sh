#!/bin/bash

# Alpine based
docker build -t teambitflow/golang-build:alpine -f alpine-build.Dockerfile .
docker build -t teambitflow/golang-build:arm32v7 -f arm32v7-build.Dockerfile .
docker build -t teambitflow/golang-build:arm64v8 -f arm64v8-build.Dockerfile .

# Debian based
docker build -t teambitflow/golang-build:debian -f debian-build.Dockerfile .
docker build -t teambitflow/golang-build:static-arm32v7 -f arm32v7-static-build.Dockerfile .
docker build -t teambitflow/golang-build:static-arm64v8 -f arm64v8-static-build.Dockerfile .

# Push updated images
docker push teambitflow/golang-build:alpine
docker push teambitflow/golang-build:arm32v7
docker push teambitflow/golang-build:arm64v8
docker push teambitflow/golang-build:debian
docker push teambitflow/golang-build:static-arm32v7
docker push teambitflow/golang-build:static-arm64v8
