#!/bin/bash

set -ex
CURDIR=$(cd $(dirname $0); pwd)
PROTO_IMPORT=$CURDIR/../../../../third_party
NATS_IMPORT=$CURDIR/../../../../transport/rpc

protoc -I$PROTO_IMPORT -I$NATS_IMPORT --proto_path=$CURDIR/ \
  --go_out=$CURDIR/ --go_opt=paths=source_relative \
  --go-grpc_out=$CURDIR/ --go-grpc_opt=paths=source_relative \
  --natsrpc_out=$CURDIR/ --natsrpc_opt=paths=source_relative \
  --http_out=$CURDIR/ --http_opt=paths=source_relative \
  --kafka_out=$CURDIR/ --kafka_opt=paths=source_relative \
  $CURDIR/*.proto