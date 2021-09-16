
set -e
# From the `redo` recipe book:
redo-always

# Try to get a version number from git, if possible.
# This is a little more relevant than commit:
# a tuple of (tag, commits since tag, is-dirty)
if ! git describe --tags --always --dirty >$3; then
    echo "$0: Falling back to static version." >&2
    echo 'UNKNOWN' >$3
fi

# If the value hasn't changed, don't consider our dependencies to have changed.
# This can save us a lot of relinking!
redo-stamp <$3
