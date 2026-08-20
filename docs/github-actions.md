# GitHub Actions

GitHub CI is defined in `.github/workflows/ci.yml`.

## Workflow behavior

- Every push and pull request runs the complete Linux test suite: Go tests, race tests, frontend typechecking, and Playwright e2e tests.
- Every push and pull request builds native Linux and Windows Wails binaries after the Linux test job passes.
- A tag matching `v*.*.*` packages both platform binaries, uploads temporary Actions artifacts, creates a GitHub Release, and attaches all packages and SHA-256 files.

Release assets use these names:

```text
datapack-vX.Y.Z-linux-amd64.tar.gz
datapack-vX.Y.Z-linux-amd64.tar.gz.sha256
datapack-vX.Y.Z-windows-amd64.zip
datapack-vX.Y.Z-windows-amd64.zip.sha256
```

## Required GitHub setup

No personal access token is required for the standard workflow. The release job uses GitHub's built-in `GITHUB_TOKEN` and requests only:

```yaml
permissions:
  contents: write
```

If repository or organization Actions settings restrict the token to read-only, allow workflow write permissions before pushing a release tag.

GitHub-hosted runners provide Linux and Windows environments. Linux builds install GTK 3 and WebKitGTK 4.1 development packages; Windows builds use native Wails packaging without the Linux `webkit2_41` tag.

## Release tag

```bash
git tag v0.1.0
git push github v0.1.0
```

The release job runs only after both platform builds finish successfully.
