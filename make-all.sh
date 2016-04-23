#!/bin/bash
set -e

if [ "$1" = "install" ]; then
    # === Package depedencies
    dpkg -l libvirt-dev &> /dev/null || sudo apt-get install -y libvirt-dev

    # === Go dependencies
    path=$(go list -e -f '{{.Dir}}' gitlab.tubit.tu-berlin.de/CIT-Huawei/monitoring)
    cd "$path"
    gpm
fi

# === Build & Install
go install gitlab.tubit.tu-berlin.de/CIT-Huawei/monitoring/data-collection-agent
cp $(which data-collection-agent) bin/data-collection-agent
