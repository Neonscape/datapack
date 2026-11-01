#!/usr/bin/env bash
set -euo pipefail

: "${GITEA_SERVER_URL:?GITEA_SERVER_URL is required}"
: "${GITEA_REPOSITORY:?GITEA_REPOSITORY is required}"
: "${GITEA_TOKEN:?GITEA_TOKEN is required}"
: "${GITEA_REF_NAME:?GITEA_REF_NAME is required}"

api="${GITEA_SERVER_URL%/}/api/v1"
repository="$GITEA_REPOSITORY"
tag="$GITEA_REF_NAME"
release_directory="${1:-release}"

request() {
    curl --fail-with-body --silent --show-error \
        --header "Authorization: token ${GITEA_TOKEN}" \
        --header "Content-Type: application/json" \
        "$@"
}

if [[ ! "$tag" =~ ^[A-Za-z0-9][A-Za-z0-9._-]*$ ]]; then
    printf 'invalid release tag\n' >&2; exit 1
fi
archive_name="datapack-${tag}-linux-amd64.tar.gz"
checksum_name="${archive_name}.sha256"
artifacts=("${release_directory}/${archive_name}" "${release_directory}/${checksum_name}")
# Validate the exact two artifacts before making any remote change.
for asset in "${artifacts[@]}"; do
    if [[ ! -f "$asset" || -L "$asset" || "$(stat -c %h -- "$asset")" != 1 ]]; then
        printf 'unsafe or missing release artifact: %s\n' "$asset" >&2; exit 1
    fi
done
release_json="$(mktemp "${TMPDIR:-/tmp}/datapack-release.XXXXXX")"
trap 'rm -f -- "$release_json"' EXIT
if request "${api}/repos/${repository}/releases/tags/${tag}" > "$release_json" 2>/dev/null; then
    release_id="$(jq -r '.id' "$release_json")"
    request --request PATCH --data "$(jq -n --arg name "$tag" --arg body "Automated release for $tag" '{name:$name,body:$body,draft:false,prerelease:false}')" \
        "${api}/repos/${repository}/releases/${release_id}" > /dev/null
else
    request --request POST --data "$(jq -n --arg tag "$tag" --arg name "$tag" --arg body "Automated release for $tag" '{tag_name:$tag,name:$name,body:$body,draft:false,prerelease:false}')" \
        "${api}/repos/${repository}/releases" > "$release_json"
    release_id="$(jq -r '.id' "$release_json")"
fi

if [[ -z "$release_id" || "$release_id" == "null" ]]; then
    printf 'Gitea release response did not contain an id\n' >&2
    exit 1
fi

# Make rerunning the same tag idempotent by removing assets from the previous run.
request "${api}/repos/${repository}/releases/${release_id}/assets" \
    | jq -r --arg archive "$archive_name" --arg checksum "$checksum_name" '.[] | select(.name == $archive or .name == $checksum) | .id' \
    | while IFS= read -r asset_id; do
        [[ -z "$asset_id" ]] && continue
        curl --fail-with-body --silent --show-error \
            --request DELETE \
            --header "Authorization: token ${GITEA_TOKEN}" \
            "${api}/repos/${repository}/releases/${release_id}/assets/${asset_id}" \
            > /dev/null
    done

for asset in "${artifacts[@]}"; do
    asset_name="$(basename "$asset")"
    curl --fail-with-body --silent --show-error \
        --request POST \
        --header "Authorization: token ${GITEA_TOKEN}" \
        --form "attachment=@${asset}" \
        "${api}/repos/${repository}/releases/${release_id}/assets?name=${asset_name}" \
        > /dev/null
done

printf 'Gitea release %s is ready\n' "$tag"
