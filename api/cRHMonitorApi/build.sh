#!/bin/bash
set -e

cd $(dirname $0)
CURR_DIR=$(pwd)

BUILD_DIR="${CURR_DIR}/build"
LIB_DIR="${CURR_DIR}/../libs"

[[ -d "${BUILD_DIR}" ]] && rm -rf "${BUILD_DIR}"

mkdir -p "${BUILD_DIR}" && cd "${BUILD_DIR}"

cmake "$*" ..

make clean && make
