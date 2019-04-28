# teambitflow/go-bitflow:static
FROM golang:1.11-alpine as build
ENV GO111MODULE=on
RUN apk --no-cache add git gcc g++ musl-dev
WORKDIR /build
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /bitflow-pipeline ./cmd/bitflow-pipeline

FROM scratch
COPY --from=build /bitflow-pipeline /
ENTRYPOINT ["/bitflow-pipeline"]

