
ARCH="$2"
OUTPUT="$(readlink -f $3)"

( cd .. && GOARCH="$ARCH" go build -o "$OUTPUT")

# Good build chains (like Cargo and `go`) do a pretty good job of caching and
# providing deterministic outputs.
redo-always
redo-stamp <"$OUTPUT"
