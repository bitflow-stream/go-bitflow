# teambitflow/golang-build:arm64v8
# docker build -t teambitflow/golang-build:arm64v8 -f arm64v8-build.Dockerfile .
FROM teambitflow/golang-build:alpine
ENV GOOS='linux'
ENV GOARCH='arm64'
ENV CC='aarch64-linux-gnu-gcc'
