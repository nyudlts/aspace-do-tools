#!/bin/bash

set -u

source './build.cfg' || exit 1

build(){
    os="$1"
    arch="$2"
    printf "build:GOOS=${os}:GOARCH=${arch}:"
    env GOOS=${os} GOARCH=${arch} go build -o ${bin_path_relative_to_build_dir}/${os}/${arch}/${executable}
    if [[ $? == 0 ]]; then
	echo "SUCCESS"
    else
	echo "ERROR: problem with build" >&2
	exit 1
    fi
}

pushd "$build_dir" &>/dev/null
build "linux"  "amd64"
build "darwin" "amd64"
build "windows" "amd64"
build "darwin" "arm64"
popd &>/dev/null
