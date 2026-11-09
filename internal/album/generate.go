package album

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	root := &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode}}}
	body := root.Content[0]
	appendMap := func(name string, values map[string]string) {
		if len(values) == 0 {
			return
		}
		body.Content = append(body.Content, scalarNode(name), mappingNode(values))
	}
	appendMap("album", c.Album)
	if len(c.Tracks) > 0 {
		tracks := &yaml.Node{Kind: yaml.MappingNode}
		for _, key := range sortedKeys(c.Tracks) {
			if len(c.Tracks[key]) == 0 {
				continue
			}
			tracks.Content = append(tracks.Content, scalarNode(key), mappingNode(c.Tracks[key]))
		}
		if len(tracks.Content) > 0 {
			body.Content = append(body.Content, scalarNode("tracks"), tracks)
		}
	}
	appendMap("artwork", c.Artwork)
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(root); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return yamlutil.TrimBOM(buf.Bytes()), nil
}

func mappingNode(values map[string]string) *yaml.Node {
	node := &yaml.Node{Kind: yaml.MappingNode}
	for _, key := range sortedKeys(values) {
		node.Content = append(node.Content, scalarNode(key), scalarNode(values[key]))
	}
	return node
}

func scalarNode(value string) *yaml.Node {
	if strings.ContainsAny(value, "\r\n") {
		value = strings.ReplaceAll(value, "\r\n", "\n")
		value = strings.ReplaceAll(value, "\r", "\n")
		lines := strings.Split(value, "\n")
		for i := range lines {
			lines[i] = strings.TrimRight(lines[i], " \t")
		}
		value = strings.Join(lines, "\n")
		if !strings.HasSuffix(value, "\n") {
			value += "\n"
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Style: yaml.LiteralStyle, Value: value}
	}
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
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
		if _, err := strconv.Atoi(value); err == nil {
			return value
		}
	}
	base := filepath.Base(input.AudioPath)
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	if number := leadingNumber(stem); number != "" {
		return number
	}
	return stem
}
func formatTrackValue(number string, total int) string {
	number = strings.TrimSpace(number)
	if number == "" {
		return ""
	}
	n, err := strconv.Atoi(number)
	if err != nil {
		return fmt.Sprintf("%s/%d", number, total)
	}
	width := 1
	if total >= 10 {
		width = len(strconv.Itoa(total))
	}
	return fmt.Sprintf("%0*d/%d", width, n, total)
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
		fields[key] = formatTrackValue(number, total)
		return
	}
	fields["Track"] = formatTrackValue(strconv.Itoa(index), total)
}
