# Datapack Editor

Wails 2 + Go + React/TypeScript desktop editor for CUE-backed game datapacks.

## Development Environment

Installed/expected toolchain:

- Go 1.26.5, installed under `$HOME/.local/opt/go-1.26.5`
- Wails CLI v2.13.0, installed under `$HOME/go/bin/wails`
- CUE CLI v0.17.1, installed under `$HOME/go/bin/cue`
- Node 20 / npm
- Debian Wails dependencies: `build-essential`, `pkg-config`, `libgtk-3-dev`, `libwebkit2gtk-4.1-dev`

This Debian setup uses WebKitGTK 4.1, so Wails commands need the `webkit2_41` build tag.

## Commands

```bash
./scripts/dev.sh
./scripts/build.sh
./scripts/cue.sh version
./scripts/cue.sh vet ./...
go test ./...
npm --prefix frontend run build
```

`tools.go` pins the CUE CLI and Go API as development dependencies. The
`scripts/cue.sh` wrapper installs the pinned CLI into `$GOPATH/bin` if it is
missing, so local commands and CI use the same tool version.

## Project Plan

The implementation plan draft is preserved at `docs/implementation-draft.md`.
