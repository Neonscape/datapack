#!/usr/bin/env bash
set -euo pipefail

version="${1:?usage: package-linux-release.sh <version> [output-directory]}"
if [[ ! "$version" =~ ^[A-Za-z0-9][A-Za-z0-9._-]*$ ]]; then
    printf 'invalid version: use letters, digits, dots, underscores or hyphens\n' >&2
    exit 1
fi
output_directory="${2:-release}"
binary="build/bin/datapack"
archive_name="datapack-${version}-linux-amd64.tar.gz"
checksum_name="${archive_name}.sha256"
archive="${output_directory}/${archive_name}"
checksum="${output_directory}/${checksum_name}"
if [[ ! -f "$binary" ]]; then
    printf 'missing Wails binary: %s\n' "$binary" >&2
    exit 1
fi
mkdir -p -- "$output_directory"
# Staging on the destination filesystem permits no-clobber link publication.
# Neither publication follows an existing symlink or writes through a hardlink.
staging_directory="$(mktemp -d "${output_directory}/.datapack-package.XXXXXX")"
published_archive=false
published_checksum=false
completed=false
cleanup() {
    if [[ "$completed" != true ]]; then
        if [[ "$published_archive" == true && "$archive" -ef "$staging_directory/$archive_name" ]]; then rm -f -- "$archive"; fi
        if [[ "$published_checksum" == true && "$checksum" -ef "$staging_directory/$checksum_name" ]]; then rm -f -- "$checksum"; fi
    fi
    rm -rf -- "$staging_directory"
}
trap cleanup EXIT
mkdir "$staging_directory/package"
cp -p -- "$binary" "$staging_directory/package/datapack"
chmod +x "$staging_directory/package/datapack"
tar -C "$staging_directory/package" -czf "$staging_directory/$archive_name" datapack
(cd "$staging_directory" && sha256sum -- "$archive_name" > "$checksum_name")

# Rerunning an identical immutable release is safe. A different artifact or any
# link/partial pair is an explicit conflict; old artifacts are never truncated.
if [[ -e "$archive" || -L "$archive" || -e "$checksum" || -L "$checksum" ]]; then
    for target in "$archive" "$checksum"; do
        if [[ ! -f "$target" || -L "$target" || "$(stat -c %h -- "$target")" != 1 ]]; then
            printf 'release artifact conflict: %s\n' "$target" >&2; exit 1
        fi
    done
    if ! cmp -s -- "$archive" "$staging_directory/$archive_name" || ! cmp -s -- "$checksum" "$staging_directory/$checksum_name"; then
        printf 'release version already has different artifacts; choose a new version\n' >&2; exit 1
    fi
else
    ln -T -- "$staging_directory/$archive_name" "$archive"
    published_archive=true
    ln -T -- "$staging_directory/$checksum_name" "$checksum"
    published_checksum=true
fi
completed=true
printf '%s\n' "$archive" "$checksum"
