// Package riff reads RIFF/WAVE chunks: LIST/INFO fields and ID3v2 tags,
// including tags appended after the RIFF data (common for WAV files).
package riff

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

const maxTagSize = 32 << 20

// Metadata is the parsed RIFF metadata.
type Metadata struct {
	// Info holds LIST/INFO subchunk values keyed by their fourcc (INAM, IART, …).
	Info map[string]string
	// ID3 holds an ID3v2 tag found inside the RIFF stream or appended to it.
	ID3 []byte
}

// ReadFile parses a WAVE file.
func ReadFile(path string) (*Metadata, error) {
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
		return nil, fmt.Errorf("read riff header: %w", err)
	}
	if string(head[0:4]) != "RIFF" || string(head[8:12]) != "WAVE" {
		return nil, fmt.Errorf("not a RIFF/WAVE file")
	}

	declaredEnd := int64(8) + int64(binary.LittleEndian.Uint32(head[4:8]))
	if declaredEnd > size || declaredEnd < 12 {
		declaredEnd = size
	}

	md := &Metadata{Info: make(map[string]string)}

	offset := int64(12)
	header := make([]byte, 8)
	for offset+8 <= declaredEnd {
		if _, err := f.ReadAt(header, offset); err != nil {
			break
		}
		id := string(header[0:4])
		chunkSize := int64(binary.LittleEndian.Uint32(header[4:8]))
		dataStart := offset + 8
		if chunkSize < 0 || dataStart+chunkSize > size {
			break
		}

		switch id {
		case "LIST":
			if chunkSize >= 4 {
				kind := make([]byte, 4)
				if _, err := f.ReadAt(kind, dataStart); err == nil && string(kind) == "INFO" {
					parseInfo(f, dataStart+4, dataStart+chunkSize, md)
				}
			}
		case "id3 ", "ID3 ":
			md.ID3 = readAt(f, dataStart, chunkSize)
		}

		offset = dataStart + chunkSize
		if chunkSize%2 == 1 {
			offset++
		}
	}

	if len(md.ID3) == 0 && declaredEnd < size {
		tailSize := size - declaredEnd
		if tailSize > maxTagSize {
			tailSize = maxTagSize
		}
		tail := readAt(f, declaredEnd, tailSize)
		if idx := bytes.Index(tail, []byte("ID3")); idx >= 0 {
			md.ID3 = tail[idx:]
		}
	}

	return md, nil
}

func parseInfo(f *os.File, start, end int64, md *Metadata) {
	offset := start
	header := make([]byte, 8)
	for offset+8 <= end {
		if _, err := f.ReadAt(header, offset); err != nil {
			return
		}
		id := string(header[0:4])
		size := int64(binary.LittleEndian.Uint32(header[4:8]))
		dataStart := offset + 8
		if size < 0 || dataStart+size > end {
			return
		}
		if isPrintableFourCC(id) {
			value := readAt(f, dataStart, size)
			text := strings.TrimSpace(strings.TrimRight(string(value), "\x00"))
			if text != "" {
				md.Info[id] = text
			}
		}
		offset = dataStart + size
		if size%2 == 1 {
			offset++
		}
	}
}

func isPrintableFourCC(id string) bool {
	for _, r := range id {
		if r < 0x20 || r > 0x7E {
			return false
		}
	}
	return true
}

func readAt(f *os.File, offset, size int64) []byte {
	if size <= 0 || size > maxTagSize {
		return nil
	}
	data := make([]byte, size)
	n, err := f.ReadAt(data, offset)
	if err != nil && err != io.EOF {
		return nil
	}
	return data[:n]
}
