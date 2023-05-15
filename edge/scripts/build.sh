#!/usr/bin/env bash
# This script builds the application for multiple platforms

set -e

cd cmd/kronos

GIT_COMMIT="$(git rev-parse HEAD)"
GIT_DESCRIBE="$(git describe --always --dirty="-dev")"

LDFLAGS='-w '
LDFLAGS+="-X devais.it/kronos/internal/pkg/version.GitCommit=$GIT_COMMIT "
LDFLAGS+="-X devais.it/kronos/internal/pkg/version.GitDescribe=$GIT_DESCRIBE "

env GOOS=linux GOARCH=386 go build -o kronos-x86-32 -ldflags "$LDFLAGS"
env GOOS=linux GOARCH=amd64 go build -o kronos-x86-64 -ldflags "$LDFLAGS"

env CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 CC=arm-linux-gnueabihf-gcc go build -o kronos-arm32 -ldflags "$LDFLAGS"
env CGO_ENABLED=1 GOOS=linux GOARCH=arm64 GOARM=7 CC=aarch64-linux-gnu-gcc go build -o kronos-arm64 -ldflags "$LDFLAGS"

cd -
