# teambitflow/golang-build:arm32v7
# docker build -t teambitflow/golang-build:arm32v7 -f arm32v7-build.Dockerfile .
FROM teambitflow/golang-build:alpine
ENV GOOS='linux'
ENV GOARCH='arm'
ENV CC='arm-linux-gnueabi-gcc'
