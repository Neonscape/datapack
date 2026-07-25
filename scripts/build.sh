#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
export PATH="$HOME/.local/bin:$HOME/go/bin:/usr/bin:/bin:$PATH"
exec wails build -tags webkit2_41 "$@"
