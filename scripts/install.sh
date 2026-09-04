#!/bin/sh
set -eu

repo="fu9zhou/finish-bit"
install_dir="${FINISHBIT_INSTALL_DIR:-${HOME}/.local/bin}"

case "$(uname -s)" in
  Linux) os="linux" ;;
  Darwin) os="darwin" ;;
  *) echo "unsupported operating system" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) arch="x86_64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "unsupported architecture" >&2; exit 1 ;;
esac

version="${FINISHBIT_VERSION:-}"
if [ -z "$version" ]; then
  version="$(curl -fsSL "https://api.github.com/repos/${repo}/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n 1)"
fi
[ -n "$version" ] || { echo "could not resolve a release version" >&2; exit 1; }

archive="finish-bit_${version#v}_${os}_${arch}.tar.gz"
base="https://github.com/${repo}/releases/download/${version}"
temporary="$(mktemp -d)"
trap 'rm -rf "$temporary"' EXIT INT TERM

curl -fsSL "${base}/${archive}" -o "${temporary}/${archive}"
curl -fsSL "${base}/checksums.txt" -o "${temporary}/checksums.txt"
expected="$(awk -v name="$archive" '$2 == name {print $1}' "${temporary}/checksums.txt")"
[ -n "$expected" ] || { echo "archive is missing from checksums.txt" >&2; exit 1; }
actual="$(sha256sum "${temporary}/${archive}" | awk '{print $1}')"
[ "$expected" = "$actual" ] || { echo "archive checksum verification failed" >&2; exit 1; }

tar -xzf "${temporary}/${archive}" -C "$temporary"
mkdir -p "$install_dir"
install -m 0755 "${temporary}/fnsh" "${install_dir}/fnsh"
echo "installed fnsh ${version} to ${install_dir}/fnsh"
