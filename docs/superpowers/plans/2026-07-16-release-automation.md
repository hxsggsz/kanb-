# Release Automation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** On `make tag`, a GitHub Actions workflow builds linux/darwin (amd64/arm64) binaries, publishes them to a GitHub Release with checksums, and a public `install.sh` lets anyone install a specific (or latest) tagged version into `~/.local/bin`.

**Architecture:** A `kanba version` cobra command backed by a build-time-injected `cmd.Version` var. A reusable `scripts/build.sh` cross-compiles a single `GOOS`/`GOARCH` binary with that var set via `-ldflags`. A GitHub Actions workflow (`.github/workflows/release.yml`) triggers on `v*` tag pushes, runs `scripts/build.sh` across a 4-way matrix, then a dependent release job packages, checksums, and publishes everything as a GitHub Release. A standalone `install.sh` at the repo root resolves a version (env var or GitHub API "latest"), downloads the matching tarball, verifies its checksum, and installs the binary.

**Tech Stack:** Go 1.25 (cobra), bash, GitHub Actions, GitHub CLI (`gh`), `curl`/`sha256sum`/`tar`.

## Global Constraints

- Target platforms: linux and darwin only, amd64 and arm64 each (4 binaries total). No Windows.
- Version tags follow `vX.Y.Z` (already enforced by the existing `make tag` target).
- Version string embedded in the binary via `-ldflags "-X kanba/cmd.Version=<version>"`, default `"dev"` when unset.
- Default install directory: `~/.local/bin`, overridable via env var.
- A release is published only if all 4 build matrix legs succeed (`fail-fast: false`, release job depends on all legs).
- `install.sh` must fail loudly (non-zero exit, clear message) on: unsupported OS/arch, nonexistent requested version, checksum mismatch. Never install silently on any of these.
- No Windows support, no re-release automation for re-tagged versions, no package manager distribution — explicitly out of scope per the spec.

---

## File Structure

- `cmd/version.go` — new file: `Version` var + `kanba version` subcommand; wires `RootCmd.Version`.
- `cmd/version_test.go` — new file: tests for the version command's output.
- `scripts/build.sh` — new file: cross-compiles one `GOOS`/`GOARCH` binary with version injected, packages a tarball.
- `scripts/build_test.sh` — new file: shell-based smoke test invoking `build.sh` and asserting on its output artifact and embedded version.
- `.github/workflows/release.yml` — new file: tag-triggered build matrix + release job.
- `install.sh` — new file at repo root: detects platform, resolves version, downloads, verifies, installs.
- `Makefile` — modify: no changes to `tag` target's behavior (out of scope), but no new targets needed since `scripts/build.sh` is called directly.

---

### Task 1: `kanba version` command

**Files:**
- Create: `cmd/version.go`
- Test: `cmd/version_test.go`

**Interfaces:**
- Produces: `cmd.Version` (package-level `var Version = "dev"`, `string` type) — consumed by `scripts/build.sh` via `-ldflags -X kanba/cmd.Version=...` in Task 2, and read at runtime by the `kanba version` subcommand and `kanba --version`.
- Produces: `cmd.VersionCmd` (`*cobra.Command`), registered on `cmd.RootCmd` via `init()`.

- [ ] **Step 1: Write the failing test**

Create `cmd/version_test.go`:

```go
package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommandPrintsVersion(t *testing.T) {
	Version = "v9.9.9"
	defer func() { Version = "dev" }()

	buf := &bytes.Buffer{}
	RootCmd.SetOut(buf)
	RootCmd.SetArgs([]string{"version"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "v9.9.9") {
		t.Errorf("expected output to contain version, got %q", buf.String())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./cmd/... -run TestVersionCommand -v`
Expected: FAIL — `Version` undefined (compile error), since `cmd/version.go` doesn't exist yet.

- [ ] **Step 3: Write minimal implementation**

Create `cmd/version.go`:

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the CLI version, injected at build time via:
// -ldflags "-X kanba/cmd.Version=vX.Y.Z"
var Version = "dev"

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the kanba version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(cmd.OutOrStdout(), Version)
		return nil
	},
}

