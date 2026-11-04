package riff

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// ReadAIFF parses an AIFF/AIFC file: text chunks (NAME, AUTH, ANNO, (c) )
// and an embedded ID3v2 chunk.
func ReadAIFF(path string) (*Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := info.Size()

	head := make([]byte, 12)
	if _, err := f.ReadAt(head, 0); err != nil {
		return nil, fmt.Errorf("read aiff header: %w", err)
	}
	if string(head[0:4]) != "FORM" {
		return nil, fmt.Errorf("not an AIFF file")
	}
	formType := string(head[8:12])
	if formType != "AIFF" && formType != "AIFC" {
		return nil, fmt.Errorf("unsupported IFF form %q", formType)
	}

	md := &Metadata{Info: make(map[string]string)}
	header := make([]byte, 8)
	offset := int64(12)
	for offset+8 <= size {
		if _, err := f.ReadAt(header, offset); err != nil {
			break
		}
		id := string(header[0:4])
		chunkSize := int64(binary.BigEndian.Uint32(header[4:8]))
		dataStart := offset + 8
		if chunkSize < 0 || dataStart+chunkSize > size {
			break
		}

		switch id {
		case "ID3 ", "id3 ":
			md.ID3 = readAt(f, dataStart, chunkSize)
		case "NAME", "AUTH", "ANNO", "(c) ":
			if chunkSize <= 64<<10 {
				value := readAt(f, dataStart, chunkSize)
				text := strings.TrimSpace(strings.TrimRight(string(value), "\x00"))
				if text != "" {
					md.Info[id] = text
				}
			}
		}

		offset = dataStart + chunkSize
		if chunkSize%2 == 1 {
			offset++
		}
	}
	return md, nil
}
