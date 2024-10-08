#!/bin/bash
MY_PATH=`cd $(dirname $0);pwd`
ROOT_PATH=$MY_PATH/../
go_name=go1.23.2

go_version=`go version |grep $go_name`
if [ $? -eq 0 ]; then
  return 
fi

go_tar="go1.23.2.linux-amd64.tar.gz"
go_tar_url="https://go.dev/dl/$go_tar"
go_data=~/.golang/$go_name
go_path=$go_data/gopath
go_root=$go_data/goroot
go_env_file=$go_data/goenv.sh

mkdir -p $go_path
mkdir -p $go_root

export GOROOT=$go_root
export GOPATH=$go_path
export GOBIN=$go_root/bin
[ -z $GOPROXY ] && export GOPROXY="https://proxy.golang.org,direct"

export PATH=$GOBIN:$PATH

set +e
rc=`go version | grep "$go_name"`
if [ ! $? -eq 0 -o ! -f $go_root/.decompressed ]; then
  rm -rf $go_root/.decompressed
  echo "Fetching $go_tar from $go_tar_url, stored to $go_data"
  curl -L $go_tar_url -o $go_data/$go_tar
  tar -zxvf $go_data/$go_tar --strip-components 1 -C $go_root
  touch $go_root/.decompressed
fi
set -e
