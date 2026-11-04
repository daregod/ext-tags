package importer

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/daregod/ext-tags/internal/ape"
)

var apeArtKeys = []struct {
	kind string
	key  string
}{
	{"front", "Cover Art (Front)"},
	{"back", "Cover Art (Back)"},
	{"artist", "Artist Picture"},
	{"disc", "Cover Art (Disc)"},
	{"icon", "Cover Art (Icon)"},
}

// ImportAPEv2 appends a bare APEv2 tag to files that keep metadata at the end
// (APE, WavPack, Musepack, TAK, AC3, DTS).
func ImportAPEv2(path string, tf *TagFile, opts Options) error {
	existing, err := ape.ReadFromFileEnd(path)
	if err != nil && !errors.Is(err, ape.ErrNoTag) {
		return err
	}
	if opts.Backup {
		if err := copyFile(path, path+".bak"); err != nil {
			return err
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	body, err := stripAPETag(data)
	if err != nil {
		return err
	}

	tag := ape.NewTag()
	replaced := make(map[string]bool, len(tf.Fields))
	for field := range tf.Fields {
		replaced[strings.ToLower(field)] = true
	}
	if !opts.Replace && existing != nil {
		for _, item := range existing.Items {
			if item.Flags&2 != 0 {
				continue
			}
			if replaced[strings.ToLower(item.Key)] {
				continue
			}
			tag.AddText(item.Key, string(item.Value))
		}
	}

	keys := make([]string, 0, len(tf.Fields))
	for key := range tf.Fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		tag.AddText(key, tf.Fields[key])
	}

	covers := tf.Covers
	if len(covers) == 0 && tf.Cover != nil {
		covers = map[string]*Cover{"front": tf.Cover}
	}
	added := make(map[string]bool)
	for _, spec := range apeArtKeys {
		cover := covers[spec.kind]
		if cover == nil || len(cover.Data) == 0 {
			continue
		}
		name := cover.Filename
		if name == "" {
			name = spec.kind + ".jpg"
		}
		tag.AddCoverArt(spec.key, name, cover.Data)
		added[spec.kind] = true
	}
	if !opts.Replace && existing != nil {
		known := make(map[string]string)
		for _, spec := range apeArtKeys {
			known[strings.ToLower(spec.key)] = spec.kind
		}
		for _, item := range existing.Items {
			if item.Flags&2 == 0 {
				continue
			}
			if kind, ok := known[strings.ToLower(item.Key)]; ok && added[kind] {
				continue
			}
			tag.AddBinary(item.Key, item.Value)
		}
	}

	embedded, err := tag.SerializeEmbedded()
	if err != nil {
		return err
	}
	out := append(append([]byte(nil), body...), embedded...)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, info.Mode().Perm())
}

func stripAPETag(data []byte) ([]byte, error) {
	if len(data) < 32 || string(data[len(data)-32:len(data)-24]) != "APETAGEX" {
		return data, nil
	}
	tagSize := int(uint32(data[len(data)-20]) | uint32(data[len(data)-19])<<8 | uint32(data[len(data)-18])<<16 | uint32(data[len(data)-17])<<24)
	if tagSize < 32 || tagSize > len(data) {
		return nil, fmt.Errorf("invalid APEv2 tag size %d", tagSize)
	}
	start := len(data) - tagSize
	if start >= 32 && string(data[start-32:start-24]) == "APETAGEX" {
		start -= 32
	}
	return data[:start], nil
}
