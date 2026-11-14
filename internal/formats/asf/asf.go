// Package asf reads metadata from ASF containers (WMV, WMA).
//
// Supported: Content Description Object (title, author, copyright,
// description, rating) and Extended Content Description descriptors
// (WM/AlbumTitle, WM/Genre, WM/Picture and friends).
package asf

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"unicode/utf16"
)

var (
	guidHeader          = [16]byte{0x30, 0x26, 0xB2, 0x75, 0x8E, 0x66, 0xCF, 0x11, 0xA6, 0xD9, 0x00, 0xAA, 0x00, 0x62, 0xCE, 0x6C}
	guidContentDesc     = [16]byte{0x33, 0x26, 0xB2, 0x75, 0x8E, 0x66, 0xCF, 0x11, 0xA6, 0xD9, 0x00, 0xAA, 0x00, 0x62, 0xCE, 0x6C}
	guidExtendedContent = [16]byte{0x40, 0xA4, 0xD0, 0xD2, 0x07, 0xE3, 0xD2, 0x11, 0x97, 0xF0, 0x00, 0xA0, 0xC9, 0x5E, 0xA8, 0x50}
	guidHeaderExtension = [16]byte{0xB5, 0x03, 0xBF, 0x5F, 0x2E, 0xA9, 0xCF, 0x11, 0x8E, 0xE3, 0x00, 0xC0, 0x0C, 0x20, 0x53, 0x65}
)

const (
	maxHeaderSize = 64 << 20
	objectHeader  = 24
)

// Descriptor is an extended-content entry.
type Descriptor struct {
	Name  string
	Type  uint16
	Value []byte
}

// Metadata is the parsed ASF metadata.
type Metadata struct {
	Title       string
	Author      string
	Copyright   string
	Description string
	Rating      string
	Descriptors []Descriptor
}

// ReadFile parses an ASF/WMV/WMA file.
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

	head := make([]byte, 30)
	if _, err := f.ReadAt(head, 0); err != nil {
		return nil, fmt.Errorf("read asf header: %w", err)
	}
	if [16]byte(head[0:16]) != guidHeader {
		return nil, fmt.Errorf("not an ASF file")
	}
	headerSize := int64(binary.LittleEndian.Uint64(head[16:24]))
	if headerSize < 30 || headerSize > info.Size() || headerSize > maxHeaderSize {
		return nil, fmt.Errorf("invalid ASF header size %d", headerSize)
	}

	data := make([]byte, headerSize)
	if _, err := f.ReadAt(data, 0); err != nil {
		return nil, fmt.Errorf("read asf header: %w", err)
	}

	md := &Metadata{}
	parseObjects(data, 30, int(headerSize), md, 0)
	return md, nil
}

func parseObjects(data []byte, start, end int, md *Metadata, depth int) {
	if depth > 8 {
		return
	}
	offset := start
	for offset+objectHeader <= end {
		guid := [16]byte(data[offset : offset+16])
		size := int64(binary.LittleEndian.Uint64(data[offset+16 : offset+24]))
		if size < objectHeader || size > int64(end-offset) {
			return
		}
		payload := data[offset+objectHeader : offset+int(size)]

		switch guid {
		case guidContentDesc:
			parseContentDescription(payload, md)
		case guidExtendedContent:
			parseExtendedContent(payload, md)
		case guidHeaderExtension:
			parseHeaderExtension(payload, md, depth)
		}
		offset += int(size)
	}
}

func parseHeaderExtension(payload []byte, md *Metadata, depth int) {
	if len(payload) < 22 {
		return
	}
	dataSize := int(binary.LittleEndian.Uint32(payload[18:22]))
	if dataSize <= 0 || 22+dataSize > len(payload) {
		return
	}
	parseObjects(payload, 22, 22+dataSize, md, depth+1)
}

func parseContentDescription(payload []byte, md *Metadata) {
	if len(payload) < 10 {
		return
	}
	lens := [5]int{}
	for i := range lens {
		lens[i] = int(binary.LittleEndian.Uint16(payload[i*2 : i*2+2]))
	}
	offset := 10
	values := make([]string, 5)
	for i, length := range lens {
		if length < 0 || offset+length > len(payload) {
			return
		}
		values[i] = UTF16String(payload[offset : offset+length])
		offset += length
	}
	md.Title = values[0]
	md.Author = values[1]
	md.Copyright = values[2]
	md.Description = values[3]
	md.Rating = values[4]
}

func parseExtendedContent(payload []byte, md *Metadata) {
	if len(payload) < 2 {
		return
	}
	count := int(binary.LittleEndian.Uint16(payload[0:2]))
	offset := 2
	for i := 0; i < count; i++ {
		if offset+2 > len(payload) {
			return
		}
		nameLen := int(binary.LittleEndian.Uint16(payload[offset : offset+2]))
		offset += 2
		if offset+nameLen+4 > len(payload) {
			return
		}
		name := UTF16String(payload[offset : offset+nameLen])
		offset += nameLen
		valueType := binary.LittleEndian.Uint16(payload[offset : offset+2])
		valueLen := int(binary.LittleEndian.Uint16(payload[offset+2 : offset+4]))
		offset += 4
		if offset+valueLen > len(payload) {
			return
		}
		value := payload[offset : offset+valueLen]
		offset += valueLen
		md.Descriptors = append(md.Descriptors, Descriptor{Name: name, Type: valueType, Value: value})
	}
}

// Picture is a decoded WM/Picture value.
type Picture struct {
	Type byte
	MIME string
	Data []byte
}

// ParsePicture decodes the value of a WM/Picture descriptor.
func ParsePicture(value []byte) (Picture, bool) {
	if len(value) < 5 {
		return Picture{}, false
	}
	pictureType := value[0]
	dataLen := int(binary.LittleEndian.Uint32(value[1:5]))
	rest := value[5:]

	mime, rest, ok := readUTF16Z(rest)
	if !ok {
		return Picture{}, false
	}
	_, rest, ok = readUTF16Z(rest) // description
	if !ok {
		return Picture{}, false
	}
	if dataLen <= 0 || dataLen > len(rest) {
		return Picture{}, false
	}
	return Picture{Type: pictureType, MIME: mime, Data: rest[:dataLen]}, true
}

func readUTF16Z(data []byte) (string, []byte, bool) {
	for i := 0; i+1 < len(data); i += 2 {
		if data[i] == 0 && data[i+1] == 0 {
			return UTF16String(data[:i]), data[i+2:], true
		}
	}
	return "", nil, false
}

// UTF16String decodes a UTF-16LE string.
func UTF16String(data []byte) string {
	if len(data) < 2 {
		return ""
	}
	if len(data)%2 != 0 {
		data = data[:len(data)-1]
	}
	units := make([]uint16, len(data)/2)
	for i := range units {
		units[i] = binary.LittleEndian.Uint16(data[i*2:])
	}
	return strings.TrimRight(string(utf16.Decode(units)), "\x00")
}
