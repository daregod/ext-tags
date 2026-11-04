package riff

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/daregod/ext-tags/internal/formats/id3"
	"github.com/daregod/ext-tags/internal/metadata"
)

// Write stores ID3v2 metadata in a WAV/RF64/Wave64/AIFF file.
func Write(path string, doc *metadata.Document) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	tag, err := buildID3(doc)
	if err != nil {
		return err
	}
	switch {
	case len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WAVE":
		data, err = rewriteRIFF(data, tag, false)
	case len(data) >= 12 && string(data[0:4]) == "RF64" && string(data[8:12]) == "WAVE":
		data, err = rewriteRIFF(data, tag, true)
	case len(data) >= 12 && string(data[0:4]) == "FORM" && (string(data[8:12]) == "AIFF" || string(data[8:12]) == "AIFC"):
		data, err = rewriteIFF(data, tag)
	case len(data) >= 40 && string(data[0:4]) == "riff" && string(data[24:28]) == "wave":
		data, err = rewriteWave64(data, tag)
	default:
		return fmt.Errorf("riff: unsupported container")
	}
	if err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, info.Mode().Perm())
}

func buildID3(doc *metadata.Document) ([]byte, error) {
	covers := make(map[string]*id3.Cover, len(doc.Artwork))
	for kind, cover := range doc.Artwork {
		if cover == nil {
			continue
		}
		covers[kind] = &id3.Cover{Filename: cover.Filename, MIME: cover.MIME, Data: cover.Data}
	}
	if doc.Cover != nil && covers[metadata.Front] == nil {
		covers[metadata.Front] = &id3.Cover{Filename: doc.Cover.Filename, MIME: doc.Cover.MIME, Data: doc.Cover.Data}
	}
	return id3.Build(doc.Fields, covers, 3)
}

func rewriteRIFF(data, tag []byte, rf64 bool) ([]byte, error) {
	declaredEnd := len(data)
	if !rf64 && len(data) >= 8 {
		if end := 8 + int(binary.LittleEndian.Uint32(data[4:8])); end >= 12 && end <= len(data) {
			declaredEnd = end
		}
	}
	var body []byte
	var err error
	if rf64 {
		body, err = stripChunksRF64(data[12:], rf64DataSize(data), "id3 ", "ID3 ", "LIST")
	} else {
		body, err = stripChunksLE(data[12:declaredEnd], "id3 ", "ID3 ", "LIST")
	}
	if err != nil {
		return nil, err
	}
	body = append(body, id3ChunkLE(tag)...)
	out := make([]byte, 12, 12+len(body))
	copy(out[0:4], data[0:4])
	if rf64 {
		binary.LittleEndian.PutUint32(out[4:8], 0xFFFFFFFF)
	} else {
		binary.LittleEndian.PutUint32(out[4:8], uint32(len(body)+4))
	}
	copy(out[8:12], "WAVE")
	out = append(out, body...)
	if rf64 {
		patchDS64(out)
	}
	return out, nil
}

func stripChunksLE(data []byte, skip ...string) ([]byte, error) {
	skipSet := make(map[string]bool, len(skip))
	for _, id := range skip {
		skipSet[id] = true
	}
	var out []byte
	offset := 0
	for offset+8 <= len(data) {
		id := string(data[offset : offset+4])
		size := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		end := offset + 8 + size
		if size%2 == 1 {
			end++
		}
		if size < 0 || end > len(data) {
			return nil, fmt.Errorf("riff: bad chunk %q", id)
		}
		if !skipSet[id] {
			out = append(out, data[offset:end]...)
		}
		offset = end
	}
	if offset != len(data) {
		return nil, fmt.Errorf("riff: trailing bytes")
	}
	return out, nil
}

func id3ChunkLE(tag []byte) []byte {
	size := len(tag)
	pad := size % 2
	out := make([]byte, 8, 8+size+pad)
	copy(out[0:4], "id3 ")
	binary.LittleEndian.PutUint32(out[4:8], uint32(size))
	out = append(out, tag...)
	if pad == 1 {
		out = append(out, 0)
	}
	return out
}

func patchDS64(data []byte) {
	if len(data) < 12 {
		return
	}
	offset := 12
	for offset+8 <= len(data) {
		id := string(data[offset : offset+4])
		size := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		end := offset + 8 + size
		if end > len(data) {
			return
		}
		if id == "ds64" && size >= 16 {
			binary.LittleEndian.PutUint64(data[offset+8:offset+16], uint64(len(data)-8))
			return
		}
		offset = end
		if size%2 == 1 {
			offset++
		}
	}
}

func rewriteIFF(data, tag []byte) ([]byte, error) {
	declaredEnd := len(data)
	if len(data) >= 8 {
		if end := 8 + int(binary.BigEndian.Uint32(data[4:8])); end >= 12 && end <= len(data) {
			declaredEnd = end
		}
	}
	body, err := stripChunksBE(data[12:declaredEnd], "ID3 ", "id3 ", "NAME", "AUTH", "ANNO", "(c) ")
	if err != nil {
		return nil, err
	}
	body = append(body, id3ChunkBE(tag)...)
	out := make([]byte, 12, 12+len(body))
	copy(out[0:4], "FORM")
	binary.BigEndian.PutUint32(out[4:8], uint32(len(body)+4))
	copy(out[8:12], data[8:12])
	return append(out, body...), nil
}

