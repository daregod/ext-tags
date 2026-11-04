// Package ogg reads Vorbis-style comments from Ogg streams that
// dhowden/tag does not support: Speex and FLAC-in-Ogg.
package ogg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

// Tag is a single comment.
type Tag struct {
	Name  string
	Value string
}

// Picture is an embedded cover image.
type Picture struct {
	MIME string
	Data []byte
}

func readPrefix(path string, limit int) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := make([]byte, limit)
	n, err := f.Read(buf)
	if n == 0 && err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// Extension returns the canonical extension for an Ogg stream, or "".
func Extension(path string) string {
	data, err := readPrefix(path, 64<<10)
	if err != nil {
		return ""
	}
	packets := packetsOf(data)
	if len(packets) == 0 {
		return ""
	}
	first := packets[0]
	switch {
	case len(first) >= 8 && string(first[:8]) == "OpusHead":
		return ".opus"
	case len(first) >= 7 && first[0] == 1 && string(first[1:7]) == "vorbis":
		return ".ogg"
	case len(first) >= 8 && string(first[:8]) == "Speex   ":
		return ".spx"
	case len(first) >= 5 && first[0] == 0x7F && string(first[1:5]) == "FLAC":
		return ".oga"
	}
	return ""
}

// ReadTags parses comments of a Speex or Ogg-FLAC stream.
func ReadTags(path string) ([]Tag, *Picture, error) {
	data, err := readPrefix(path, 4<<20)
	if err != nil {
		return nil, nil, err
	}
	packets := packetsOf(data)
	if len(packets) == 0 {
		return nil, nil, fmt.Errorf("no ogg packets")
	}

	first := packets[0]
	switch {
	case bytes.HasPrefix(first, []byte("Speex   ")):
		if len(packets) < 2 {
			return nil, nil, nil
		}
		tags, err := parseCommentBlock(packets[1])
		return tags, nil, err
	case len(first) > 5 && first[0] == 0x7F && string(first[1:5]) == "FLAC":
		return parseFLAC(packets[1:])
	}
	return nil, nil, fmt.Errorf("unsupported ogg codec")
}

// packetsOf reassembles packets of the first logical stream.
func packetsOf(data []byte) [][]byte {
	var (
		packets [][]byte
		partial []byte
		primary uint32
		have    bool
	)
	offset := 0
	for offset+27 <= len(data) {
		if string(data[offset:offset+4]) != "OggS" {
			break
		}
		headerType := data[offset+5]
		_ = headerType
		serial := binary.LittleEndian.Uint32(data[offset+14 : offset+18])
		segments := int(data[offset+26])
		offset += 27
		if offset+segments > len(data) {
			break
		}
		lacing := data[offset : offset+segments]
		offset += segments
		if !have {
			primary = serial
			have = true
		}
		for _, seg := range lacing {
			size := int(seg)
			if offset+size > len(data) {
				return packets
			}
			if serial == primary {
				partial = append(partial, data[offset:offset+size]...)
			}
			offset += size
			if seg < 255 {
				if serial == primary {
					packets = append(packets, partial)
					partial = nil
				}
			}
		}
	}
	return packets
}

func parseFLAC(packets [][]byte) ([]Tag, *Picture, error) {
	var (
		tags []Tag
		pic  *Picture
	)
	for _, packet := range packets {
		if len(packet) < 4 {
			continue
		}
		blockType := packet[0] & 0x7F
		length := int(packet[1])<<16 | int(packet[2])<<8 | int(packet[3])
		if 4+length > len(packet) {
			continue
		}
		block := packet[4 : 4+length]
		switch blockType {
		case 4:
			parsed, err := parseCommentBlock(block)
			if err != nil {
				return nil, nil, err
			}
			tags = append(tags, parsed...)
		case 6:
			if p, ok := parseFlacPicture(block); ok {
				pic = p
			}
		}
	}
	return tags, pic, nil
}

func parseCommentBlock(data []byte) ([]Tag, error) {
	if len(data) < 8 {
		return nil, nil
	}
	offset := 0
	vendorLen := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	if vendorLen < 0 || offset+vendorLen+4 > len(data) {
		return nil, nil
	}
	offset += vendorLen
	count := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4

	var tags []Tag
	for i := 0; i < count; i++ {
		if offset+4 > len(data) {
			break
		}
		length := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
		offset += 4
		if length < 0 || offset+length > len(data) {
			break
		}
		line := string(data[offset : offset+length])
		offset += length
		if eq := stringsIndexByte(line, '='); eq > 0 {
			tags = append(tags, Tag{Name: line[:eq], Value: line[eq+1:]})
		}
	}
	return tags, nil
}

func parseFlacPicture(data []byte) (*Picture, bool) {
	if len(data) < 8 {
		return nil, false
	}
	offset := 0
	// picture type
	offset += 4
	mimeLen := int(binary.BigEndian.Uint32(data[offset : offset+4]))
	offset += 4
	if offset+mimeLen+4 > len(data) {
		return nil, false
	}
	mime := string(data[offset : offset+mimeLen])
	offset += mimeLen
	descLen := int(binary.BigEndian.Uint32(data[offset : offset+4]))
	offset += 4
	if offset+descLen+20 > len(data) {
		return nil, false
	}
	offset += descLen
	offset += 16 // width, height, depth, colors
	dataLen := int(binary.BigEndian.Uint32(data[offset : offset+4]))
	offset += 4
	if dataLen < 0 || offset+dataLen > len(data) {
		return nil, false
	}
	return &Picture{MIME: mime, Data: data[offset : offset+dataLen]}, true
}

func stringsIndexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}
