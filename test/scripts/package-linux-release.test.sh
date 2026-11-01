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
    (cd 'output files' && sha256sum --check 'datapack-v1.2.3-linux-amd64.tar.gz.sha256' >/dev/null)
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

for kind in symlink hardlink; do
    for target in datapack-vlink-linux-amd64.tar.gz datapack-vlink-linux-amd64.tar.gz.sha256; do
        directory="links-$kind-$target"
        mkdir "$directory"
        printf 'external sentinel\n' > sentinel
        if [[ "$kind" == symlink ]]; then ln -s "$fixture/sentinel" "$directory/$target"; else ln sentinel "$directory/$target"; fi
        if bash "$repository/scripts/package-linux-release.sh" vlink "$directory" >/dev/null 2>&1; then printf 'accepted link artifact\n' >&2; exit 1; fi
        test "$(cat sentinel)" = 'external sentinel'
        test "$(cat "$directory/$target")" = 'external sentinel'
    done
done
mkdir old-artifacts
printf 'previous archive\n' > old-artifacts/datapack-vold-linux-amd64.tar.gz
printf 'previous checksum\n' > old-artifacts/datapack-vold-linux-amd64.tar.gz.sha256
if bash "$repository/scripts/package-linux-release.sh" vold old-artifacts >/dev/null 2>&1; then printf 'overwrote old version\n' >&2; exit 1; fi
test "$(cat old-artifacts/datapack-vold-linux-amd64.tar.gz)" = 'previous archive'
test "$(cat old-artifacts/datapack-vold-linux-amd64.tar.gz.sha256)" = 'previous checksum'
mkdir mock-bin
cat > mock-bin/ln <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail
count=0
[[ ! -f "$DATAPACK_TEST_COUNTER" ]] || count="$(cat "$DATAPACK_TEST_COUNTER")"
count=$((count+1)); printf '%s' "$count" > "$DATAPACK_TEST_COUNTER"
[[ "$count" != 2 ]] || exit 1
exec /usr/bin/ln "$@"
MOCK
chmod +x mock-bin/ln
if DATAPACK_TEST_COUNTER="$fixture/links-count" PATH="$fixture/mock-bin:$PATH" bash "$repository/scripts/package-linux-release.sh" vfail failed-publication >/dev/null 2>&1; then printf 'injected publication did not fail\n' >&2; exit 1; fi
test -z "$(ls -A failed-publication)"
printf 'artifact links/conflicts/failed publication: passed\n'
