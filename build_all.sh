#!/usr/bin/env bash
GOARCH="amd64"
GOOS_LINUX="linux"
GOOS_DARWIN="darwin"
go clean
rm -rf sjs-*-${GOARCH}
# update deps
go get -d -t -u -v
# build linux
GOOS=${GOOS_LINUX} go build -o sjs-${GOOS_LINUX}-${GOARCH}
GOOS=${GOOS_DARWIN} go build -o sjs-${GOOS_DARWIN}-${GOARCH}
