#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
assets="$root/backend/internal/server/web/dist"
# Usage: ./scripts/build.sh [output-directory] [version]
# VERSION can also supply the version; the second argument takes precedence.
output=${1:-"$root/build"}
version=${2:-${VERSION:-dev}}
case "$version" in
  *[!a-zA-Z0-9._+-]*)
    printf 'Invalid version: %s\n' "$version" >&2
    exit 1
    ;;
esac
case "$output" in
  /*) ;;
  *) output="$(pwd)/$output" ;;
esac

cd "$root/frontend"
npm ci
npm run build

# Keep generated files out of the source tree after building.
if [ -e "$assets" ]; then
  echo "Refusing to overwrite existing embedded assets: $assets" >&2
  exit 1
fi
trap 'rm -rf "$assets"' EXIT HUP INT TERM
cp -R "$root/frontend/dist" "$assets"
mkdir -p "$output"
cd "$root/backend"

for target in darwin/arm64 windows/amd64 windows/arm64 linux/amd64 linux/arm64; do
  os=${target%/*}
  arch=${target#*/}
  binary="rewind-$os-$arch"
  if [ "$os" = windows ]; then
    binary="$binary.exe"
  fi
  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -ldflags "-X main.Version=$version" -o "$output/$binary" ./cmd/rewind
  printf 'Built %s\n' "$output/$binary"
done
