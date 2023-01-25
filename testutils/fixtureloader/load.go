package fixtureloader

import (
	"bytes"
	"embed"
	"encoding/json"
	"io"
	"testing"
	"text/template"

	"github.com/stretchr/testify/require"
)

type Loader struct {
	fixtures embed.FS
}

func NewLoader(fixtures embed.FS) *Loader {
	return &Loader{fixtures: fixtures}
}

// LoadAPIFixture is not a Loader method because it is forbidden for method to have [T any] construction by Golang
func LoadAPIFixture[T any](t *testing.T, loader *Loader, path string) T {
	var data T

	fsFile, err := loader.fixtures.Open(path)
	require.NoError(t, err)

	defer func() { _ = fsFile.Close() }()

	err = json.NewDecoder(fsFile).Decode(&data)
	require.NoError(t, err)

	return data
}

func (l *Loader) LoadString(t *testing.T, path string) string {
	fsFile, err := l.fixtures.Open(path)
	require.NoError(t, err)

	defer func() { _ = fsFile.Close() }()

	data, err := io.ReadAll(fsFile)
	require.NoError(t, err)

	return string(data)
}

func (l *Loader) LoadFile(t *testing.T, path string) NamedReadCloser {
	fsFile, err := l.fixtures.Open(path)
	require.NoError(t, err)

	return NewFile(t, fsFile)
}

func (l *Loader) LoadTemplate(t *testing.T, path string, data any) string {
	tempData := l.LoadString(t, path)

	temp, err := template.New(path).Parse(tempData)
	require.NoError(t, err)

	buf := bytes.Buffer{}

	err = temp.Execute(&buf, data)
	require.NoError(t, err)

	return buf.String()
}
