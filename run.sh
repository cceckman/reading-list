#!/bin/sh -ex

ARCH="$(uname -m)"
case "$(uname -m)" in
  x86_64) ARCH="amd64";;
  aarch64) ARCH="arm64";;
esac

redo -j$(nproc) server/reading-list."$ARCH" server/test

export TAILSCALE_USE_WIP_CODE=true
exec ./server/reading-list."$ARCH" \
  --localTemplates server/dynamic \
  --localStatic server/static \
  --logmodule all \
  --storage testdata \
  --tsnet reading-list-dev

