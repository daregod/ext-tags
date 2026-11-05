#!/usr/bin/env bash
set -euo pipefail

version="${VERSION:?set VERSION}"
arch="${ARCH:-x86-64}"
binary="${BINARY:?set BINARY to the Linux ext-tags binary}"
out="${OUT:-ext-tags_${version}_${arch}.ipk}"
out_dir="$(dirname "$out")"
out_name="$(basename "$out")"
mkdir -p "$out_dir"
root="$(cd "$(dirname "$0")/../.." && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

mkdir -p "$tmp/control" "$tmp/data/opt/bin"
sed -e "s|@VERSION@|${version}|g" -e "s|@ARCH@|${arch}|g" \
  "$root/packaging/opkg/control.in" > "$tmp/control/control"
gzip -c "$binary" > "$tmp/data/opt/bin/ext-tags.bin.gz"
cp "$root/packaging/opkg/postinst" "$root/packaging/opkg/prerm" "$tmp/control/"
chmod 755 "$tmp/control/postinst" "$tmp/control/prerm"

tar -czf "$tmp/control.tar.gz" -C "$tmp/control" .
tar -czf "$tmp/data.tar.gz" -C "$tmp/data" .
printf '2.0\n' > "$tmp/debian-binary"
( cd "$tmp" && ar rcs "$out_name" debian-binary control.tar.gz data.tar.gz )
mv "$tmp/$out_name" "$out"
echo "$out"
