# teambitflow/golang-build:arm
# docker build -f golang-build-arm.Dockerfile -t teambitflow/golang-build:arm .
FROM teambitflow/golang-build:1.12-stretch as build
RUN apt-get update && apt-get install -y git gcc-arm-linux-gnueabi gcc-aarch64-linux-gnu
WORKDIR /build
