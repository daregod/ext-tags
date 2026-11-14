package id3

import (
	"strings"
)

// ReadCovers extracts all APIC/PIC frames from an ID3v2 tag.
func ReadCovers(data []byte) map[string]*Cover {
	out := make(map[string]*Cover)
	frames, _, _, _, err := parseTag(data)
	if err != nil {
		return out
	}
	for _, frame := range frames {
		if frame.id != "APIC" && frame.id != "PIC" {
			continue
		}
		kind, cover := parseAPIC(frame.id == "PIC", frame.payload)
		if kind != "" && cover != nil && len(cover.Data) > 0 {
			out[kind] = cover
		}
	}
	return out
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
