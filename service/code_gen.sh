#!/bin/bash

INPUT_FILES="*.proto"

if [[ $# -gt 1 ]]; then
    INPUT_FILES="$@"
fi

# go install github.com/gogo/protobuf/protoc-gen-gofast
# go install github.com/gogo/protobuf/protoc-gen-gogofast
# go install github.com/gogo/protobuf/protoc-gen-gogoslick

protoc -I. --gogoslick_out=plugins=grpc:. ${INPUT_FILES}
# protoc -I. --go_out=plugins=grpc:. ${INPUT_FILES}