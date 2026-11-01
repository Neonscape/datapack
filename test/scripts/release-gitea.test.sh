#!/usr/bin/env bash
# Contract test: curl is replaced by a fixture executable, never real network I/O.
set -euo pipefail
repository="$(cd "$(dirname "$0")/../.." && pwd)"
fixture="$(mktemp -d "${TMPDIR:-/tmp}/datapack-release-test.XXXXXX")"
trap 'rm -rf -- "$fixture"' EXIT
cd "$fixture"
mkdir -p build/bin mock-bin tmp
printf 'binary\n' > build/bin/datapack
bash "$repository/scripts/package-linux-release.sh" vtest output >/dev/null
printf 'private bytes\n' > output/private-backup.tar.gz
printf 'old release\n' > output/datapack-vold-linux-amd64.tar.gz
printf 'unrelated hash\n' > output/private.sha256
printf 'external sentinel\n' > sentinel
ln -s "$fixture/sentinel" output/release.json
cat > mock-bin/curl <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail
url="${!#}"
method=GET
for ((i=1;i<=$#;i++)); do
    if [[ "${!i}" == --request ]]; then j=$((i+1)); method="${!j}"; fi
    if [[ "${!i}" == --data ]]; then j=$((i+1)); printf '%s\n' "${!j}" > "$DATAPACK_TEST_PAYLOAD"; fi
    if [[ "${!i}" == attachment=@* ]]; then value="${!i}"; printf 'UPLOAD %s\n' "$(basename "${value#attachment=@}")" >> "$DATAPACK_TEST_LOG"; fi
done
if [[ "$method" == DELETE ]]; then printf 'DELETE %s\n' "$url" >> "$DATAPACK_TEST_LOG"; fi
case "$url" in
    */releases/tags/*) printf '{"id":1}\n';;
    */releases/1/assets) printf '[{"id":3,"name":"datapack-vtest-linux-amd64.tar.gz"},{"id":4,"name":"datapack-vtest-linux-amd64.tar.gz.sha256"},{"id":99,"name":"private-backup.tar.gz"}]\n';;
    *) printf '{}\n';;
esac
MOCK
chmod +x mock-bin/curl
export PATH="$fixture/mock-bin:$PATH" TMPDIR="$fixture/tmp"
export GITEA_SERVER_URL=https://mock.invalid GITEA_REPOSITORY=fixture/repo GITEA_REF_NAME=vtest GITEA_TOKEN=fixture-only-token
export DATAPACK_TEST_LOG="$fixture/requests.log" DATAPACK_TEST_PAYLOAD="$fixture/payload.json"
bash "$repository/scripts/release-gitea.sh" output >/dev/null
jq -e '.name=="vtest" and .prerelease==false and .draft==false' "$DATAPACK_TEST_PAYLOAD" >/dev/null
sort "$DATAPACK_TEST_LOG" > actual
cat > expected <<'EXPECTED'
DELETE https://mock.invalid/api/v1/repos/fixture/repo/releases/1/assets/3
DELETE https://mock.invalid/api/v1/repos/fixture/repo/releases/1/assets/4
UPLOAD datapack-vtest-linux-amd64.tar.gz
UPLOAD datapack-vtest-linux-amd64.tar.gz.sha256
EXPECTED
cmp actual expected
test "$(cat sentinel)" = 'external sentinel'
test -z "$(ls -A tmp)"
rm output/datapack-vtest-linux-amd64.tar.gz
ln -s "$fixture/sentinel" output/datapack-vtest-linux-amd64.tar.gz
: > "$DATAPACK_TEST_LOG"
if bash "$repository/scripts/release-gitea.sh" output >/dev/null 2>&1; then printf 'accepted unsafe release artifact\n' >&2; exit 1; fi
test ! -s "$DATAPACK_TEST_LOG"
printf 'release artifact scope (mock network only): passed\n'
