#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

go get github.com/ugorji/go/codec@v1.2.7
go mod tidy -compat=1.19
git diff --exit-code go.*
