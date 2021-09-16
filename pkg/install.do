
PACKAGE=""

case "$(uname -m)" in
  x86_64) PACKAGE="amd64";;
  aarch64) PACKAGE="arm64";;
  *) PACKAGE="$(uname -m)";;
esac

redo-ifchange "$PACKAGE".deb
sudo dpkg -i "$PACKAGE".deb >&2
