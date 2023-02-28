#!/bin/bash

go_name=go1.19.6

set +e
rc=`go version | grep $go_name`
if [ $? -eq 0 ]; then
    exit 0
fi
set -e

echo "Will change go version to $go_name"

go_tar="go1.19.6.linux-amd64.tar.gz"
go_tar_url="https://go.dev/dl/$go_tar"
go_data=/tmp/$go_name
go_path=$go_data/gopath
go_root=$go_data/goroot
go_env_file=$go_data/goenv.sh

mkdir -p $go_path
mkdir -p $go_root

if [ ! -f "$go_data/$go_tar" ];then
    wget $go_tar_url -P $go_data
    tar -zxvf $go_data/$go_tar --strip-components 1 -C $go_root
fi


export GOROOT=$go_root
export GOPATH=$go_path
export GOBIN=$go_root/bin
export GOPROXY="https://proxy.golang.org,direct"

shopt -s expand_aliases
alias go="$go_root/bin/go"