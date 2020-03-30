# bitflowstream/bitflow-pipeline:static-arm64v8
# Build from root of the repository:
# docker build -t bitflowstream/bitflow-pipeline:static-arm64v8 -f build/multi-stage/arm64v8-static.Dockerfile .
FROM golang:1.14.1-buster as build
RUN apt-get update && apt-get install -y git mercurial qemu-user gcc-aarch64-linux-gnu
WORKDIR /build
ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=arm64
ENV CC=aarch64-linux-gnu-gcc
ENV CGO_ENABLED=1

# Copy go.mod first and download dependencies, to enable the Docker build cache
COPY go.mod .
RUN sed -i $(find -name go.mod) -e '\_//.*gitignore$_d' -e '\_#.*gitignore$_d'
RUN go mod download

# Copy rest of the source code and build
# Delete go.sum files and clean go.mod files form local 'replace' directives
COPY . .
RUN find -name go.sum -delete
RUN sed -i $(find -name go.mod) -e '\_//.*gitignore$_d' -e '\_#.*gitignore$_d'
RUN ./build/native-static-build.sh

FROM scratch
COPY --from=build /build/build/_output/static/bitflow-pipeline /
ENTRYPOINT ["/bitflow-pipeline"]
