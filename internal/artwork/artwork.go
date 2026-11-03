// Package artwork finds cover images stored next to an audio file.
package artwork

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrNotFound means no cover image matched the configured names.
var ErrNotFound = errors.New("no cover image found")

// DefaultNames are the base names checked when the configuration is empty.
var DefaultNames = []string{"cover", "folder", "box", "disc", "media"}

// Extensions are checked in this order for every base name.
var Extensions = []string{".jpg", ".jpeg", ".png", ".webp", ".gif"}

// MaxSize is the largest sidecar image we are willing to embed.
const MaxSize = 32 << 20

// Image is a cover image found on disk.
type Image struct {
	Path     string
	Filename string
	MIME     string
	Data     []byte
}

// Find looks for <name><ext> (case-insensitive) in dir, honouring the order
// of names and extensions. Empty names fall back to DefaultNames.
func Find(dir string, names []string) (*Image, error) {
	if len(names) == 0 {
		names = DefaultNames
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	index := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		index[strings.ToLower(entry.Name())] = entry.Name()
	}

	for _, name := range names {
		for _, ext := range Extensions {
			actual, ok := index[strings.ToLower(name)+ext]
			if !ok {
				continue
			}
			path := filepath.Join(dir, actual)
			info, err := os.Stat(path)
			if err != nil || info.Size() > MaxSize {
				continue
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("read cover %s: %w", path, err)
			}
			return &Image{Path: path, Filename: actual, MIME: mimeFor(ext), Data: data}, nil
		}
	}
	return nil, ErrNotFound
}

func mimeFor(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	}
	return "application/octet-stream"
}
