# Release automation (build, publish, install) — Design

Date: 2026-07-16

## Goal

Today `make tag` creates and pushes a semver git tag (`vX.Y.Z`) from `main`, but nothing builds or publishes a binary for it. We want tag pushes to automatically produce a multi-platform GitHub Release (binaries + checksums), similar in spirit to [Neovim's releases](https://github.com/neovim/neovim/releases), plus a one-line install script that installs a chosen version.

## Scope

- Add a `version` subcommand/flag to the CLI, with the version embedded at build time.
- Add a GitHub Actions workflow triggered on tag push (`v*`) that cross-compiles binaries for linux/darwin × amd64/arm64, and publishes them to a GitHub Release with checksums and an auto-generated changelog.
- Add a reusable `scripts/build.sh` used by both CI and local cross-compile testing.
- Add a public `install.sh` that downloads and installs a specific (or latest) release binary to `~/.local/bin`.

Out of scope (explicitly not building now):
- Windows binaries.
- Re-running/overwriting an existing release for a re-tagged version — if a release needs to be redone, the tag and release are deleted manually first.
- Package manager distribution (Homebrew, apt, etc).

## 1. Version command

- New file `cmd/version.go` in the `cmd` package.
- Package-level `var Version = "dev"` in `cmd`, overridden at build time via `-ldflags "-X kanba/cmd.Version=v1.2.0"`.
- Cobra subcommand `kanba version` prints `Version`.
- `RootCmd` also gets a `--version` flag (cobra's built-in `Version` field set to `Version`) so `kanba --version` works too.

## 2. `scripts/build.sh`

- Inputs: `GOOS`, `GOARCH`, `VERSION` (positional args or env vars).
- Runs `go build -ldflags "-X kanba/cmd.Version=$VERSION" -o dist/kanba-$VERSION-$GOOS-$GOARCH/kanba`.
- Used identically by the CI matrix and, optionally, locally to sanity-check a cross-compile before tagging.

## 3. GitHub Actions workflow (`.github/workflows/release.yml`)

- Trigger: `on: push: tags: ["v*"]`.
- **Build job**: matrix of `{linux, darwin} × {amd64, arm64}` (4 combinations), `fail-fast: false`. Each leg:
  - Checks out the tag commit.
  - Runs `scripts/build.sh` for its `GOOS`/`GOARCH` with `VERSION` = the pushed tag.
  - Packages the binary as `kanba-<version>-<os>-<arch>.tar.gz`.
  - Uploads it as a build artifact.
- **Release job**: depends on all 4 build legs succeeding (so a single-arch build failure blocks the release, not partially publishes it).
  - Downloads all build artifacts.
  - Generates `sha256sum` checksums for all tarballs.
  - Generates release notes from `git log <previous-tag>..<current-tag> --oneline --no-merges` (same source data the `make tag` message already uses).
  - Creates the GitHub Release for the tag via `gh release create`, attaching the 4 tarballs + a `checksums.txt`.

## 4. `install.sh`

Published at the repo root, run via:
```
curl -fsSL https://raw.githubusercontent.com/hxsggsz/kanb-/main/install.sh | bash
```

Behavior:
1. Detect `OS` (`linux`/`darwin`) and `ARCH` (`amd64`/`arm64`) via `uname -s` / `uname -m`. Unsupported combos (e.g. Windows, `386`) fail with an explicit error listing supported platforms.
2. Resolve the target version:
   - If `KANBA_VERSION` env var is set, use it directly (e.g. `KANBA_VERSION=1.2.0 curl ... | bash`).
   - Otherwise, query the GitHub API `repos/hxsggsz/kanb-/releases/latest` for the newest tag.
3. Verify the resolved release exists (HTTP 404 → fail with a clear "version not found" message, do not attempt a broken install).
4. Download the matching `kanba-<version>-<os>-<arch>.tar.gz` and `checksums.txt`, verify the tarball's checksum, abort on mismatch.
5. Extract the binary, `mkdir -p ~/.local/bin`, move the binary there as `~/.local/bin/kanba`, `chmod +x`.
6. If `~/.local/bin` isn't on `$PATH`, print a warning telling the user to add it.

## End-to-end flow

```
merge → main
make tag NEW_TAG=1.2.0   → creates + pushes tag v1.2.0
                          → GitHub Actions triggers on tag push
                          → builds 4 binaries + checksums (fails closed if any leg fails)
                          → publishes GitHub Release v1.2.0
user: KANBA_VERSION=1.2.0 curl .../install.sh | bash
                          → installs kanba v1.2.0 into ~/.local/bin
kanba version            → prints "v1.2.0"
```

## Error handling summary

| Case | Behavior |
|---|---|
| One matrix leg fails to build | Other legs still run (`fail-fast: false`); release job does not run since it depends on all legs |
| Tag re-pushed after a release already exists | Not automated — user deletes the tag + release manually and re-tags |
| `install.sh` given a nonexistent `KANBA_VERSION` | Fails explicitly on GitHub API 404, no silent fallback |
| `install.sh` on unsupported OS/arch | Fails explicitly, lists supported combinations |
| `~/.local/bin` missing | Created via `mkdir -p` |
| Downloaded tarball checksum mismatch | Install aborts before moving any file into place |
