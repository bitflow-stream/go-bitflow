FROM golang:1.11 as build

ENV GO111MODULE=on

WORKDIR /go/src/github.com/bitflow-stream/go-bitflow
COPY . .
RUN GOOS=linux GOARCH=amd64 go install -ldflags "-linkmode external -extldflags -static" ./cmd/bitflow-pipeline

FROM alpine
WORKDIR /root/
COPY --from=build /go/bin/bitflow-pipeline .
ENTRYPOINT ["./bitflow-pipeline"]

