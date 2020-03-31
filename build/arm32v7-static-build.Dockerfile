# bitflowstream/golang-build:static-arm32v7
# docker build -t bitflowstream/golang-build:static-arm32v7 -f arm32v7-static-build.Dockerfile .
FROM bitflowstream/golang-build:debian
RUN apt-get install -y gcc-arm-linux-gnueabihf
ENV GOOS=linux
ENV GOARCH=arm
ENV CC=arm-linux-gnueabihf-gcc
ENV CGO_ENABLED=1