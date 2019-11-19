#! /bin/bash

# setup install variabels
GOVERSION=$(go version | awk '{print $3}' | tr -d "go" | awk -F "." '{print $2}')
WORKDIR="/tmp/temporal-workdir"
# handle golagn version detection
if [[ "$GOVERSION" -lt 11 ]]; then
    echo "[ERROR] golang is less than 1.11 and will produce errors"
    exit 1
fi
if [[ "$GOVERSION" -lt 12 ]]; then
    echo "[WARN] detected golang version is less than 1.12 and may produce errors"
fi
# create working directory
mkdir "$WORKDIR"
cd "$WORKDIR"
# download temporal
git clone https://github.com/RTradeLtd/Temporal.git
cd Temporal
# initialize submodules, and download all dependencies
make setup
# make cli binary
make install