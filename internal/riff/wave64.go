package riff

import (
	"encoding/binary"
	"fmt"
	"os"
)

// ReadWave64 parses a Sony Wave64 file: GUID chunks with 64-bit sizes.
// The size of every chunk includes its 24-byte header.
func ReadWave64(path string) (*Metadata, error) {
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

	head := make([]byte, 40)
	if _, err := f.ReadAt(head, 0); err != nil {
		return nil, fmt.Errorf("read wave64 header: %w", err)
	}
	if string(head[0:4]) != "riff" || string(head[24:28]) != "wave" {
		return nil, fmt.Errorf("not a Wave64 file")
	}

	md := &Metadata{Info: make(map[string]string)}
	header := make([]byte, 24)
	offset := int64(40)
	for offset+24 <= size {
		if _, err := f.ReadAt(header, offset); err != nil {
			break
		}
		guid := [16]byte(header[:16])
		chunkSize := int64(binary.LittleEndian.Uint64(header[16:24]))
		if chunkSize < 24 || offset+chunkSize > size {
			break
		}
		id := string(guid[0:4])
		dataStart := offset + 24
		dataLen := chunkSize - 24

		switch id {
		case "ID3 ", "id3 ":
			md.ID3 = readAt(f, dataStart, dataLen)
		case "LIST":
			if dataLen >= 4 {
				kind := make([]byte, 4)
				if _, err := f.ReadAt(kind, dataStart); err == nil && string(kind) == "INFO" {
					parseInfo(f, dataStart+4, dataStart+dataLen, md)
				}
			}
		}

		offset += chunkSize
		offset = (offset + 7) &^ 7
	}
	if len(md.ID3) == 0 && len(md.Info) == 0 {
		return md, nil
	}
	return md, nil
}
