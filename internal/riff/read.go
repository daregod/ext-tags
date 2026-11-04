package riff

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/dhowden/tag"

	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/read"
)

// Read detects the RIFF-family container and returns its metadata.
func Read(path string) (*metadata.Document, error) {
	magic, err := read.ReadMagic(path, 40)
	if err != nil {
		return nil, err
	}
	switch {
	case len(magic) >= 12 && string(magic[0:4]) == "RIFF" && string(magic[8:12]) == "WAVE":
		md, err := ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read wav: %w", err)
		}
		return buildDocument(md, "WAVE")
	case len(magic) >= 12 && string(magic[0:4]) == "RF64" && string(magic[8:12]) == "WAVE":
		md, err := ReadRF64(path)
		if err != nil {
			return nil, fmt.Errorf("read rf64: %w", err)
		}
		return buildDocument(md, "RF64")
	case len(magic) >= 40 && string(magic[0:4]) == "riff" && string(magic[24:28]) == "wave":
		md, err := ReadWave64(path)
		if err != nil {
			return nil, fmt.Errorf("read wave64: %w", err)
		}
		return buildDocument(md, "WAVE64")
	case len(magic) >= 12 && string(magic[0:4]) == "FORM" && (string(magic[8:12]) == "AIFF" || string(magic[8:12]) == "AIFC"):
		md, err := ReadAIFF(path)
		if err != nil {
			return nil, fmt.Errorf("read aiff: %w", err)
		}
		return buildAIFF(md)
	}
	return nil, read.ErrNoTags
}

func buildDocument(md *Metadata, format string) (*metadata.Document, error) {
	res := read.New()
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
			read.ApplyTagMetadata(m, res)
		}
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

func buildAIFF(md *Metadata) (*metadata.Document, error) {
	res := read.New()
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
			read.ApplyTagMetadata(m, res)
		}
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
