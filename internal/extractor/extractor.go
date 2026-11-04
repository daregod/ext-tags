// Package extractor is a compatibility facade over the format registry.
package extractor

import (
	"fmt"
	"path/filepath"

	"github.com/daregod/ext-tags/internal/format"
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/read"
	"github.com/daregod/ext-tags/internal/registry"
)

// ErrNoTags means the file has no metadata block at all.
var ErrNoTags = read.ErrNoTags

// Cover is a picture embedded in the container.
type Cover = metadata.Cover

// Result is raw metadata, before any field rules are applied.
type Result = metadata.Document

// Options tune container parsing.
type Options struct {
	CoverNames []string
}

// IsSupported reports whether path has one of the configured extensions.
func IsSupported(path string, extensions []string) bool {
	return read.IsSupported(path, extensions)
}

// Extract reads metadata from path with default options.
func Extract(path string) (*Result, error) {
	return ExtractWithOptions(path, Options{})
}

// ExtractWithOptions reads metadata through the default format registry.
func ExtractWithOptions(path string, opts Options) (*Result, error) {
	reader, ok := registry.NewDefault().ReaderFor(path)
	if !ok {
		return nil, fmt.Errorf("unsupported format %s", filepath.Ext(path))
	}
	return reader.Read(path, format.ReadOptions{CoverNames: opts.CoverNames})
}
