#!/bin/bash
home=`dirname $(readlink -e $0)`

if ! type esc &> /dev/null; then
    echo "Installing esc tool..."
	go get github.com/mjibson/esc
fi

cd "$home"
esc -o static_files_generated.go -pkg plotHttp -prefix static/ static
