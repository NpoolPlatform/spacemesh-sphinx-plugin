#!/usr/bin/env bash
MY_PATH=`cd $(dirname $0);pwd`
source $MY_PATH/golang-env.sh

set -o errexit
set -o nounset
set -o pipefail

go mod tidy -compat=1.23.2
git diff --exit-code go.*
