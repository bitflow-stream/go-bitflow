#!/bin/bash
home=`dirname $(readlink -f $0)`
root=`readlink -f "$home/.."`
cd "$home"
go build -o "$home/_output/bitflow-pipeline" $@ "$root/cmd/bitflow-pipeline"
