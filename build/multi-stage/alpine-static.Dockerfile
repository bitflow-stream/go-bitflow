# teambitflow/go-bitflow:static
# Build from root of the repository:
# docker build -t teambitflow/go-bitflow:static build/multi-stage/alpine-static.Dockerfile .
FROM golang:1.12-alpine as build
RUN apk --no-cache add git gcc g++ musl-dev
WORKDIR /build

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

# Copy go.mod first and download dependencies, to enable the Docker build cache
COPY go.mod .
RUN sed -i $(find -name go.mod) -e '\_//.*gitignore$_d' -e '\_#.*gitignore$_d'
RUN go mod download

# Copy rest of the source code and build
# Delete go.sum files and clean go.mod files form local 'replace' directives
COPY . .
RUN find -name go.sum -delete
RUN sed -i $(find -name go.mod) -e '\_//.*gitignore$_d' -e '\_#.*gitignore$_d'
RUN go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /bitflow-pipeline ./cmd/bitflow-pipeline

FROM scratch
COPY --from=build /bitflow-pipeline /
ENTRYPOINT ["/bitflow-pipeline"]