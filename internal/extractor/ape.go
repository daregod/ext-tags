package extractor

import (
	"fmt"
	"strings"

	"github.com/daregod/ext-tags/internal/ape"
)

var apeMagic = []byte("MAC ")

// extractAPEv2 reads an APEv2 tag appended to a file (APE, WavPack, …).
func extractAPEv2(path, format string) (*Result, error) {
	tag, err := ape.ReadFromFileEnd(path)
	if err != nil {
		return nil, fmt.Errorf("read ape tag: %w", err)
	}

	res := empty()
	res.Format = format

	for _, item := range tag.Items {
		if item.Flags&2 != 0 {
			if strings.EqualFold(item.Key, "Cover Art (Front)") {
				res.Cover = apeCover(item.Value)
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

	track, trackTotal := splitPair(fieldValue(res.Fields, "Track"))
	disc, discTotal := splitPair(fieldValue(res.Fields, "Disc"))
	putVar(res.Vars, "tracknumber", track)
	putVar(res.Vars, "totaltracks", trackTotal)
	putVar(res.Vars, "discnumber", disc)
	putVar(res.Vars, "disctotal", discTotal)

	if len(res.Fields) == 0 && res.Cover == nil {
		return res, ErrNoTags
	}
	return res, nil
}

func apeCover(value []byte) *Cover {
	name := "cover"
	data := value
	for i, b := range value {
		if b == 0 {
			name = string(value[:i])
			data = value[i+1:]
			break
		}
	}
	return &Cover{Filename: coverName(name, ""), MIME: "", Data: data}
}

func splitPair(value string) (string, string) {
	if value == "" {
		return "", ""
	}
	parts := strings.SplitN(value, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}
