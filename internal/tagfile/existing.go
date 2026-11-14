package tagfile

import (
	"bytes"
	"path/filepath"
	"strings"

	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/metadata"
)

// ExistingPictures returns all binary artwork items from an existing .tag in order.
func ExistingPictures(existing *ape.ParsedTag) []metadata.Picture {
	if existing == nil {
		return nil
	}
	keys := map[string]string{
		"cover art (front)": "front",
		"cover art (back)":  "back",
		"artist picture":    "artist",
		"cover art (disc)":  "disc",
		"cover art (icon)":  "icon",
	}
	var out []metadata.Picture
	for _, item := range existing.Items {
		if item.Flags&2 == 0 {
			continue
		}
		slot := ""
		if kind, ok := keys[strings.ToLower(item.Key)]; ok {
			slot = kind
		}
		filename, data := splitCoverValue(item.Value)
		if len(data) == 0 {
			continue
		}
		out = append(out, metadata.Picture{Source: "tag", Slot: slot, Filename: filename, MIME: coverMIME(filename), Data: data})
	}
	return out
}

// ExistingArtwork returns binary artwork slots from an existing .tag.
func ExistingArtwork(existing *ape.ParsedTag) map[string]*metadata.Cover {
	out := make(map[string]*metadata.Cover)
	if existing == nil {
		return out
	}
	keys := map[string]string{
		"cover art (front)": "front",
		"cover art (back)":  "back",
		"artist picture":    "artist",
		"cover art (disc)":  "disc",
		"cover art (icon)":  "icon",
	}
	for _, item := range existing.Items {
		if item.Flags&2 == 0 {
			continue
		}
		kind, ok := keys[strings.ToLower(item.Key)]
		if !ok {
			continue
		}
		filename, data := splitCoverValue(item.Value)
		if len(data) == 0 {
			continue
		}
		out[kind] = &metadata.Cover{Filename: filename, MIME: coverMIME(filename), Data: data}
	}
	return out
}

func splitCoverValue(value []byte) (string, []byte) {
	if i := bytes.IndexByte(value, 0); i >= 0 {
		return string(value[:i]), value[i+1:]
	}
	return "", value
}

func coverMIME(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	}
	return "image/jpeg"
}
