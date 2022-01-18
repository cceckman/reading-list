
# redo file to build a Go binary for a requested architecture.
set -eu

# We need embedded content to be prepared.
redo static/all

# Get the architecture from the requested target.
GOARCH="$(echo "$1" | cut -d'.' -f2)"

if test -z "$GOARCH"
then
  echo >&2 "Unknown or unsupported architecture $GOARCH; expected target like <binary>.<architecture>.bin"
  exit 1
fi

OUTPUT="$(readlink -f $3)"
( GOARCH="$GOARCH" go build -o "$OUTPUT" )

# `go` is pretty good at caching and reproducibility.
# Always rebuild, but use stamping to suppress downstream targets.
redo-always
redo-stamp <"$3"

