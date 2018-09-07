#! /bin/bash

# Cross-compile Temporal using xgo, injecting appropriate tags.

RELEASE=$(git describe --tags)
TARGETS="linux/amd64,linux/386,linux/arm,darwin/amd64,windows/amd64"

mkdir -p release

# dest:    output folder
# out:     output file name
# ldflags: set build version (git tag)
# targets: platforms to compile for (see $TARGET)
# deps:    for ethereum, see https://github.com/ethereum/go-ethereum/wiki/Cross-compiling-Ethereum#building-ethereum
xgo \
  --dest="release" \
  --out="temporal-$(git describe --tags)" \
  --ldflags="-X main.Version=$RELEASE" \
  --targets="$TARGETS" \
  --deps="https://gmplib.org/download/gmp/gmp-6.0.0a.tar.bz2" \
  ./cmd/temporal
