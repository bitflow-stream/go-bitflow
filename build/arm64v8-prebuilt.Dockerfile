# bitflowstream/bitflow-pipeline:latest-arm64v8
# Copies pre-built binaries into the container. The binaries are built on the local machine beforehand:
# ./native-build.sh
# docker build -t bitflowstream/bitflow-pipeline:latest-arm64v8 -f arm64v8-prebuilt.Dockerfile _output
FROM arm64v8/alpine:3.11.5
COPY bitflow-pipeline /
ENTRYPOINT ["/bitflow-pipeline"]
