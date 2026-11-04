package extractor

import (
	"github.com/daregod/ext-tags/internal/ogg"
)

var oggMagic = []byte("OggS")

// extractOgg reads comments from Ogg streams that dhowden/tag does not
// support (Speex, FLAC-in-Ogg). Other Ogg codecs fall back to the library.
func extractOgg(path string) (*Result, error) {
	tags, pic, err := ogg.ReadTags(path)
	if err != nil {
		return extractTaglib(path)
	}

	res := empty()
	res.Format = "OGG"

	for _, tag := range tags {
		if tag.Name == "" || tag.Value == "" {
			continue
		}
		if prev, ok := res.Fields[tag.Name]; ok {
			if prev != tag.Value {
				res.Fields[tag.Name] = prev + "; " + tag.Value
			}
			continue
		}
		res.Fields[tag.Name] = tag.Value
	}

	if pic != nil && len(pic.Data) > 0 {
		res.Cover = &Cover{Filename: coverName("", pic.MIME), MIME: pic.MIME, Data: pic.Data}
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
