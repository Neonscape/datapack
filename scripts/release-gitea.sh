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

release_json="${release_directory}/release.json"
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
    | jq -r '.[].id' \
    | while IFS= read -r asset_id; do
        [[ -z "$asset_id" ]] && continue
        curl --fail-with-body --silent --show-error \
            --request DELETE \
            --header "Authorization: token ${GITEA_TOKEN}" \
            "${api}/repos/${repository}/releases/${release_id}/assets/${asset_id}" \
            > /dev/null
    done

while IFS= read -r asset; do
    [[ -z "$asset" ]] && continue
    asset_name="$(basename "$asset")"
    curl --fail-with-body --silent --show-error \
        --request POST \
        --header "Authorization: token ${GITEA_TOKEN}" \
        --form "attachment=@${asset}" \
        "${api}/repos/${repository}/releases/${release_id}/assets?name=${asset_name}" \
        > /dev/null
done < <(find "$release_directory" -maxdepth 1 -type f \( -name '*.tar.gz' -o -name '*.sha256' \) -print | sort)

printf 'Gitea release %s is ready\n' "$tag"
