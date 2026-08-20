#!/usr/bin/env bash
set -euo pipefail

version="${1:?usage: package-linux-release.sh <version> [output-directory]}"
output_directory="${2:-release}"
binary="build/bin/datapack"
archive="${output_directory}/datapack-${version}-linux-amd64.tar.gz"

if [[ ! -f "$binary" ]]; then
    printf 'missing Wails binary: %s\n' "$binary" >&2
    exit 1
fi

rm -rf "$output_directory"
mkdir -p "$output_directory/package"
cp "$binary" "$output_directory/package/datapack"
chmod +x "$output_directory/package/datapack"
tar -C "$output_directory/package" -czf "$archive" datapack
sha256sum "$archive" > "${archive}.sha256"
rm -rf "$output_directory/package"

printf '%s\n' "$archive" "${archive}.sha256"
