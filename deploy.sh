#!/bin/sh -ex

usage() {
  echo >&2 "Usage: "
  echo >&2 "$0 <hostname>"
  echo >&2
  echo >&2 "Build & deploy to the specified hostname"
  exit 1
}

if test "$#" -ne 1
then
  echo >&2 "Wrong number of arguments"
  usage
fi

TARGET="$1"

ARCH="$(ssh "$TARGET" uname -m)"
case "$ARCH" in
  x86_64) ARCH="amd64";;
  aarch64) ARCH="arm64";;
esac

PKG="reading-list.$ARCH.deb"

redo -j$(nproc) "pkg/$PKG" server/test
scp pkg/"$PKG" "$TARGET":"$PKG"
ssh "$TARGET" "sudo dpkg -i $PKG"
