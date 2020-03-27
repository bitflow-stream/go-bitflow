# teambitflow/go-bitflow:latest
# Copies pre-built binaries into the container. The binaries are built on the local machine beforehand:
# ./native-build.sh
# docker build -t teambitflow/go-bitflow:latest -f alpine-prebuilt.Dockerfile _output
FROM alpine:3.9
RUN apk --no-cache add libstdc++
COPY bitflow-pipeline /
ENTRYPOINT ["/bitflow-pipeline"]
