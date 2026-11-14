package id3

import (
	"os"
	"strings"
)

// ReadFilePictures extracts all APIC/PIC frames from an ID3v2 tag at the start of path.
func ReadFilePictures(path string) []*Cover {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return ReadPictures(data)
}

// ReadCovers extracts all APIC/PIC frames from an ID3v2 tag.
func ReadCovers(data []byte) map[string]*Cover {
	out := make(map[string]*Cover)
	for _, cover := range ReadPictures(data) {
		if kind := pictureKind(cover.Type); kind != "" {
			out[kind] = cover
		}
	}
	return out
}

// ReadPictures returns all APIC/PIC frames in native order.
func ReadPictures(data []byte) []*Cover {
	frames, _, _, _, err := parseTag(data)
	if err != nil {
		return nil
	}
	var out []*Cover
	for _, frame := range frames {
		if frame.id != "APIC" && frame.id != "PIC" {
			continue
		}
		if _, cover := parseAPIC(frame.id == "PIC", frame.payload); cover != nil && len(cover.Data) > 0 {
			cover.Type = apicType(frame.id == "PIC", frame.payload)
			out = append(out, cover)
		}
	}
	return out
}

func pictureKind(pictureType byte) string {
	switch pictureType {
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

func apicType(v22 bool, payload []byte) byte {
	if len(payload) < 2 {
		return 0
	}
	pos := 1
	if v22 {
		pos = 4
	} else {
		for pos < len(payload) && payload[pos] != 0 {
			pos++
		}
		pos++
	}
	if pos < len(payload) {
		return payload[pos]
	}
	return 0
}

func parseAPIC(v22 bool, payload []byte) (string, *Cover) {
	if len(payload) < 2 {
		return "", nil
	}
	encoding := payload[0]
	pos := 1
	mime := ""
	if v22 {
		if len(payload) < 5 {
			return "", nil
		}
		format := strings.ToUpper(string(payload[1:4]))
		switch format {
		case "PNG":
			mime = "image/png"
		case "GIF":
			mime = "image/gif"
		default:
			mime = "image/jpeg"
		}
		pos = 4
	} else {
		end := pos
		for end < len(payload) && payload[end] != 0 {
			end++
		}
		if end >= len(payload) {
			return "", nil
		}
		mime = string(payload[pos:end])
		pos = end + 1
	}
	if pos >= len(payload) {
		return "", nil
	}
	kind := apicKind(0, payload)
	pos++
	pos = skipDescription(encoding, payload, pos)
	if pos > len(payload) {
		return "", nil
	}
	return kind, &Cover{Filename: coverFilename(mime), MIME: mime, Data: payload[pos:]}
}

func skipDescription(encoding byte, payload []byte, pos int) int {
	if encoding == 1 || encoding == 2 {
		for pos+1 < len(payload) {
			if payload[pos] == 0 && payload[pos+1] == 0 {
				return pos + 2
			}
			pos += 2
		}
		return len(payload)
	}
	for pos < len(payload) && payload[pos] != 0 {
		pos++
	}
	if pos < len(payload) {
		pos++
	}
	return pos
}

func coverFilename(mime string) string {
	switch strings.ToLower(mime) {
	case "image/png":
		return "cover.png"
	case "image/gif":
		return "cover.gif"
	case "image/webp":
		return "cover.webp"
	}
	return "cover.jpg"
}
