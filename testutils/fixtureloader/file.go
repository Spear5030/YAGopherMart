package fixtureloader

import (
	"io"
	"io/fs"
	"testing"
)

// NamedReadCloser represents a named ReadCloser interface
type NamedReadCloser interface {
	io.ReadCloser
	Name() string
}

type file struct {
	t *testing.T
	fs.File
}

func NewFile(t *testing.T, fsFile fs.File) *file {
	return &file{t, fsFile}
}

func (f *file) Name() string {
	stat, err := f.Stat()
	if err != nil {
		f.t.Fatal(err)
	}

	return stat.Name()
}
