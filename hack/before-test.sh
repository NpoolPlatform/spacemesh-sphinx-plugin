#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

PLATFORM=linux/amd64
OUTPUT=./output

mkdir -p $OUTPUT/$PLATFORM
for service_name in `ls $(pwd)/cmd`; do
    cd $OUTPUT/$PLATFORM; ./$service_name run > /dev/null 2>&1 &
done
