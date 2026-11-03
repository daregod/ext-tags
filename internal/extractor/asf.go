package extractor

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"external-tags/internal/asf"
)

var asfMagic = []byte{0x30, 0x26, 0xB2, 0x75}

// extractASF reads Windows Media metadata (WMV/WMA).
func extractASF(path string) (*Result, error) {
	md, err := asf.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read asf: %w", err)
	}

	res := empty()
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
		case "year":
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
				if pic, ok := asf.ParsePicture(d.Value); ok {
					res.Cover = &Cover{Filename: coverName("", pic.MIME), MIME: pic.MIME, Data: pic.Data}
				}
			}
			continue
		case "track":
			// WM/Track is zero-based.
			if d.Type == 0 {
				if n, err := strconv.Atoi(strings.TrimSpace(asf.UTF16String(d.Value))); err == nil && n >= 0 {
					add("TRACKNUMBER", strconv.Itoa(n+1))
				}
			}
			continue
		default:
			// ASF extended content is mostly technical (WMFSDK*, IsVBR,
			// EncodingSettings); unknown descriptors are ignored.
			continue
		}

		switch d.Type {
		case 0: // unicode string
			add(key, asf.UTF16String(d.Value))
		case 2: // bool
			if len(d.Value) >= 4 && binary.LittleEndian.Uint32(d.Value) != 0 {
				add(key, "1")
			} else {
				add(key, "0")
			}
		case 3: // DWORD
			if len(d.Value) >= 4 {
				add(key, strconv.FormatUint(uint64(binary.LittleEndian.Uint32(d.Value)), 10))
			}
		case 5: // WORD
			if len(d.Value) >= 2 {
				add(key, strconv.FormatUint(uint64(binary.LittleEndian.Uint16(d.Value)), 10))
			}
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
