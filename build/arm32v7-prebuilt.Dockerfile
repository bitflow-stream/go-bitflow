# bitflowstream/bitflow-pipeline:latest-arm32v7
# Copies pre-built binaries into the container. The binaries are built on the local machine beforehand:
# ./native-build.sh
# docker build -t bitflowstream/bitflow-pipeline:latest-arm32v7 -f arm32v7-prebuilt.Dockerfile _output
FROM arm32v7/alpine:3.11.5
COPY bitflow-pipeline /
ENTRYPOINT ["/bitflow-pipeline"]
