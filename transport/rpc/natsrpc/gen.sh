#!/bin/bash
set -ex

DIR="$( cd "$( dirname "$0"  )" && pwd  )"
ROOT=$DIR

 protoc \
  -I=$ROOT \
  -I=$ROOT/../../../third_party/ \
  --go_out=paths=source_relative:$ROOT \
  $ROOT/*.proto
