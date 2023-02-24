#!/bin/bash

INPUT_FILES="*.proto"

if [[ $# -gt 1 ]]; then
    INPUT_FILES="$@"
fi

protoc -I. --gogoslick_out=plugins=grpc:. ${INPUT_FILES}