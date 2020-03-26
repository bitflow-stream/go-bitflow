# teambitflow/go-bitflow:static-arm32v7
FROM teambitflow/golang-build:1.12-stretch as build
RUN apt-get update && apt-get install -y git gcc-arm-linux-gnueabi
WORKDIR /build

ENV GOOS=linux
ENV GOARCH=arm
ENV CC=arm-linux-gnueabi-gcc
ENV CGO_ENABLED=1

# Copy go.mod first and download dependencies, to enable the Docker build cache
COPY ../go.mod .
RUN sed -i $(find -name go.mod) -e '\_//.*gitignore$_d' -e '\_#.*gitignore$_d'
RUN go mod download

# Copy rest of the source code and build
# Delete go.sum files and clean go.mod files form local 'replace' directives
COPY .. .
RUN find -name go.sum -delete
RUN sed -i $(find -name go.mod) -e '\_//.*gitignore$_d' -e '\_#.*gitignore$_d'
RUN go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /bitflow-pipeline ./cmd/bitflow-pipeline

FROM scratch
COPY --from=build /bitflow-pipeline /
ENTRYPOINT ["/bitflow-pipeline"]
