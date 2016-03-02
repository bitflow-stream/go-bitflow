#!/bin/bash
set -e
path=$(go list -e -f '{{.Dir}}' github.com/citlab/monitoring)
cd "$path"
gpm
go install github.com/citlab/monitoring/data-collection-agent
