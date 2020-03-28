# teambitflow/golang-build:static-arm32v7
# docker build -t teambitflow/golang-build:static-arm32v7 -f arm32v7-static-build.Dockerfile .
FROM teambitflow/golang-build:debian
RUN apt-get install -y gcc-arm-linux-gnueabi
ENV GOOS=linux
ENV GOARCH=arm
ENV CC=arm-linux-gnueabi-gcc
ENV CGO_ENABLED=1
