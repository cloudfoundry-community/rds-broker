#!/bin/sh

set -e -x

export GOPATH=$(pwd)/gopath
export PATH=$PATH:$GOPATH/bin

go get github.com/tools/godep

cd gopath/src/github.com/18F/aws-broker

cp secrets-test.yml secrets.yml
cp catalog-test.yml catalog.yml

godep get
godep go build ./...
godep go test ./...
