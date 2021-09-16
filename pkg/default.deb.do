
# redo file to generate a Debian package.
# Format:
#   ${PACKAGE}.${ARCH}.deb
# Inputs:
# ../bin/${PACKAGE}.$ARCH
# ../${PACKAGE}.service
# ../${PACKAGE}.debcontrol
# ../version.txt
set -eu
exec >&2

ARCH="$(echo "$2" | cut -d. -f2)"
PACKAGE="$(echo "$2" | cut -d. -f1)"

if test -z "$ARCH" || test -z "$PACKAGE"
then
  echo >&2 "Expected target like <pkg>.<architecture>.deb"
  exit 1
fi

BIN="$(readlink -f ..)/bin/${PACKAGE}.${ARCH}"
SERVICE="$(readlink -f ..)/${PACKAGE}.service"
CONTROL="$(readlink -f ..)/${PACKAGE}.debcontrol"
STAMP="$(readlink -f ..)/version.txt"

redo-ifchange "$BIN" "$SERVICE" "$STAMP" "$CONTROL" *.template

VERSION="$(cat $STAMP | sed 's/^v//')"
WORKDIR="$(mktemp -d -p . XXXXXX.tmp)/${PACKAGE}_${VERSION}_${ARCH}"
echo >&2 "Building package in $WORKDIR"

# Generate the tree for dpkg-deb --build.
mkdir -p "$WORKDIR/DEBIAN"

# Control file:
cp "$CONTROL" "$WORKDIR/DEBIAN/control"
echo "Architecture: $ARCH" >>"$WORKDIR/DEBIAN/control"
echo "Version: $VERSION" >>"$WORKDIR/DEBIAN/control"

# Binary and unit:
mkdir -p "$WORKDIR/usr/bin"
mkdir -p "$WORKDIR/lib/systemd/system"
cp "$BIN" "$WORKDIR/usr/bin/${PACKAGE}"
cp "$SERVICE" "$WORKDIR/lib/systemd/system/${PACKAGE}.service"

# Systemd management scripts:
for f in "postinst" "postrm" "prerm"
do
  sed "s/%UNIT%/${PACKAGE}.service/g" "${f}.template" >"$WORKDIR/DEBIAN/${f}"
  chmod 755 "$WORKDIR/DEBIAN/${f}"
done

dpkg-deb --build "$WORKDIR"
mv "$WORKDIR".deb "$3"

# On success, clean up:
rm -rf "$WORKDIR"