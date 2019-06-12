# teambitflow/go-bitflow
FROM golang:1.12-alpine as build
RUN apk --no-cache add git gcc g++ musl-dev
WORKDIR /build
COPY . .
RUN go build -o /bitflow-pipeline ./cmd/bitflow-pipeline

FROM alpine:3.9
RUN apk --no-cache add libstdc++
COPY --from=build /bitflow-pipeline /
ENTRYPOINT ["/bitflow-pipeline"]

