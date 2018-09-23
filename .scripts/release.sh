#! /bin/bash

# Cross-compile Temporal using gox, injecting appropriate tags.
go get -u github.com/mitchellh/gox

RELEASE=$(git describe --tags)
TARGETS="linux/amd64 linux/386 linux/arm darwin/amd64 windows/amd64"

mkdir -p release

gox -output="release/temporal-$(git describe --tags)-{{.OS}}-{{.Arch}}" \
    -ldflags "-X main.Version=$RELEASE" \
    -osarch="$TARGETS" \
    ./cmd/temporal
