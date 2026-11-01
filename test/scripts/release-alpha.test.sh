#!/usr/bin/env bash
# No network: validate the actual alpha payload and refusal paths via fake curl.
set -euo pipefail
repository="$(cd "$(dirname "$0")/../.." && pwd)"
fixture="$(mktemp -d "${TMPDIR:-/tmp}/datapack-alpha-test.XXXXXX")"
trap 'rm -rf -- "$fixture"' EXIT
cd "$fixture"
mkdir -p build/bin mock-bin tmp
printf 'fixture binary\n' > build/bin/datapack
bash "$repository/scripts/package-linux-release.sh" v0.1.0-alpha.1 output >/dev/null
cat > mock-bin/curl <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail
url="${!#}";method=GET;out="";write="";data=""
for ((i=1;i<=$#;i++)); do
    case "${!i}" in
        --request) j=$((i+1));method="${!j}";;
        --output) j=$((i+1));out="${!j}";;
        --write-out) j=$((i+1));write="${!j}";;
        --data) j=$((i+1));data="${!j}";;
        attachment=@*) value="${!i}";printf 'UPLOAD %s\n' "$(basename "${value#attachment=@}")" >> "$DATAPACK_TEST_LOG";;
    esac
done
printf '%s %s\n' "$method" "$url" >> "$DATAPACK_TEST_LOG"
if [[ "$url" == */releases/tags/* ]]; then
    printf '{"id":1}\n' > "$out"
    [[ "$write" == '%{http_code}' ]] || exit 2
    printf '%s' "$DATAPACK_TEST_STATUS"
    [[ "$DATAPACK_TEST_STATUS" == 200 ]] && exit 0 || exit 22
fi
if [[ "$method" == POST && "$url" == */releases ]]; then
    printf '%s\n' "$data" > "$DATAPACK_TEST_PAYLOAD"
    printf '{"id":1}\n'
else printf '{}\n';fi
MOCK
chmod +x mock-bin/curl
export PATH="$fixture/mock-bin:$PATH" TMPDIR="$fixture/tmp"
export GITEA_SERVER_URL=https://mock.invalid GITEA_REPOSITORY=fixture/repo GITEA_REF_NAME=v0.1.0-alpha.1 GITEA_TOKEN=fixture-only-token
export DATAPACK_TEST_LOG="$fixture/requests.log" DATAPACK_TEST_PAYLOAD="$fixture/payload.json"
export DATAPACK_TEST_STATUS=404
bash "$repository/scripts/release-gitea.sh" output >/dev/null
jq -e '.tag_name=="v0.1.0-alpha.1" and .prerelease==true and .draft==false and (.body|contains("预发行"))' "$DATAPACK_TEST_PAYLOAD" >/dev/null
! grep -Eq '^DELETE|^PATCH|private-backup' "$DATAPACK_TEST_LOG"
test "$(grep -c '^UPLOAD ' "$DATAPACK_TEST_LOG")" = 2
for status in 200 403 500 000; do
    export DATAPACK_TEST_STATUS="$status"
    : > "$DATAPACK_TEST_LOG";rm -f "$DATAPACK_TEST_PAYLOAD"
    if bash "$repository/scripts/release-gitea.sh" output >/dev/null 2>&1;then printf 'unsafe alpha lookup accepted: %s\n' "$status" >&2;exit 1;fi
    test ! -e "$DATAPACK_TEST_PAYLOAD"
    ! grep -Eq '^POST|^PATCH|^DELETE|^UPLOAD' "$DATAPACK_TEST_LOG"
done
test -z "$(ls -A tmp)"
printf 'alpha prerelease payload, asset scope, existing/error refusals: passed\n'
