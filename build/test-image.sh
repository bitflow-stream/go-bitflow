#!/usr/bin/env sh

test $# = 1 || { echo "Need 1 parameter: image tag to test"; exit 1; }
IMAGE="bitflowstream/bitflow-pipeline"
TAG="$1"

# Sanity check: image starts, outputs valid JSON, and terminates.
docker run "$IMAGE:$TAG" -json-capabilities | jq . > /dev/null
