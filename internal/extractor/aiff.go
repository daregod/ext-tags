package extractor

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/dhowden/tag"

	"github.com/daregod/ext-tags/internal/riff"
)

var formMagic = []byte("FORM")

// extractAIFF reads AIFF metadata: NAME/AUTH/ANNO text chunks and an
// embedded ID3v2 chunk.
func extractAIFF(path string) (*Result, error) {
	md, err := riff.ReadAIFF(path)
	if err != nil {
		return nil, fmt.Errorf("read aiff: %w", err)
	}

	res := empty()
	res.Format = "AIFF"

	add := func(key, value string) {
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			return
		}
		if prev, ok := res.Fields[key]; ok {
			if prev != value {
				res.Fields[key] = prev + "; " + value
			}
			return
		}
		res.Fields[key] = value
	}

	for id, value := range md.Info {
		switch id {
		case "NAME":
			add("TITLE", value)
		case "AUTH":
			add("ARTIST", value)
		case "ANNO":
			add("COMMENT", value)
		case "(c) ":
			add("COPYRIGHT", value)
		}
	}

	if len(md.ID3) > 0 {
		if m, err := tag.ReadFrom(bytes.NewReader(md.ID3)); err == nil {
			applyTagMetadata(m, res)
		}
	}

	track, trackTotal := splitPair(fieldValue(res.Fields, "TRACKNUMBER"))
	disc, discTotal := splitPair(fieldValue(res.Fields, "DISCNUMBER"))
	putVar(res.Vars, "tracknumber", track)
	putVar(res.Vars, "totaltracks", trackTotal)
	putVar(res.Vars, "discnumber", disc)
	putVar(res.Vars, "disctotal", discTotal)

	if len(res.Fields) == 0 && res.Cover == nil {
		return res, ErrNoTags
	}
	return res, nil
}
