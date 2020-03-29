# bitflowstream/bitflow-pipeline:static
# Copies pre-built static binaries into the container. The binaries are built on the local machine beforehand:
# ./native-static-build.sh
# docker build -t bitflowstream/bitflow-pipeline:static -f static-prebuilt.Dockerfile _output/static
FROM scratch
COPY bitflow-pipeline /
ENTRYPOINT ["/bitflow-pipeline"]
