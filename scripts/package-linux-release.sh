#!/usr/bin/env bash
set -euo pipefail

version="${1:?usage: package-linux-release.sh <version> [output-directory]}"
if [[ ! "$version" =~ ^[A-Za-z0-9][A-Za-z0-9._-]*$ ]]; then
    printf 'invalid version: use letters, digits, dots, underscores or hyphens\n' >&2
    exit 1
fi
output_directory="${2:-release}"
binary="build/bin/datapack"
archive="${output_directory}/datapack-${version}-linux-amd64.tar.gz"

if [[ ! -f "$binary" ]]; then
    printf 'missing Wails binary: %s\n' "$binary" >&2
    exit 1
fi

# Only this invocation's temporary staging directory belongs to the packager.
# The destination may already contain other releases or unrelated user files.
staging_directory="$(mktemp -d "${TMPDIR:-/tmp}/datapack-package.XXXXXX")"
trap 'rm -rf -- "$staging_directory"' EXIT
mkdir -p -- "$output_directory"
cp -- "$binary" "$staging_directory/datapack"
chmod +x "$staging_directory/datapack"
tar -C "$staging_directory" -czf "$archive" datapack
sha256sum -- "$archive" > "${archive}.sha256"

printf '%s\n' "$archive" "${archive}.sha256"
