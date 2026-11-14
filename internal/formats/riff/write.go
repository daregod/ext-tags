package riff

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"github.com/daregod/ext-tags/internal/formats/id3"
	"github.com/daregod/ext-tags/internal/metadata"
)

func onlyInfoFields(only []string) map[string]bool {
	out := make(map[string]bool)
	for _, field := range only {
		switch field {
		case "Title", "Artist", "Album", "Year", "Genre", "Comment", "Track", "Copyright":
			out[strings.ToLower(field)] = true
		}
	}
	return out
}

func filterChunksLE(data []byte, dataSize int64, remove map[string]bool) ([]byte, error) {
	var out []byte
	offset := 0
	for offset+8 <= len(data) {
		id := string(data[offset : offset+4])
		size := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		if id == "data" && size == 0xFFFFFFFF && dataSize > 0 {
			size = int(dataSize)
		}
		end := offset + 8 + size
		if size%2 == 1 {
			end++
		}
		if size < 0 || end > len(data) {
			return nil, fmt.Errorf("riff: bad chunk %q", id)
		}
		if id == "LIST" && size >= 4 && string(data[offset+8:offset+12]) == "INFO" {
			inner := data[offset+12 : offset+8+size]
			filtered := filterInfoLE(inner, remove)
			out = append(out, makeChunkLE("LIST", append([]byte("INFO"), filtered...))...)
		} else {
			out = append(out, data[offset:end]...)
		}
		offset = end
	}
	if offset != len(data) {
		return nil, fmt.Errorf("riff: trailing bytes")
	}
	return out, nil
}

func filterInfoLE(data []byte, remove map[string]bool) []byte {
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
			break
		}
		if key := infoField(id); key == "" || !remove[strings.ToLower(key)] {
			out = append(out, data[offset:end]...)
		}
		offset = end
	}
	return out
}

func makeChunkLE(id string, payload []byte) []byte {
	size := len(payload)
	pad := size % 2
	out := make([]byte, 8, 8+size+pad)
	copy(out[0:4], id)
	binary.LittleEndian.PutUint32(out[4:8], uint32(size))
	out = append(out, payload...)
	if pad == 1 {
		out = append(out, 0)
	}
	return out
}

func filterIFF(data []byte, replace bool, remove map[string]bool) ([]byte, error) {
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
		drop := false
		switch id {
		case "ID3 ", "id3 ":
			drop = true
		case "NAME":
			drop = replace || remove["title"]
		case "AUTH":
			drop = replace || remove["artist"]
		case "ANNO":
			drop = replace || remove["comment"]
		case "(c) ":
			drop = replace || remove["copyright"]
		}
		if !drop {
			out = append(out, data[offset:end]...)
		}
		offset = end
	}
	if offset != len(data) {
		return nil, fmt.Errorf("riff: trailing IFF bytes")
	}
	return out, nil
}

// Write stores ID3v2 metadata in a WAV/RF64/Wave64/AIFF file.
func Write(path string, doc *metadata.Document) error {
	return WriteSelective(path, doc, nil, false)
}

// WriteSelective applies fields/artwork; only removes selected values and
// replace drops all metadata that is not in doc.
func WriteSelective(path string, doc *metadata.Document, only []string, replace bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	existing := existingID3(path, data)
	var tag []byte
	if replace {
		tag, err = buildID3(doc)
	} else if len(only) > 0 {
		tag, err = buildID3Only(doc, existing, only)
	} else {
		tag, err = buildID3Merged(doc, existing)
	}
	if err != nil {
		return err
	}
	removeInfo := onlyInfoFields(only)
	switch {
	case len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WAVE":
		data, err = rewriteRIFF(data, tag, false, replace, removeInfo)
	case len(data) >= 12 && string(data[0:4]) == "RF64" && string(data[8:12]) == "WAVE":
		data, err = rewriteRIFF(data, tag, true, replace, removeInfo)
	case len(data) >= 12 && string(data[0:4]) == "FORM" && (string(data[8:12]) == "AIFF" || string(data[8:12]) == "AIFC"):
		data, err = rewriteIFF(data, tag, replace, removeInfo)
	case len(data) >= 40 && string(data[0:4]) == "riff" && string(data[24:28]) == "wave":
		data, err = rewriteWave64(data, tag, replace, removeInfo)
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
	return buildID3Merged(doc, nil)
}

func buildID3Only(doc *metadata.Document, existing []byte, only []string) ([]byte, error) {
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
	return id3.BuildMergedOnly(existing, doc.Fields, covers, 3, only)
}

func buildID3Merged(doc *metadata.Document, existing []byte) ([]byte, error) {
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
	return id3.BuildMerged(existing, doc.Fields, covers, 3)
}

func existingID3(path string, data []byte) []byte {
	if len(data) < 12 {
		return nil
	}
	switch {
	case string(data[0:4]) == "RIFF" && string(data[8:12]) == "WAVE":
		if md, err := ReadFile(path); err == nil {
			return md.ID3
		}
	case string(data[0:4]) == "RF64" && string(data[8:12]) == "WAVE":
		if md, err := ReadRF64(path); err == nil {
			return md.ID3
		}
	case string(data[0:4]) == "FORM" && (string(data[8:12]) == "AIFF" || string(data[8:12]) == "AIFC"):
		if md, err := ReadAIFF(path); err == nil {
			return md.ID3
		}
	case len(data) >= 28 && string(data[0:4]) == "riff" && string(data[24:28]) == "wave":
		if md, err := ReadWave64(path); err == nil {
			return md.ID3
		}
	}
	return nil
}
func rewriteRIFF(data, tag []byte, rf64, replace bool, remove map[string]bool) ([]byte, error) {
	declaredEnd := len(data)
	if !rf64 && len(data) >= 8 {
		if end := 8 + int(binary.LittleEndian.Uint32(data[4:8])); end >= 12 && end <= len(data) {
			declaredEnd = end
		}
	}
	var body []byte
	var err error
	bodyData := data[12:declaredEnd]
	dataSize := int64(0)
	if rf64 {
		bodyData = data[12:]
		dataSize = rf64DataSize(data)
	}
	if replace {
		if rf64 {
			body, err = stripChunksRF64(bodyData, dataSize, "id3 ", "ID3 ", "LIST")
		} else {
			body, err = stripChunksLE(bodyData, "id3 ", "ID3 ", "LIST")
		}
	} else if len(remove) > 0 {
		body, err = filterChunksLE(bodyData, dataSize, remove)
	} else if rf64 {
		body, err = stripChunksRF64(bodyData, dataSize, "id3 ", "ID3 ")
	} else {
		body, err = stripChunksLE(bodyData, "id3 ", "ID3 ")
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

func rewriteIFF(data, tag []byte, replace bool, remove map[string]bool) ([]byte, error) {
	declaredEnd := len(data)
	if len(data) >= 8 {
		if end := 8 + int(binary.BigEndian.Uint32(data[4:8])); end >= 12 && end <= len(data) {
			declaredEnd = end
		}
	}
	body, err := filterIFF(data[12:declaredEnd], replace, remove)
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

func rewriteWave64(data, tag []byte, replace bool, remove map[string]bool) ([]byte, error) {
	if len(remove) > 0 {
		return nil, fmt.Errorf("riff: --only is not supported for Wave64 yet")
	}
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
		if id != "ID3 " && id != "id3 " {
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
