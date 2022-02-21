

redo-ifchange \
  static/all \
  $(find . -name '*.go') \
  ./go.mod \
  ./go.sum

exec >&2
go test ./...

