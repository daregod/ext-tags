#!/usr/bin/env bash
set -euo pipefail

version="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null | sed 's/^v//')}"
version="${version:-0.0.0}"
nfpm_bin="${NFPM:-nfpm}"
root="$(cd "$(dirname "$0")/.." && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

for arch in amd64 arm64; do
  binary="$root/dist/ext-tags_linux_${arch}"
  if [[ ! -f "$binary" ]]; then
    echo "missing $binary" >&2
    exit 1
  fi
  config="$tmp/nfpm-${arch}.yaml"
  sed \
    -e "s|@ARCH@|${arch}|g" \
    -e "s|@VERSION@|${version}|g" \
    -e "s|@BINARY@|${binary}|g" \
    "$root/packaging/nfpm.yaml" > "$config"

  "$nfpm_bin" package --config "$config" --packager deb --target "$root/dist/ext-tags_${version}_${arch}.deb"
  "$nfpm_bin" package --config "$config" --packager rpm --target "$root/dist/ext-tags-${version}-1.${arch}.rpm"
  "$nfpm_bin" package --config "$config" --packager apk --target "$root/dist/ext-tags_${version}_${arch}.apk"
done
