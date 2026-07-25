#!/usr/bin/env bash
set -euo pipefail

cue_bin="$(go env GOPATH)/bin/cue"

if [[ ! -x "$cue_bin" ]]; then
    go install cuelang.org/go/cmd/cue@v0.17.1
fi

exec "$cue_bin" "$@"
