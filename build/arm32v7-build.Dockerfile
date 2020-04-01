# bitflowstream/golang-build:arm32v7
# docker build -t bitflowstream/golang-build:arm32v7 -f arm32v7-build.Dockerfile .
FROM bitflowstream/golang-build:alpine
ENV GOOS='linux'
ENV GOARCH='arm'
