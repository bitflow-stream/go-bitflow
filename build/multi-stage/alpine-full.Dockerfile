# teambitflow/go-bitflow
# Build from root of the repository:
# docker build -t teambitflow/go-bitflow build/multi-stage/alpine-full.Dockerfile .
FROM golang:1.12-alpine as build
RUN apk --no-cache add git gcc g++ musl-dev
WORKDIR /build

# Copy go.mod first and download dependencies, to enable the Docker build cache
COPY ../../go.mod .
RUN sed -i $(find -name go.mod) -e '\_//.*gitignore$_d' -e '\_#.*gitignore$_d'
RUN go mod download

# Copy rest of the source code and build
# Delete go.sum files and clean go.mod files form local 'replace' directives
COPY ../.. .
RUN find -name go.sum -delete
RUN sed -i $(find -name go.mod) -e '\_//.*gitignore$_d' -e '\_#.*gitignore$_d'
RUN go build -o /bitflow-pipeline ./cmd/bitflow-pipeline

FROM alpine:3.9
RUN apk --no-cache add libstdc++
COPY --from=build /bitflow-pipeline /
ENTRYPOINT ["/bitflow-pipeline"]
