#!/usr/bin/env bash
MY_PATH=`cd $(dirname $0);pwd`
source $MY_PATH/golang-env.sh

set -o errexit
set -o nounset
set -o pipefail

go get github.com/ugorji/go/codec@v1.2.7
go mod tidy -compat=1.23
git diff --exit-code go.*
