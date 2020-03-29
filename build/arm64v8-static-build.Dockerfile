# bitflowstream/golang-build:static-arm64v8
# docker build -t bitflowstream/golang-build:static-arm64v8 -f arm64v8-static-build.Dockerfile .
FROM bitflowstream/golang-build:debian
RUN apt-get install -y gcc-aarch64-linux-gnu
ENV GOOS=linux
ENV GOARCH=arm64
ENV CC=aarch64-linux-gnu-gcc
ENV CGO_ENABLED=1
