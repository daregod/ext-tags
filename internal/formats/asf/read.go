package asf

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/read"
)

// Read reads ASF/WMA/WMV metadata.
func Read(path string) (*metadata.Document, error) {
	md, err := ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read asf: %w", err)
	}
	res := read.New()
	res.Format = "ASF"
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
	add("TITLE", md.Title)
	add("ARTIST", md.Author)
	add("COPYRIGHT", md.Copyright)
	add("COMMENT", md.Description)
	add("RATING", md.Rating)
	for _, d := range md.Descriptors {
		name := strings.TrimPrefix(d.Name, "WM/")
		var key string
		switch strings.ToLower(name) {
		case "albumtitle":
			key = "ALBUM"
		case "albumartist":
			key = "ALBUMARTIST"
		case "genre":
			key = "GENRE"
		case "year", "date":
			key = "YEAR"
		case "tracknumber":
			key = "TRACKNUMBER"
		case "partofset":
			key = "DISCNUMBER"
		case "composer":
			key = "COMPOSER"
		case "publisher":
			key = "PUBLISHER"
		case "author":
			key = "ARTIST"
		case "title":
			key = "TITLE"
		case "description":
			key = "COMMENT"
		case "picture":
			if d.Type == 1 {
				if pic, ok := ParsePicture(d.Value); ok {
					cover := &metadata.Cover{Filename: read.CoverName("", pic.MIME), MIME: pic.MIME, Data: pic.Data}
					res.Cover = cover
					res.Artwork[metadata.Front] = cover
				}
			}
			continue
		case "track":
			if d.Type == 0 {
				if n, err := strconv.Atoi(strings.TrimSpace(UTF16String(d.Value))); err == nil && n >= 0 {
					add("TRACKNUMBER", strconv.Itoa(n+1))
				}
			}
			continue
		default:
			continue
		}
		switch d.Type {
		case 0:
			add(key, UTF16String(d.Value))
		case 2:
			if len(d.Value) >= 4 && binary.LittleEndian.Uint32(d.Value) != 0 {
				add(key, "1")
			} else {
				add(key, "0")
			}
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
