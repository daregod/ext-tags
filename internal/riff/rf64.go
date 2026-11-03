package riff

import (
	"encoding/binary"
	"fmt"
	"os"
)

// ReadRF64 parses an RF64 file (RIFF with 64-bit sizes).
func ReadRF64(path string) (*Metadata, error) {
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
		return nil, fmt.Errorf("read rf64 header: %w", err)
	}
	if string(head[0:4]) != "RF64" || string(head[8:12]) != "WAVE" {
		return nil, fmt.Errorf("not an RF64 file")
	}

	md := &Metadata{Info: make(map[string]string)}
	var dataSize int64

	header := make([]byte, 8)
	offset := int64(12)
	for offset+8 <= size {
		if _, err := f.ReadAt(header, offset); err != nil {
			break
		}
		id := string(header[0:4])
		chunkSize := int64(binary.LittleEndian.Uint32(header[4:8]))
		dataStart := offset + 8

		switch id {
		case "ds64":
			// riffSize, dataSize, sampleCount are 64-bit.
			payload := readAt(f, dataStart, chunkSize)
			if len(payload) >= 24 {
				dataSize = int64(binary.LittleEndian.Uint64(payload[8:16]))
			}
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

		if chunkSize == 0xFFFFFFFF {
			// The data chunk stores its real size in ds64.
			if id == "data" {
				if dataSize <= 0 {
					break
				}
				offset = dataStart + dataSize
				if dataSize%2 == 1 {
					offset++
				}
				continue
			}
			chunkSize = 0
		}
		if chunkSize < 0 || dataStart+chunkSize > size {
			break
		}
		offset = dataStart + chunkSize
		if chunkSize%2 == 1 {
			offset++
		}
	}

	return md, nil
}
