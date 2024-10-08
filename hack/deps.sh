#!/usr/bin/env bash
MY_PATH=`cd $(dirname $0);pwd`
ROOT_PATH=$MY_PATH/../
source $MY_PATH/golang-env.sh

set -o errexit
set -o nounset
set -o pipefail

cd $ROOT_PATH
go get ./...
go mod tidy
