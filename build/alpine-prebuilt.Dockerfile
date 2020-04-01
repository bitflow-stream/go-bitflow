# bitflowstream/bitflow-pipeline:latest
# Copies pre-built binaries into the container. The binaries are built on the local machine beforehand:
# ./native-build.sh
# docker build -t bitflowstream/bitflow-pipeline:latest -f alpine-prebuilt.Dockerfile _output
FROM alpine:3.11.5
RUN apk --no-cache add libstdc++
COPY bitflow-pipeline /
ENTRYPOINT ["/bitflow-pipeline"]
