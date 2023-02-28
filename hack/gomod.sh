#!/usr/bin/env bash
MY_PATH=`cd $(dirname $0);pwd`
ROOT_PATH=$MY_PATH/../
source $MY_PATH/golang-env.sh

set -o errexit
set -o nounset
set -o pipefail

cd $ROOT_PATH
GITREPO=$(shell git remote -v | grep fetch | awk '{print $$2}' | sed 's/\.git//g' | sed 's/https:\/\///g')

go mod init ${GITREPO}
go mod tidy -compat=1.19