func stripChunksBE(data []byte, skip ...string) ([]byte, error) {
	skipSet := make(map[string]bool, len(skip))
	for _, id := range skip {
		skipSet[id] = true
	}
	var out []byte
	offset := 0
	for offset+8 <= len(data) {
		id := string(data[offset : offset+4])
		size := int(binary.BigEndian.Uint32(data[offset+4 : offset+8]))
		end := offset + 8 + size
		if size%2 == 1 {
			end++
		}
		if end > len(data) {
			return nil, fmt.Errorf("riff: bad IFF chunk %q", id)
		}
		if !skipSet[id] {
			out = append(out, data[offset:end]...)
		}
		offset = end
	}
	if offset != len(data) {
		return nil, fmt.Errorf("riff: trailing bytes")
	}
	return out, nil
}

func id3ChunkBE(tag []byte) []byte {
	size := len(tag)
	pad := size % 2
	out := make([]byte, 8, 8+size+pad)
	copy(out[0:4], "ID3 ")
	binary.BigEndian.PutUint32(out[4:8], uint32(size))
	out = append(out, tag...)
	if pad == 1 {
		out = append(out, 0)
	}
	return out
}

func rewriteWave64(data, tag []byte) ([]byte, error) {
	body, err := stripWave64Chunks(data[40:])
	if err != nil {
		return nil, err
	}
	body = insertWave64ID3(body, wave64ID3Chunk(tag))
	out := append([]byte(nil), data[:40]...)
	out = append(out, body...)
	binary.LittleEndian.PutUint64(out[16:24], uint64(len(out)))
	return out, nil
}

func stripWave64Chunks(data []byte) ([]byte, error) {
	var out []byte
	offset := 0
	for offset+24 <= len(data) {
		id := string(data[offset : offset+4])
		size := int(binary.LittleEndian.Uint64(data[offset+16 : offset+24]))
		if size < 24 || offset+size > len(data) {
			return nil, fmt.Errorf("riff: bad Wave64 chunk %q", id)
		}
		next := (offset + size + 7) &^ 7
		if next > len(data) {
			next = len(data)
		}
		if id != "ID3 " && id != "id3 " && id != "LIST" {
			out = append(out, data[offset:next]...)
		}
		offset = next
	}
	if offset != len(data) {
		return nil, fmt.Errorf("riff: trailing Wave64 bytes")
	}
	return out, nil
}

func wave64ID3Chunk(tag []byte) []byte {
	chunkSize := 24 + len(tag)
	padded := (chunkSize + 7) &^ 7
	out := make([]byte, 24, padded)
	copy(out[0:16], []byte{'I', 'D', '3', ' ', 0xF3, 0xAC, 0xD3, 0x11, 0x8C, 0xD1, 0x00, 0xC0, 0x4F, 0x8E, 0xDB, 0x8A})
	binary.LittleEndian.PutUint64(out[16:24], uint64(chunkSize))
	out = append(out, tag...)
	for len(out) < padded {
		out = append(out, 0)
	}
	return out
}
func rf64DataSize(data []byte) int64 {
	offset := 12
	for offset+8 <= len(data) {
		id := string(data[offset : offset+4])
		size := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		if id == "ds64" && size >= 16 && offset+8+16 <= len(data) {
			return int64(binary.LittleEndian.Uint64(data[offset+8+8 : offset+8+16]))
		}
		offset += 8 + size
		if size%2 == 1 {
			offset++
		}
	}
	return 0
}

func stripChunksRF64(data []byte, dataSize int64, skip ...string) ([]byte, error) {
	skipSet := make(map[string]bool, len(skip))
	for _, id := range skip {
		skipSet[id] = true
	}
	var out []byte
	offset := 0
	for offset+8 <= len(data) {
		id := string(data[offset : offset+4])
		size := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		if id == "data" && size == 0xFFFFFFFF {
			if dataSize <= 0 {
				return nil, fmt.Errorf("riff: rf64 data size missing")
			}
			size = int(dataSize)
		}
		end := offset + 8 + size
		if size%2 == 1 {
			end++
		}
		if end > len(data) {
			return nil, fmt.Errorf("riff: bad rf64 chunk %q", id)
		}
		if !skipSet[id] {
			out = append(out, data[offset:end]...)
		}
		offset = end
	}
	return out, nil
}
func insertWave64ID3(body, chunk []byte) []byte {
	offset := 0
	for offset+24 <= len(body) {
		id := string(body[offset : offset+4])
		size := int(binary.LittleEndian.Uint64(body[offset+16 : offset+24]))
		if size < 24 || offset+size > len(body) {
			break
		}
		if id == "data" {
			out := append([]byte(nil), body[:offset]...)
			out = append(out, chunk...)
			return append(out, body[offset:]...)
		}
		offset = (offset + size + 7) &^ 7
	}
	return append(body, chunk...)
}
