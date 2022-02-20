#!/bin/sh -ex

ARCH="$(uname -m)"
case "$(uname -m)" in
  x86_64) ARCH="amd64";;
  aarch64) ARCH="arm64";;
esac

redo -j$(nproc) server/reading-list."$ARCH"

export TAILSCALE_USE_WIP_CODE=true
cd "$(dirname $(realpath $0))"/server
exec ./reading-list."$ARCH" \
  --allowLocal \
  --logmodule=all \
  --storageDir=$(pwd)/../testdata/
