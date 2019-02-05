FROM golang:1.11-alpine as build
ENV GO111MODULE=on
RUN apk --no-cache add git gcc g++ musl-dev
WORKDIR /build
COPY . .
RUN go build -o /bitflow-pipeline ./cmd/bitflow-pipeline
ENTRYPOINT ["/bitflow-pipeline"]

FROM alpine
RUN apk --no-cache add libstdc++
COPY --from=build /bitflow-pipeline /
ENTRYPOINT ["/bitflow-pipeline"]

