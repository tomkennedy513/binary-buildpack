#!/usr/bin/env bash

set -euo pipefail

rm -rf bin

GOOS="linux" CGO_ENABLED=0 GOARCH="amd64" go build -ldflags='-s -w' -o bin/run -tags osusergo ./cmd/run/

ln -fs run bin/build
ln -fs run bin/detect