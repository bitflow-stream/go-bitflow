#!/bin/bash
home=`dirname $(readlink -f $0)`
root=`readlink -f "$home/.."`
cd "$home"
export CGO_ENABLED=1
go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o "$home/_output/static/bitflow-pipeline" $@ "$root/cmd/bitflow-pipeline"
