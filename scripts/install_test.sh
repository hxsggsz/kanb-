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
