FROM golang:1.8.1 as build

RUN mkdir -p /go/src/github.com/antongulenko/go-bitflow
COPY . /go/src/github.com/antongulenko/go-bitflow
RUN GOOS=linux GOARCH=amd64 go get -ldflags "-linkmode external -extldflags -static" github.com/antongulenko/go-bitflow/cmd/bitflow-pipeline

FROM alpine
WORKDIR /root/
COPY --from=build /go/bin/bitflow-pipeline .
ENTRYPOINT ["./bitflow-pipeline"]

