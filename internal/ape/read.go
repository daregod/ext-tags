package ape

import (
	"errors"
	"fmt"
	"strings"

	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/read"
)

var artKeys = map[string]string{
	"cover art (front)": "front",
	"cover art (back)":  "back",
	"artist picture":    "artist",
	"cover art (disc)":  "disc",
	"cover art (icon)":  "icon",
}

// Read reads an APEv2 tag appended to a file.
func Read(path, format string) (*metadata.Document, error) {
	tag, err := ReadFromFileEnd(path)
	if err != nil {
		if errors.Is(err, ErrNoTag) {
			return nil, read.ErrNoTags
		}
		return nil, fmt.Errorf("read ape tag: %w", err)
	}
	res := read.New()
	res.Format = format
	for _, item := range tag.Items {
		if item.Flags&2 != 0 {
			if kind, ok := artKeys[strings.ToLower(item.Key)]; ok {
				cover := apeCover(item.Value)
				if len(cover.Data) > 0 {
					res.Artwork[kind] = cover
					if kind == metadata.Front {
						res.Cover = cover
					}
				}
			}
			continue
		}
		key := item.Key
		value := strings.TrimSpace(string(item.Value))
		if key == "" || value == "" {
			continue
		}
		if prev, ok := res.Fields[key]; ok {
			if prev != value {
				res.Fields[key] = prev + "; " + value
			}
			continue
		}
		res.Fields[key] = value
	}
	track, trackTotal := read.SplitPair(read.FieldValue(res.Fields, "Track"))
	disc, discTotal := read.SplitPair(read.FieldValue(res.Fields, "Disc"))
	read.PutVar(res.Vars, "tracknumber", track)
	read.PutVar(res.Vars, "totaltracks", trackTotal)
	read.PutVar(res.Vars, "discnumber", disc)
	read.PutVar(res.Vars, "disctotal", discTotal)
	if len(res.Fields) == 0 && len(res.Artwork) == 0 {
		return res, read.ErrNoTags
	}
	return res, nil
}

func apeCover(value []byte) *metadata.Cover {
	name := "cover"
	data := value
	for i, b := range value {
		if b == 0 {
			name = string(value[:i])
			data = value[i+1:]
			break
		}
	}
	return &metadata.Cover{Filename: read.CoverName(name, ""), Data: data}
}
