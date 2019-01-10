FROM golang:1.11 as build
ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO_LD_FLAGS="-linkmode external -extldflags -static"
WORKDIR /go/src/github.com/bitflow-stream/go-bitflow
COPY . .
RUN go build -ldflags "$GO_LD_FLAGS" -o /bitflow-pipeline ./cmd/bitflow-pipeline
ENTRYPOINT ["/bitflow-pipeline"]

FROM alpine
COPY --from=build /bitflow-pipeline /
ENTRYPOINT ["/bitflow-pipeline"]
