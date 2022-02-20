// Provide some additional filesystem handling types on top of io/fs.
// See also https://github.com/golang/go/issues/45757.
package fs

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// A filesystem that also allows "create" operations.
type CreateFS interface {
	fs.FS

	// Creates the file with the given name, iff it doesn't already exist.
	Create(name string) (fs.File, error)
}

// Creates an operating-system level
func NativeFS(dir string) CreateFS {
	return &nativeFS{
		FS:  os.DirFS(dir),
		dir: dir,
	}
}

type nativeFS struct {
	fs.FS
	dir string
}

func (n *nativeFS) Create(name string) (fs.File, error) {
	return os.Create(filepath.Join(n.dir, name))
}

// A readable-writeable file.
type RWFile interface {
	fs.File
	io.Reader
	io.Writer
	io.Closer
	io.Seeker
}
