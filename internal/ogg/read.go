package ogg

import (
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/read"
)

// Read reads Ogg metadata, falling back to dhowden/tag for codecs not covered
// by the native Ogg parser.
func Read(path string) (*metadata.Document, error) {
	tags, pic, err := ReadTags(path)
	if err != nil {
		return read.Taglib(path)
	}
	res := read.New()
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
		cover := &metadata.Cover{Filename: read.CoverName("", pic.MIME), MIME: pic.MIME, Data: pic.Data}
		res.Cover = cover
		res.Artwork[metadata.Front] = cover
	}
	track, trackTotal := read.SplitPair(read.FieldValue(res.Fields, "TRACKNUMBER"))
	disc, discTotal := read.SplitPair(read.FieldValue(res.Fields, "DISCNUMBER"))
	read.PutVar(res.Vars, "tracknumber", track)
	read.PutVar(res.Vars, "totaltracks", trackTotal)
	read.PutVar(res.Vars, "discnumber", disc)
	read.PutVar(res.Vars, "disctotal", discTotal)
	if len(res.Fields) == 0 && res.Cover == nil {
		return res, read.ErrNoTags
	}
	return res, nil
}
