// Package album loads optional per-album metadata overrides.
package album

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/daregod/ext-tags/internal/artwork"
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/yamlutil"
)

var names = []string{"album.yaml", "album.yml", ".ext-tags.album.yaml"}

// Config is an optional album.yaml overlay stored next to the audio files.
type Config struct {
	Album   map[string]string            `yaml:"album,omitempty"`
	Tracks  map[string]map[string]string `yaml:"tracks,omitempty"`
	Artwork map[string]string            `yaml:"artwork,omitempty"`

	dir  string
	path string
}

// Path returns the album.yaml path, if one was loaded.
func (c *Config) Path() string {
	if c == nil {
		return ""
	}
	return c.path
}

// Load finds and parses album.yaml/album.yml/.ext-tags.album.yaml in dir.
// A missing file is not an error and returns (nil, nil).
func Load(dir string) (*Config, error) {
	for _, name := range names {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		cfg := &Config{dir: dir, path: path}
		if err := yaml.Unmarshal(yamlutil.TrimBOM(data), cfg); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		return cfg, nil
	}
	return nil, nil
}

// Apply overlays album-level fields and the matching track override onto doc.
func (c *Config) Apply(doc *metadata.Document, audioPath string) map[string]string {
	if c == nil || doc == nil {
		return nil
	}
	overrides := c.fieldsFor(doc, audioPath)
	for key, value := range overrides {
		if value == "" {
			deleteField(doc, key)
			continue
		}
		setField(doc.Fields, key, value)
	}
	updateTrackVars(doc)
	return overrides
}

func (c *Config) fieldsFor(doc *metadata.Document, audioPath string) map[string]string {
	out := make(map[string]string, len(c.Album))
	for key, value := range c.Album {
		out[key] = value
	}
	for _, key := range c.trackKeys(doc, audioPath) {
		if fields, ok := c.Tracks[key]; ok {
			for field, value := range fields {
				out[field] = value
			}
			break
		}
	}
	return out
}

// ArtworkCovers loads files referenced by the artwork map.
func (c *Config) ArtworkCovers() (map[string]*metadata.Cover, error) {
	if c == nil || len(c.Artwork) == 0 {
		return nil, nil
	}
	covers := make(map[string]*metadata.Cover, len(c.Artwork))
	for kind, name := range c.Artwork {
		if name == "" {
			continue
		}
		img, err := artwork.Load(filepath.Join(c.dir, name))
		if err != nil {
			return nil, fmt.Errorf("album artwork %s: %w", kind, err)
		}
		covers[kind] = &metadata.Cover{Filename: img.Filename, MIME: img.MIME, Data: img.Data}
	}
	return covers, nil
}

func (c *Config) trackKeys(doc *metadata.Document, audioPath string) []string {
	var keys []string
	for _, track := range []string{
		fieldsValue(doc.Fields, "Track"),
		fieldsValue(doc.Fields, "TRACKNUMBER"),
		doc.Vars["tracknumber"],
	} {
		if track == "" {
			continue
		}
		keys = append(keys, track)
		if i := strings.IndexByte(track, '/'); i >= 0 {
			keys = append(keys, track[:i])
		}
		keys = append(keys, trimNumber(track))
	}
	base := filepath.Base(audioPath)
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	keys = append(keys, base, stem)
	if number := leadingNumber(stem); number != "" {
		keys = append(keys, number, trimNumber(number))
	}
	return keys
}

func leadingNumber(value string) string {
	i := 0
	for i < len(value) && value[i] >= '0' && value[i] <= '9' {
		i++
	}
	if i == 0 {
		return ""
	}
	return value[:i]
}

func updateTrackVars(doc *metadata.Document) {
	if doc.Vars == nil {
		doc.Vars = map[string]string{}
	}
	for _, spec := range []struct {
		field, number, total string
	}{
		{"Track", "tracknumber", "totaltracks"},
		{"Disc", "discnumber", "disctotal"},
	} {
		value := fieldsValue(doc.Fields, spec.field)
		if value == "" {
			continue
		}
		num, total, ok := strings.Cut(value, "/")
		doc.Vars[spec.number] = trimNumber(num)
		if ok {
			doc.Vars[spec.total] = trimNumber(total)
		}
	}
}

func fieldsValue(fields map[string]string, name string) string {
	for key, value := range fields {
		if strings.EqualFold(key, name) {
			return value
		}
	}
	return ""
}

func trimNumber(value string) string {
	value = strings.TrimSpace(value)
	if n, err := strconv.Atoi(value); err == nil {
		return strconv.Itoa(n)
	}
	return strings.TrimLeft(value, "0")
}
func deleteField(doc *metadata.Document, name string) {
	for key := range doc.Fields {
		if strings.EqualFold(key, name) {
			delete(doc.Fields, key)
		}
	}
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "track":
		delete(doc.Vars, "tracknumber")
		delete(doc.Vars, "totaltracks")
	case "disc":
		delete(doc.Vars, "discnumber")
		delete(doc.Vars, "disctotal")
	}
}

func setField(fields map[string]string, name, value string) {
	found := false
	for key := range fields {
		if strings.EqualFold(key, name) {
			fields[key] = value
			found = true
		}
	}
	if !found {
		fields[name] = value
	}
}
