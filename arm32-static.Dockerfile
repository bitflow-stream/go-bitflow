# teambitflow/go-bitflow:static-arm32
FROM teambitflow/golang-build:1.12-stretch as build
RUN apt-get update && apt-get install -y git gcc-arm-linux-gnueabi
WORKDIR /build
COPY . .
RUN env CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /bitflow-pipeline ./cmd/bitflow-pipeline

FROM scratch
COPY --from=build /bitflow-pipeline /
ENTRYPOINT ["/bitflow-pipeline"]