func init() {
	RootCmd.Version = Version
	RootCmd.AddCommand(VersionCmd)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./cmd/... -run TestVersionCommand -v`
Expected: PASS

Note: the real check that `-ldflags` correctly sets `RootCmd.Version` at actual build time happens in Task 2's build script test, which executes a real cross-compiled binary and asserts on its printed version.

- [ ] **Step 5: Run the full test suite to check for regressions**

Run: `go test ./... -v`
Expected: All tests PASS (existing tests untouched)

- [ ] **Step 6: Commit**

```bash
git add cmd/version.go cmd/version_test.go
git commit -m "feat: add kanba version command"
```

---

### Task 2: `scripts/build.sh` cross-compile script

**Files:**
- Create: `scripts/build.sh`
- Test: `scripts/build_test.sh`

**Interfaces:**
- Consumes: `cmd.Version` (from Task 1) as the `-ldflags -X` target.
- Produces: invocation contract `scripts/build.sh <goos> <goarch> <version>`, writing a binary to `dist/kanba-<version>-<goos>-<goarch>/kanba` and a tarball `dist/kanba-<version>-<goos>-<goarch>.tar.gz` — consumed by the GitHub Actions workflow in Task 3 (same two artifact paths).

- [ ] **Step 1: Write the failing test**

Create `scripts/build_test.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
rm -rf dist

./scripts/build.sh linux amd64 v9.9.9

BIN_DIR="dist/kanba-v9.9.9-linux-amd64"
TARBALL="dist/kanba-v9.9.9-linux-amd64.tar.gz"

if [ ! -f "$BIN_DIR/kanba" ]; then
  echo "FAIL: expected binary at $BIN_DIR/kanba"
  exit 1
fi

if [ ! -f "$TARBALL" ]; then
  echo "FAIL: expected tarball at $TARBALL"
  exit 1
fi

ACTUAL_VERSION=$("$BIN_DIR/kanba" version)
if [ "$ACTUAL_VERSION" != "v9.9.9" ]; then
  echo "FAIL: expected version v9.9.9, got $ACTUAL_VERSION"
  exit 1
fi

echo "PASS: build.sh produced a working v9.9.9 linux/amd64 binary"
rm -rf dist
```

- [ ] **Step 2: Run test to verify it fails**

Run: `chmod +x scripts/build_test.sh && ./scripts/build_test.sh`
Expected: FAIL — `scripts/build.sh: No such file or directory`

- [ ] **Step 3: Write minimal implementation**

Create `scripts/build.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

if [ $# -ne 3 ]; then
  echo "Usage: $0 <goos> <goarch> <version>" >&2
  echo "Example: $0 linux amd64 v1.2.0" >&2
  exit 1
fi

GOOS="$1"
GOARCH="$2"
VERSION="$3"

cd "$(dirname "$0")/.."

OUT_DIR="dist/kanba-${VERSION}-${GOOS}-${GOARCH}"
mkdir -p "$OUT_DIR"

echo "Building kanba ${VERSION} for ${GOOS}/${GOARCH}..."
GOOS="$GOOS" GOARCH="$GOARCH" go build \
  -ldflags "-X kanba/cmd.Version=${VERSION}" \
  -o "${OUT_DIR}/kanba" \
  .

tar -czf "dist/kanba-${VERSION}-${GOOS}-${GOARCH}.tar.gz" -C "dist" "kanba-${VERSION}-${GOOS}-${GOARCH}"

echo "Built ${OUT_DIR}/kanba and dist/kanba-${VERSION}-${GOOS}-${GOARCH}.tar.gz"
```

- [ ] **Step 4: Run test to verify it passes**

Run: `chmod +x scripts/build.sh && ./scripts/build_test.sh`
Expected: `PASS: build.sh produced a working v9.9.9 linux/amd64 binary`

(This only cross-compiles for the host's own `GOOS`/`GOARCH` reliably in the test since it executes the resulting binary — the test above targets `linux/amd64` which matches CI/dev runners; for other combinations in the matrix, only `go build` succeeding is verified, not execution, which is what Task 3's CI matrix does.)

- [ ] **Step 5: Commit**

```bash
git add scripts/build.sh scripts/build_test.sh
git commit -m "feat: add cross-compile build script"
```

---

### Task 3: GitHub Actions release workflow

**Files:**
- Create: `.github/workflows/release.yml`

**Interfaces:**
- Consumes: `scripts/build.sh <goos> <goarch> <version>` (Task 2) — invoked once per matrix leg with `version` = `${{ github.ref_name }}`.
- Consumes: artifact paths `dist/kanba-<version>-<goos>-<goarch>.tar.gz` (Task 2's output contract).
- Produces: a GitHub Release on the pushed tag with 4 tarballs + `checksums.txt` attached — consumed by `install.sh` in Task 4 (same file naming pattern and a `checksums.txt` listing `sha256  filename` per line).

There is no local unit test for a GitHub Actions workflow file. Validation here is: YAML validity, then an end-to-end dry run against a real throwaway tag pushed to the repo (documented as the verification step).

- [ ] **Step 1: Write the workflow file**

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  build:
    strategy:
      fail-fast: false
      matrix:
        include:
          - goos: linux
            goarch: amd64
          - goos: linux
            goarch: arm64
          - goos: darwin
            goarch: amd64
          - goos: darwin
            goarch: arm64
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "1.25"

      - name: Build
        run: ./scripts/build.sh ${{ matrix.goos }} ${{ matrix.goarch }} ${{ github.ref_name }}

      - uses: actions/upload-artifact@v4
        with:
          name: kanba-${{ github.ref_name }}-${{ matrix.goos }}-${{ matrix.goarch }}
          path: dist/kanba-${{ github.ref_name }}-${{ matrix.goos }}-${{ matrix.goarch }}.tar.gz

  release:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/download-artifact@v4
        with:
          path: dist
          merge-multiple: true

      - name: Generate checksums
        working-directory: dist
        run: sha256sum *.tar.gz > checksums.txt

      - name: Generate release notes
        id: notes
        run: |
          PREV_TAG=$(git describe --tags --abbrev=0 "${{ github.ref_name }}^" 2>/dev/null || echo "")
          if [ -n "$PREV_TAG" ]; then
            git log "${PREV_TAG}..${{ github.ref_name }}" --oneline --no-merges --pretty=format:"- %s" > notes.md
          else
            git log "${{ github.ref_name }}" --oneline --no-merges --pretty=format:"- %s" > notes.md
          fi

      - name: Create release
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          gh release create "${{ github.ref_name }}" \
            dist/*.tar.gz dist/checksums.txt \
            --title "${{ github.ref_name }}" \
            --notes-file notes.md
```

- [ ] **Step 2: Validate YAML syntax locally**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/release.yml'))" && echo "YAML valid"`
Expected: `YAML valid`

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "feat: add tag-triggered release workflow"
```

- [ ] **Step 4: End-to-end verification (manual, after this task is merged to main)**

This step can't run in CI-of-CI — it requires an actual tag push. Perform it once after merging:

```bash
git checkout main && git pull
make tag NEW_TAG=0.0.1-test
```

Then check:
- `gh run list --workflow=release.yml` shows a run for the `v0.0.1-test` tag, status `completed`/`success`.
- `gh release view v0.0.1-test` shows 4 `.tar.gz` assets + `checksums.txt`.

Clean up the test tag/release afterward:
```bash
gh release delete v0.0.1-test --yes
git push origin :refs/tags/v0.0.1-test
git tag -d v0.0.1-test
```

---

### Task 4: `install.sh`

**Files:**
- Create: `install.sh` (repo root)
- Test: `scripts/install_test.sh`

**Interfaces:**
- Consumes: GitHub Releases API (`https://api.github.com/repos/hxsggsz/kanb-/releases/latest` and `/releases/tags/<version>`) and release asset URLs following the `kanba-<version>-<os>-<arch>.tar.gz` / `checksums.txt` naming from Task 3.
- Consumes: env vars `KANBA_VERSION` (optional, defaults to latest release) and `KANBA_INSTALL_DIR` (optional, defaults to `~/.local/bin`).
- Produces: installed executable at `$KANBA_INSTALL_DIR/kanba` (default `~/.local/bin/kanba`).

Since this script hits the real GitHub API and downloads real release assets, its test runs against the actual repo's releases and is written to be safe to run repeatedly (installs into a temp dir, never touches the real `~/.local/bin`). This test can only pass once Task 3 has produced at least one real release — see Step 2's note.

- [ ] **Step 1: Write the failing test**

Create `scripts/install_test.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

TMP_INSTALL_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_INSTALL_DIR"' EXIT

KANBA_INSTALL_DIR="$TMP_INSTALL_DIR" ./install.sh

if [ ! -x "$TMP_INSTALL_DIR/kanba" ]; then
  echo "FAIL: expected executable at $TMP_INSTALL_DIR/kanba"
  exit 1
fi

"$TMP_INSTALL_DIR/kanba" version

echo "PASS: install.sh installed a working kanba binary into a custom dir"
```

- [ ] **Step 2: Run test to verify it fails**

Run: `chmod +x scripts/install_test.sh && ./scripts/install_test.sh`
Expected: FAIL — `install.sh: No such file or directory`

Note: this test also requires at least one real GitHub Release to exist (produced by Task 3's workflow having run at least once) to pass once `install.sh` is implemented. If no release exists yet when running this locally, use `KANBA_VERSION=v0.0.1-test` pointed at the throwaway tag from Task 3 Step 4, or wait until the first real tag is cut.

- [ ] **Step 3: Write minimal implementation**

Create `install.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

REPO="hxsggsz/kanb-"
INSTALL_DIR="${KANBA_INSTALL_DIR:-$HOME/.local/bin}"

detect_platform() {
  local os arch
  os=$(uname -s)
  arch=$(uname -m)

  case "$os" in
    Linux) os="linux" ;;
    Darwin) os="darwin" ;;
    *)
      echo "Error: unsupported OS '$os'. Supported: Linux, Darwin." >&2
      exit 1
      ;;
  esac

  case "$arch" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *)
      echo "Error: unsupported architecture '$arch'. Supported: amd64, arm64." >&2
      exit 1
      ;;
  esac

  echo "${os} ${arch}"
}

resolve_version() {
  if [ -n "${KANBA_VERSION:-}" ]; then
    echo "$KANBA_VERSION"
    return
  fi

  local latest
  latest=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')

  if [ -z "$latest" ]; then
    echo "Error: could not resolve the latest release from GitHub API." >&2
    exit 1
  fi

  echo "$latest"
}

main() {
  read -r OS ARCH <<< "$(detect_platform)"
  VERSION=$(resolve_version)

  local asset="kanba-${VERSION}-${OS}-${ARCH}.tar.gz"
  local base_url="https://github.com/${REPO}/releases/download/${VERSION}"
  local tmp_dir
  tmp_dir=$(mktemp -d)
  trap 'rm -rf "$tmp_dir"' EXIT

  echo "Installing kanba ${VERSION} (${OS}/${ARCH})..."

  if ! curl -fsSL -o "${tmp_dir}/${asset}" "${base_url}/${asset}"; then
    echo "Error: release asset not found: ${base_url}/${asset}" >&2
    echo "Check that version '${VERSION}' exists: https://github.com/${REPO}/releases" >&2
    exit 1
  fi

  curl -fsSL -o "${tmp_dir}/checksums.txt" "${base_url}/checksums.txt"

  echo "Verifying checksum..."
  (cd "$tmp_dir" && grep "$asset" checksums.txt | sha256sum -c -) || {
    echo "Error: checksum verification failed for ${asset}. Aborting install." >&2
    exit 1
  }

  tar -xzf "${tmp_dir}/${asset}" -C "$tmp_dir"

  mkdir -p "$INSTALL_DIR"
  mv "${tmp_dir}/kanba-${VERSION}-${OS}-${ARCH}/kanba" "${INSTALL_DIR}/kanba"
  chmod +x "${INSTALL_DIR}/kanba"

  echo "Installed kanba ${VERSION} to ${INSTALL_DIR}/kanba"

  case ":$PATH:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
      echo "Warning: ${INSTALL_DIR} is not in your PATH. Add it with:"
      echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
      ;;
  esac
}

main
```

- [ ] **Step 4: Run test to verify it passes**

Run: `chmod +x install.sh && ./scripts/install_test.sh`
Expected: `PASS: install.sh installed a working kanba binary into a custom dir`

(Requires a real release to exist per the Step 2 note — run this after Task 3's end-to-end verification tag exists, or after the first real version tag is cut.)

- [ ] **Step 5: Commit**

```bash
git add install.sh scripts/install_test.sh
git commit -m "feat: add install.sh for downloading tagged releases"
```

---

## Self-Review Notes

- **Spec coverage:** version command (Task 1), reusable build script (Task 2), tag-triggered matrix + release publishing with checksums and changelog (Task 3), install script with version resolution/platform detection/checksum verification (Task 4) — all four spec sections have a task. Error-handling table entries (unsupported OS/arch, missing version, checksum mismatch, missing `~/.local/bin`) are implemented in Task 4's `install.sh` code, and the matrix `fail-fast: false` + `release` job dependency is in Task 3.
- **Placeholder scan:** no TBD/TODO; all steps contain complete runnable code.
- **Type/interface consistency:** `cmd.Version` (Task 1) is the exact `-ldflags -X` target used in Task 2's `build.sh` and Task 3's workflow. The artifact naming `kanba-<version>-<goos>-<goarch>.tar.gz` is identical across Task 2 (produces), Task 3 (produces to Release), and Task 4 (consumes). `checksums.txt` produced in Task 3 is read verbatim by Task 4.
