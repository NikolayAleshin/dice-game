#!/bin/bash

set -e

PROTO_DIR=./proto

OUTPUT_DIR=./proto/gen

mkdir -p $OUTPUT_DIR

protoc -I=$PROTO_DIR \
  --go_out=$OUTPUT_DIR --go_opt=paths=source_relative \
  --go-grpc_out=$OUTPUT_DIR --go-grpc_opt=paths=source_relative \
  $PROTO_DIR/*.proto

echo "Proto files successfully generated"