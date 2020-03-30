# bitflowstream/golang-build:arm64v8
# docker build -t bitflowstream/golang-build:arm64v8 -f arm64v8-build.Dockerfile .
FROM bitflowstream/golang-build:alpine
ENV GOOS='linux'
ENV GOARCH='arm64'
