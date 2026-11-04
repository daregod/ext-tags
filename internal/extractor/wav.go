package extractor

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/dhowden/tag"

	"github.com/daregod/ext-tags/internal/riff"
)

var (
	riffMagic = []byte("RIFF")
	rf64Magic = []byte("RF64")
	wvpkMagic = []byte("wvpk")
)

// extractWAV reads WAVE metadata: LIST/INFO fields and ID3v2 (including tags
// appended after the RIFF data).
func extractWAV(path string) (*Result, error) {
	md, err := riff.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read wav: %w", err)
	}
	return buildRIFF(md, "WAVE")
}

// extractRF64 reads RF64 metadata (RIFF with 64-bit sizes).
func extractRF64(path string) (*Result, error) {
	md, err := riff.ReadRF64(path)
	if err != nil {
		return nil, fmt.Errorf("read rf64: %w", err)
	}
	return buildRIFF(md, "RF64")
}

func buildRIFF(md *riff.Metadata, format string) (*Result, error) {
	res := empty()
	res.Format = format

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
		if key := infoField(id); key != "" {
			add(key, value)
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

func infoField(id string) string {
	switch id {
	case "INAM":
		return "TITLE"
	case "IART":
		return "ARTIST"
	case "IPRD":
		return "ALBUM"
	case "ICRD":
		return "YEAR"
	case "IGNR":
		return "GENRE"
	case "ICMT":
		return "COMMENT"
	case "ITRK", "IPRT":
		return "TRACKNUMBER"
	case "ICOP":
		return "COPYRIGHT"
	case "ISFT":
		return "ENCODER"
	}
	return ""
}
