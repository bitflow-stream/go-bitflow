# teambitflow/go-bitflow:latest-arm32v7
# Copies pre-built binaries into the container. The binaries are built on the local machine beforehand:
# ./native-build.sh
# docker build -t teambitflow/go-bitflow:latest-arm32v7 -f arm32v7-prebuilt.Dockerfile _output
FROM arm32v7/alpine:3.9
RUN apk --no-cache add libstdc++
COPY bitflow-pipeline /
ENTRYPOINT ["/bitflow-pipeline"]
