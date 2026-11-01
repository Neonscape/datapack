#!/usr/bin/env bash
# All writes and deletions occur in this test's own temporary directory.
set -euo pipefail
repository="$(cd "$(dirname "$0")/../.." && pwd)"
fixture="$(mktemp -d "${TMPDIR:-/tmp}/datapack-package-test.XXXXXX")"
trap 'rm -rf -- "$fixture"' EXIT
cd "$fixture"
mkdir -p build/bin 'output files/package' staging
printf 'test binary\n' > build/bin/datapack
printf 'keep other releases\n' > 'output files/other.tar.gz'
printf 'keep user directory\n' > 'output files/package/unrelated.txt'

for iteration in 1 2; do
    TMPDIR="$fixture/staging" bash "$repository/scripts/package-linux-release.sh" v1.2.3 'output files' >/dev/null
    test "$(cat 'output files/other.tar.gz')" = 'keep other releases'
    test "$(cat 'output files/package/unrelated.txt')" = 'keep user directory'
    test "$(tar -xOzf 'output files/datapack-v1.2.3-linux-amd64.tar.gz' datapack)" = 'test binary'
    tar -tvzf 'output files/datapack-v1.2.3-linux-amd64.tar.gz' | awk '{if (substr($1,4,1)!="x") exit 1}'
    sha256sum --check 'output files/datapack-v1.2.3-linux-amd64.tar.gz.sha256' >/dev/null
    test -z "$(ls -A staging)"
done
if bash "$repository/scripts/package-linux-release.sh" '../escape' 'output files' >/dev/null 2>&1; then
    printf 'accepted path traversal in release version\n' >&2; exit 1
fi
mkdir -p blocked-output
printf 'not a directory\n' > blocked-output/file
if TMPDIR="$fixture/staging" bash "$repository/scripts/package-linux-release.sh" v1 blocked-output/file >/dev/null 2>&1; then
    printf 'accepted file as output directory\n' >&2; exit 1
fi
test -z "$(ls -A staging)"
printf 'packaging safety: passed\n'
