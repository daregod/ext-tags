// Package vorbis implements Vorbis comments and FLAC picture blocks.
package vorbis

import (
	"encoding/binary"
	"fmt"
)

// Comment is one Vorbis-comment entry (NAME=value).
type Comment struct {
	Name  string
	Value string
}

// Picture is a FLAC PICTURE metadata block.
type Picture struct {
	Type        uint32
	MIME        string
	Description string
	Width       uint32
	Height      uint32
	Depth       uint32
	Colors      uint32
	Data        []byte
}

// MarshalComments encodes a Vorbis-comment block. framing adds the Vorbis
// framing bit (used by Ogg Vorbis, not by FLAC).
func MarshalComments(vendor string, comments []Comment, framing bool) []byte {
	out := make([]byte, 4, 64)
	binary.LittleEndian.PutUint32(out, uint32(len(vendor)))
	out = append(out, vendor...)
	count := make([]byte, 4)
	binary.LittleEndian.PutUint32(count, uint32(len(comments)))
	out = append(out, count...)
	for _, c := range comments {
		line := c.Name + "=" + c.Value
		length := make([]byte, 4)
		binary.LittleEndian.PutUint32(length, uint32(len(line)))
		out = append(out, length...)
		out = append(out, line...)
	}
	if framing {
		out = append(out, 1)
	}
	return out
}

// ParseComments decodes a Vorbis-comment block.
func ParseComments(data []byte, framing bool) (string, []Comment, error) {
	if len(data) < 8 {
		return "", nil, fmt.Errorf("vorbis comments too short")
	}
	vendorLen := int(binary.LittleEndian.Uint32(data[0:4]))
	offset := 4
	if vendorLen < 0 || offset+vendorLen+4 > len(data) {
		return "", nil, fmt.Errorf("bad vendor length")
	}
	vendor := string(data[offset : offset+vendorLen])
	offset += vendorLen
	count := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	offset += 4
	comments := make([]Comment, 0, count)
	for i := 0; i < count; i++ {
		if offset+4 > len(data) {
			return "", nil, fmt.Errorf("truncated comment length")
		}
		length := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
		offset += 4
		if length < 0 || offset+length > len(data) {
			return "", nil, fmt.Errorf("truncated comment")
		}
		line := string(data[offset : offset+length])
		offset += length
		name, value, ok := Split(line)
		if ok {
			comments = append(comments, Comment{Name: name, Value: value})
		}
	}
	return vendor, comments, nil
}

// Split splits NAME=value.
func Split(line string) (string, string, bool) {
	for i := 0; i < len(line); i++ {
		if line[i] == '=' {
			return line[:i], line[i+1:], true
		}
	}
	return "", "", false
}

// MarshalPicture encodes a FLAC PICTURE block payload.
func MarshalPicture(p Picture) []byte {
	out := make([]byte, 0, 32+len(p.Data))
	out = appendU32(out, p.Type)
	out = appendString(out, p.MIME)
	out = appendString(out, p.Description)
	out = appendU32(out, p.Width)
	out = appendU32(out, p.Height)
	out = appendU32(out, p.Depth)
	out = appendU32(out, p.Colors)
	out = appendU32(out, uint32(len(p.Data)))
	return append(out, p.Data...)
}

// ParsePicture decodes a FLAC PICTURE block payload.
func ParsePicture(data []byte) (Picture, error) {
	p := Picture{}
	offset := 0
	readU32 := func() (uint32, error) {
		if offset+4 > len(data) {
			return 0, fmt.Errorf("truncated picture block")
		}
		v := binary.BigEndian.Uint32(data[offset : offset+4])
		offset += 4
		return v, nil
	}
	readString := func() (string, error) {
		n, err := readU32()
		if err != nil {
			return "", err
		}
		if offset+int(n) > len(data) {
			return "", fmt.Errorf("truncated picture string")
		}
		s := string(data[offset : offset+int(n)])
		offset += int(n)
		return s, nil
	}
	var err error
	if p.Type, err = readU32(); err != nil {
		return p, err
	}
	if p.MIME, err = readString(); err != nil {
		return p, err
	}
	if p.Description, err = readString(); err != nil {
		return p, err
	}
	if p.Width, err = readU32(); err != nil {
		return p, err
	}
	if p.Height, err = readU32(); err != nil {
		return p, err
	}
	if p.Depth, err = readU32(); err != nil {
		return p, err
	}
	if p.Colors, err = readU32(); err != nil {
		return p, err
	}
	dataLen, err := readU32()
	if err != nil {
		return p, err
	}
	if offset+int(dataLen) > len(data) {
		return p, fmt.Errorf("truncated picture data")
	}
	p.Data = append([]byte(nil), data[offset:offset+int(dataLen)]...)
	return p, nil
}

func appendU32(dst []byte, v uint32) []byte {
	return append(dst, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

func appendString(dst []byte, s string) []byte {
	dst = appendU32(dst, uint32(len(s)))
	return append(dst, s...)
}
