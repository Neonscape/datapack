# Gitea Actions

The repository contains a Linux CI and release workflow at `.gitea/workflows/ci.yml`.

## Workflow behavior

- Every branch push runs Go tests, Go race tests, frontend typechecking, Playwright e2e tests, and a Linux Wails production build.
- A tag matching `v*.*.*` also packages `datapack-vX.Y.Z-linux-amd64.tar.gz` and its SHA-256 file.
- A matching tag creates or updates a Gitea release and uploads both files as release assets.
- Re-running the same tag is idempotent: assets from the previous run are removed before the new assets are uploaded.

## Required Gitea setup

1. Enable Actions for the repository in **Settings > Actions > General**.
2. Register a trusted Linux Gitea Runner. The runner needs Docker if jobs use container execution.
3. Ensure the runner image or host has Go 1.25+, Node.js 20+, npm, Wails CLI 2.13, GTK 3, and WebKitGTK 4.1 development packages. The workflow checks the toolchain and installs the Debian/Ubuntu system packages.
4. Ensure the repository Actions token maximum permissions allow `code: read` and `releases: write`.

The workflow uses the built-in `${{ secrets.GITEA_TOKEN }}`. You normally do not need to create this secret manually: Gitea supplies it to Actions jobs. If the instance disables release write access for the built-in token, create a repository secret named `GITEA_TOKEN` containing a Gitea personal access token with repository release write permission, or adjust the repository/owner Actions token permissions.

Do not commit a personal access token to the repository or workflow file.

The workflow does not use `setup-go` or `setup-node`: those actions download toolchains and can fail before useful diagnostics are available when the Runner service has an invalid proxy environment. Instead, the Runner host must provide the pinned toolchain and the workflow prints its versions. Configure the proxy on the Runner service itself, not only in your interactive shell. The Gitea hostname is always placed in `NO_PROXY` so checkout and release API calls go directly to the local instance.

If the Runner is managed by systemd, inspect the service environment and restart it after changes. A proxy configured as `127.0.0.1:7897` is valid only if the proxy is listening in the same network namespace as the Runner process. Verify it as the Runner service user, not only from your desktop shell.

The workflow also injects an empty Git `http.proxy` configuration with Git's highest-priority configuration environment, because a proxy stored in the Runner's global or system Git config is not affected by `NO_PROXY`. For a permanent runner fix, inspect the runner host and remove stale proxy settings from every Git config layer:

```bash
git config --show-origin --get-regexp '(^|\.)https?\.proxy$'
git config --global --unset-all http.proxy || true
git config --global --unset-all https.proxy || true
```

If the output identifies `/etc/gitconfig`, remove the setting there as root or update the runner's container image/environment. The workflow override is intentionally kept so checkout remains local even if the runner service later inherits the old setting again.

## Release tag

Push a semantic-version tag to publish a release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The first workflow run can be used to verify that the runner has the required Wails, GTK/WebKitGTK, Node, and Playwright dependencies. A failed build prevents release creation or leaves no valid release assets.

## Local packaging

After a successful Linux build, the release packaging script can be tested locally:

```bash
./scripts/build.sh -clean
./scripts/package-linux-release.sh v0.1.0
```

The generated files are placed in `release/`.
