// Package id3 writes ID3v1 and ID3v2.2/2.3/2.4 tags.
package id3

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf16"
)

// Cover is an embedded front cover.
type Cover struct {
	Filename string
	MIME     string
	Data     []byte
}

// Options controls the writer.
type Options struct {
	Version int      // ID3v2 major version: 2, 3 or 4 (default 3)
	V1      bool     // also write an ID3v1 tag
	Replace bool     // drop existing frames that are not in fields
	Only    []string // selected fields; absent ones are removed
}

type rawFrame struct {
	id      string
	payload []byte
}

// Build returns a standalone ID3v2 tag for fields and artwork.
func Build(fields map[string]string, covers map[string]*Cover, version int) ([]byte, error) {
	if version != 2 && version != 4 {
		version = 3
	}
	return composeTag(version, buildFrames(version, fields, covers)), nil
}

// Write replaces the ID3v2 tag of path with fields and cover.
func Write(path string, fields map[string]string, covers map[string]*Cover, opts Options) error {
	version := opts.Version
	if version != 2 && version != 4 {
		version = 3
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	existing, oldVersion, tagLen, hasTag, err := parseTag(data)
	if err != nil {
		if !opts.Replace {
			return fmt.Errorf("parse existing id3: %w", err)
		}
		existing, tagLen, hasTag = nil, 0, false
	}

	frames := buildFrames(version, fields, covers)
	if !opts.Replace && hasTag && oldVersion == version {
		overwrite := frameIDs(frames)
		artOnly := onlyArtworkTypes(opts.Only)
		if len(artOnly) > 0 {
			delete(overwrite, apicID(version))
		}
		removeIDs, removeTXXX := onlyRemovals(version, opts.Only)
		for _, frame := range existing {
			if frame.id == apicID(version) {
				if artOnly[apicKind(version, frame.payload)] {
					continue
				}
				frames = append(frames, frame)
				continue
			}
			if overwrite[frame.id] || removeIDs[frame.id] {
				continue
			}
			if frame.id == txxxID(version) && removeTXXX[strings.ToLower(txxxDescription(frame.payload))] {
				continue
			}
			frames = append(frames, frame)
		}
	}

	tag := composeTag(version, frames)

	rest := data
	if hasTag {
		rest = data[tagLen:]
	}
	if opts.V1 {
		rest = stripV1(rest)
		rest = append(rest, buildV1(fields)...)
	}

	out := make([]byte, 0, len(tag)+len(rest))
	out = append(out, tag...)
	out = append(out, rest...)
	return os.WriteFile(path, out, info.Mode().Perm())
}

func buildFrames(version int, fields map[string]string, covers map[string]*Cover) []rawFrame {
	ids := frameIDsFor(version)
	var frames []rawFrame
	var unknown []string
	for field, value := range fields {
		if value == "" {
			continue
		}
		switch field {
		case "Year":
			payload := textPayload(value, encodingFor(version))
			frames = append(frames, rawFrame{id: ids["Year"], payload: payload})
		case "Comment":
			frames = append(frames, rawFrame{id: commentID(version), payload: commentPayload(value, encodingFor(version))})
		default:
			if id, ok := ids[field]; ok {
				frames = append(frames, rawFrame{id: id, payload: textPayload(value, encodingFor(version))})
				continue
			}
			unknown = append(unknown, field)
		}
	}
	for _, field := range unknown {
		frames = append(frames, rawFrame{id: txxxID(version), payload: txxxPayload(field, fields[field], encodingFor(version))})
	}
	for _, art := range artworkTypes {
		cover := covers[art.kind]
		if cover == nil || len(cover.Data) == 0 {
			continue
		}
		frames = append(frames, rawFrame{
			id:      apicID(version),
			payload: apicPayload(cover, encodingFor(version), version, art.pictureType),
		})
	}
	return frames
}

var artworkTypes = []struct {
	kind        string
	pictureType byte
}{
	{"front", 3},
	{"back", 4},
	{"artist", 8},
	{"disc", 6},
	{"icon", 1},
}

func encodingFor(version int) byte {
	if version == 4 {
		return 3 // UTF-8
	}
	return 1 // UTF-16 with BOM
}

func onlyRemovals(version int, only []string) (map[string]bool, map[string]bool) {
	ids := frameIDsFor(version)
	removeIDs := make(map[string]bool)
	removeTXXX := make(map[string]bool)
	for _, field := range only {
		if artworkKind(field) != "" {
			continue
		}
		if strings.EqualFold(field, "Comment") {
			removeIDs[commentID(version)] = true
			continue
		}
		if id, ok := ids[field]; ok {
			removeIDs[id] = true
			continue
		}
		removeTXXX[strings.ToLower(field)] = true
	}
	return removeIDs, removeTXXX
}

func txxxDescription(payload []byte) string {
	if len(payload) < 2 {
		return ""
	}
	encoding := payload[0]
	body := payload[1:]
	switch encoding {
	case 1, 2:
		var u []uint16
		for i := 0; i+1 < len(body); i += 2 {
			v := uint16(body[i]) | uint16(body[i+1])<<8
			if encoding == 2 {
				v = uint16(body[i])<<8 | uint16(body[i+1])
			}
			if v == 0 {
				break
			}
			u = append(u, v)
		}
		if encoding == 2 {
			return string(utf16.Decode(u))
		}
		if len(u) > 0 && u[0] == 0xFEFF {
			u = u[1:]
		}
		return string(utf16.Decode(u))
	case 3:
		if i := bytes.IndexByte(body, 0); i >= 0 {
			return string(body[:i])
		}
		return string(body)
	default:
		if i := bytes.IndexByte(body, 0); i >= 0 {
			return string(body[:i])
		}
		return string(body)
	}
}

func artworkKind(field string) string {
	names := map[string]string{
		"cover art (front)": "front",
		"cover art (back)":  "back",
		"artist picture":    "artist",
		"cover art (disc)":  "disc",
		"cover art (icon)":  "icon",
	}
	return names[strings.ToLower(strings.TrimSpace(field))]
}

func onlyArtworkTypes(fields []string) map[string]bool {
	out := make(map[string]bool)
	for _, field := range fields {
		if kind := artworkKind(field); kind != "" {
			out[kind] = true
		}
	}
	return out
}

func apicKind(version int, payload []byte) string {
	if len(payload) < 2 {
		return ""
	}
	pos := 1
	if version == 2 {
		pos += 3
	} else {
		for pos < len(payload) && payload[pos] != 0 {
			pos++
		}
		pos++
	}
	if pos >= len(payload) {
		return ""
	}
	switch payload[pos] {
	case 0, 3:
		return "front"
	case 4:
		return "back"
	case 8:
		return "artist"
	case 6:
		return "disc"
	case 1:
		return "icon"
	}
	return ""
}
func frameIDsFor(version int) map[string]string {
	if version == 2 {
		return map[string]string{
			"Title": "TT2", "Artist": "TP1", "Album": "TAL", "Album artist": "TP2",
			"Track": "TRK", "Disc": "TPA", "Genre": "TCO", "Composer": "TCM",
			"Publisher": "TP4", "Encoder": "TSS", "Year": "TYE",
		}
	}
	ids := map[string]string{
		"Title": "TIT2", "Artist": "TPE1", "Album": "TALB", "Album artist": "TPE2",
		"Track": "TRCK", "Disc": "TPOS", "Genre": "TCON", "Composer": "TCOM",
		"Publisher": "TPUB", "Encoder": "TSSE", "Year": "TYER",
	}
	if version == 4 {
		ids["Year"] = "TDRC"
	}
	return ids
}

func commentID(version int) string {
	if version == 2 {
		return "COM"
	}
	return "COMM"
}

func txxxID(version int) string {
	if version == 2 {
		return "TXX"
	}
	return "TXXX"
}

func apicID(version int) string {
	if version == 2 {
		return "PIC"
	}
	return "APIC"
}

func frameIDs(frames []rawFrame) map[string]bool {
	out := make(map[string]bool, len(frames))
	for _, f := range frames {
		out[f.id] = true
	}
	return out
}

func composeTag(version int, frames []rawFrame) []byte {
	body := make([]byte, 0, 1024)
	for _, f := range frames {
		body = append(body, encodeFrame(version, f)...)
	}
	tag := make([]byte, 10, 10+len(body))
	copy(tag, "ID3")
	tag[3] = byte(version)
	tag[4] = 0
	tag[5] = 0
	copy(tag[6:10], synchsafe(len(body)))
	return append(tag, body...)
}

func encodeFrame(version int, f rawFrame) []byte {
	if version == 2 {
		frame := make([]byte, 6, 6+len(f.payload))
		copy(frame, f.id)
		frame[3] = byte(len(f.payload) >> 16)
		frame[4] = byte(len(f.payload) >> 8)
		frame[5] = byte(len(f.payload))
		return append(frame, f.payload...)
	}
	frame := make([]byte, 10, 10+len(f.payload))
	copy(frame, f.id)
	if version == 4 {
		copy(frame[4:8], synchsafe(len(f.payload)))
	} else {
		binary.BigEndian.PutUint32(frame[4:8], uint32(len(f.payload)))
	}
	return append(frame, f.payload...)
}

func textPayload(value string, encoding byte) []byte {
	out := []byte{encoding}
	out = append(out, encodedTextTerm(value, encoding)...)
	return out
}

func encodedTextTerm(value string, encoding byte) []byte {
	out := encodedText(value, encoding)
	return append(out, terminator(encoding)...)
}

func commentPayload(value string, encoding byte) []byte {
	out := []byte{encoding}
	out = append(out, 'e', 'n', 'g')
	out = append(out, emptyText(encoding)...)
	return append(out, encodedTextTerm(value, encoding)...)
}

func txxxPayload(description, value string, encoding byte) []byte {
	out := []byte{encoding}
	out = append(out, encodedText(description, encoding)...)
	out = append(out, terminator(encoding)...)
	return append(out, encodedText(value, encoding)...)
}

func apicPayload(cover *Cover, encoding byte, version int, pictureType byte) []byte {
	out := []byte{encoding}
	if version == 2 {
		format := "JPG"
		if strings.EqualFold(cover.MIME, "image/png") {
			format = "PNG"
		}
		out = append(out, format...)
	} else {
		out = append(out, cover.MIME...)
		out = append(out, 0)
	}
	out = append(out, pictureType)
	out = append(out, emptyText(encoding)...)
	return append(out, cover.Data...)
}

func emptyText(encoding byte) []byte {
	if encoding == 3 {
		return []byte{0}
	}
	return []byte{0xFF, 0xFE, 0, 0}
}

func terminator(encoding byte) []byte {
	if encoding == 3 {
		return []byte{0}
	}
	return []byte{0, 0}
}

func encodedText(value string, encoding byte) []byte {
	if encoding == 3 {
		return []byte(value)
	}
	out := []byte{0xFF, 0xFE}
	return appendUTF16(out, value)
}

func appendUTF16(dst []byte, s string) []byte {
	for _, r := range utf16.Encode([]rune(s)) {
		dst = append(dst, byte(r), byte(r>>8))
	}
	return dst
}

func synchsafe(n int) []byte {
	return []byte{byte(n >> 21 & 0x7F), byte(n >> 14 & 0x7F), byte(n >> 7 & 0x7F), byte(n & 0x7F)}
}

func parseSynchsafe(b []byte) int {
	return int(b[0]&0x7F)<<21 | int(b[1]&0x7F)<<14 | int(b[2]&0x7F)<<7 | int(b[3]&0x7F)
}

func parseTag(data []byte) ([]rawFrame, int, int, bool, error) {
	if len(data) < 10 || string(data[0:3]) != "ID3" {
		return nil, 0, 0, false, nil
	}
	version := int(data[3])
	if version < 2 || version > 4 {
		return nil, 0, 0, true, fmt.Errorf("unsupported ID3v2.%d", version)
	}
	flags := data[5]
	size := parseSynchsafe(data[6:10])
	total := 10 + size
	if total > len(data) {
		return nil, 0, 0, true, fmt.Errorf("tag size %d exceeds file", size)
	}
	if flags&0x80 != 0 {
		return nil, 0, 0, true, fmt.Errorf("unsynchronised tags are not supported")
	}
	offset := 10
	if flags&0x40 != 0 {
		switch version {
		case 3:
			if offset+4 > total {
				return nil, 0, 0, true, fmt.Errorf("truncated extended header")
			}
			ext := int(binary.BigEndian.Uint32(data[offset : offset+4]))
			offset += 4 + ext
		case 4:
			if offset+4 > total {
				return nil, 0, 0, true, fmt.Errorf("truncated extended header")
			}
			ext := parseSynchsafe(data[offset : offset+4])
			offset += ext
		}
	}

	var frames []rawFrame
	for offset < total {
		if data[offset] == 0 {
			break
		}
		var id string
		var frameSize, headerLen int
		switch version {
		case 2:
			if offset+6 > total {
				return nil, 0, 0, true, fmt.Errorf("truncated frame header")
			}
			id = string(data[offset : offset+3])
			frameSize = int(data[offset+3])<<16 | int(data[offset+4])<<8 | int(data[offset+5])
			headerLen = 6
		case 3:
			if offset+10 > total {
				return nil, 0, 0, true, fmt.Errorf("truncated frame header")
			}
			id = string(data[offset : offset+4])
			frameSize = int(binary.BigEndian.Uint32(data[offset+4 : offset+8]))
			headerLen = 10
		case 4:
			if offset+10 > total {
				return nil, 0, 0, true, fmt.Errorf("truncated frame header")
			}
			id = string(data[offset : offset+4])
			frameSize = parseSynchsafe(data[offset+4 : offset+8])
			headerLen = 10
		}
		if frameSize < 0 || offset+headerLen+frameSize > total {
			return nil, 0, 0, true, fmt.Errorf("frame %s is truncated", id)
		}
		payload := append([]byte(nil), data[offset+headerLen:offset+headerLen+frameSize]...)
		frames = append(frames, rawFrame{id: id, payload: payload})
		offset += headerLen + frameSize
	}
	return frames, version, total, true, nil
}

func stripV1(data []byte) []byte {
	if len(data) >= 128 && string(data[len(data)-128:len(data)-125]) == "TAG" {
		return data[:len(data)-128]
	}
	return data
}

var genres = []string{
	"Blues", "Classic Rock", "Country", "Dance", "Disco", "Funk", "Grunge", "Hip-Hop",
	"Jazz", "Metal", "New Age", "Oldies", "Other", "Pop", "R&B", "Rap", "Reggae", "Rock",
	"Techno", "Industrial", "Alternative", "Ska", "Death Metal", "Pranks", "Soundtrack",
	"Euro-Techno", "Ambient", "Trip-Hop", "Vocal", "Jazz+Funk", "Fusion", "Trance",
	"Classical", "Instrumental", "Acid", "House", "Game", "Sound Clip", "Gospel", "Noise",
	"AlternRock", "Bass", "Soul", "Punk", "Space", "Meditative", "Instrumental Pop",
	"Instrumental Rock", "Ethnic", "Gothic", "Darkwave", "Techno-Industrial", "Electronic",
	"Pop-Folk", "Eurodance", "Dream", "Southern Rock", "Comedy", "Cult", "Gangsta",
	"Top 40", "Christian Rap", "Pop/Funk", "Jungle", "Native American", "Cabaret",
	"New Wave", "Psychadelic", "Rave", "Showtunes", "Trailer", "Lo-Fi", "Tribal",
	"Acid Punk", "Acid Jazz", "Polka", "Retro", "Musical", "Rock & Roll", "Hard Rock",
	"Folk", "Folk-Rock", "National Folk", "Swing", "Fast Fusion", "Bebob", "Latin",
	"Revival", "Celtic", "Bluegrass", "Avantgarde", "Gothic Rock", "Progressive Rock",
	"Psychedelic Rock", "Symphonic Rock", "Slow Rock", "Big Band", "Chorus", "Easy Listening",
	"Acoustic", "Humour", "Speech", "Chanson", "Opera", "Chamber Music", "Sonata",
	"Symphony", "Booty Bass", "Primus", "Porn Groove", "Satire", "Slow Jam", "Club",
	"Tango", "Samba", "Folklore", "Ballad", "Power Ballad", "Rhythmic Soul", "Freestyle",
	"Duet", "Punk Rock", "Drum Solo", "A capella", "Euro-House", "Dance Hall",
}

func buildV1(fields map[string]string) []byte {
	out := make([]byte, 128)
	copy(out[0:3], "TAG")
	copy(out[3:33], latin1(fields["Title"], 30))
	copy(out[33:63], latin1(fields["Artist"], 30))
	copy(out[63:93], latin1(fields["Album"], 30))
	copy(out[93:97], latin1(fields["Year"], 4))
	track, _ := strconv.Atoi(strings.Split(fields["Track"], "/")[0])
	if track > 0 {
		copy(out[97:125], latin1(fields["Comment"], 28))
		out[125] = 0
		out[126] = byte(track)
	} else {
		copy(out[97:127], latin1(fields["Comment"], 30))
	}
	genre := byte(255)
	for i, name := range genres {
		if strings.EqualFold(name, fields["Genre"]) {
			genre = byte(i)
			break
		}
	}
	out[127] = genre
	return out
}

func latin1(s string, n int) []byte {
	out := make([]byte, 0, n)
	for _, r := range s {
		b := byte('?')
		if r < 256 {
			b = byte(r)
		}
		out = append(out, b)
		if len(out) == n {
			break
		}
	}
	return out
}
