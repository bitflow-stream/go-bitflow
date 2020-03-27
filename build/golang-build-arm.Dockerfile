# teambitflow/golang-build:arm
# docker build -f golang-build-arm.Dockerfile -t teambitflow/golang-build:arm .

# TODO This image is based on Debian, while arm*-prebuilt.Dockerifle are based on Alpine. Should be consistent.

FROM teambitflow/golang-build:debian
RUN apt-get update && apt-get install -y git gcc-arm-linux-gnueabi gcc-aarch64-linux-gnu
