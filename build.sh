#!/bin/bash

set -exo pipefail

env CGO_ENABLED="1" CGO_CFLAGS="-O3 -g -march=native" CGO_CXXFLAGS="-O3 -g -march=native" CC=clang CXX=clang++ GOOS=darwin GOARCH=arm64 go build -o ./bin/$(basename $PWD)_darwin_arm64 ./main.go
