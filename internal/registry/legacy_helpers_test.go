package registry_test

import (
	"fmt"

	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/registry"
	"github.com/daregod/ext-tags/internal/tagformat"
)

func readPath(path string, coverNames []string) (*metadata.Document, error) {
	reg := registry.NewDefault()
	reader, ok := reg.ReaderFor(path)
	if !ok {
		return nil, fmt.Errorf("no reader for %s", path)
	}
	return reader.Read(path, tagformat.ReadOptions{CoverNames: coverNames})
}
