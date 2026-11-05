package album

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/daregod/ext-tags/internal/yamlutil"
)

// Input is one track used to build an album.yaml skeleton.
type Input struct {
	AudioPath string
	Fields    map[string]string
}

var generatedIgnored = map[string]bool{
	"encoder":  true,
	"language": true,
	"vendor":   true,
}

// Generate builds an album skeleton from already normalised track fields.
func Generate(inputs []Input, artwork map[string]string) *Config {
	cfg := &Config{}
	if len(inputs) == 0 {
		return cfg
	}
	for i := range inputs {
		ensureTrack(inputs[i].Fields, len(inputs), i+1)
	}

	common := make(map[string]string)
	for key, value := range inputs[0].Fields {
		if generatedIgnored[strings.ToLower(key)] || value == "" {
			continue
		}
		same := true
		for _, input := range inputs[1:] {
			if input.Fields[key] != value {
				same = false
				break
			}
		}
		if same {
			common[key] = value
		}
	}
	if len(common) > 0 {
		cfg.Album = common
	}

	used := map[string]bool{}
	for _, input := range inputs {
		key := generatedTrackKey(input)
		if used[key] {
			key = strings.TrimSuffix(filepath.Base(input.AudioPath), filepath.Ext(input.AudioPath))
		}
		used[key] = true

		fields := make(map[string]string)
		for field, value := range input.Fields {
			if generatedIgnored[strings.ToLower(field)] || value == "" {
				continue
			}
			if _, ok := common[field]; ok {
				continue
			}
			fields[field] = value
		}
		if len(fields) > 0 {
			if cfg.Tracks == nil {
				cfg.Tracks = make(map[string]map[string]string)
			}
			cfg.Tracks[key] = fields
		}
	}
	if len(artwork) > 0 {
		cfg.Artwork = artwork
	}
	return cfg
}

// Marshal serialises the album skeleton as YAML without a BOM.
func (c *Config) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(c); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return yamlutil.TrimBOM(buf.Bytes()), nil
}

// Save writes album.yaml atomically. It refuses to overwrite unless force.
func (c *Config) Save(dir string, force bool) (string, error) {
	data, err := c.Marshal()
	if err != nil {
		return "", err
	}
	target := filepath.Join(dir, "album.yaml")
	if !force {
		if _, err := os.Stat(target); err == nil {
			return target, fmt.Errorf("%s already exists (use --force)", target)
		}
	}
	tmp, err := os.CreateTemp(dir, ".album-*.yaml")
	if err != nil {
		return target, err
	}
	name := tmp.Name()
	defer func() { _ = os.Remove(name) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return target, err
	}
	if err := tmp.Close(); err != nil {
		return target, err
	}
	if err := os.Chmod(name, 0644); err != nil {
		return target, err
	}
	if force {
		_ = os.Remove(target)
	}
	if err := os.Rename(name, target); err != nil {
		return target, err
	}
	return target, nil
}

func generatedTrackKey(input Input) string {
	for key, value := range input.Fields {
		if !strings.EqualFold(key, "Track") {
			continue
		}
		value = strings.TrimSpace(value)
		if i := strings.IndexByte(value, '/'); i >= 0 {
			value = value[:i]
		}
		if n, err := strconv.Atoi(value); err == nil {
			return strconv.Itoa(n)
		}
	}
	base := filepath.Base(input.AudioPath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}
func ensureTrack(fields map[string]string, total, index int) {
	if fields == nil {
		return
	}
	for key, value := range fields {
		if !strings.EqualFold(key, "Track") {
			continue
		}
		number := strings.TrimSpace(value)
		if i := strings.IndexByte(number, '/'); i >= 0 {
			number = number[:i]
		}
		if number == "" {
			number = strconv.Itoa(index)
		}
		fields[key] = fmt.Sprintf("%s/%d", number, total)
		return
	}
	fields["Track"] = fmt.Sprintf("%d/%d", index, total)
}
