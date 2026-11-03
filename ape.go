package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	apeSignature = "APETAGEX"
	apeVersion   = 2000
	apeTAGFILE   = "TAGFILE"

	flagText   = uint32(0)
	flagBinary = uint32(2)

	// APEv2 tag-level flags
	tagFlagHasHeader = uint32(0x80000000) // bit 31
	tagFlagIsHeader  = uint32(0x20000000) // bit 29
)

type APEItem struct {
	Flags uint32
	Key   string
	Value []byte
}

type APETag struct {
	Items []APEItem
}

func NewAPETag() *APETag { return &APETag{} }

func (t *APETag) AddText(key, value string) {
	if key == "" || value == "" {
		return
	}
	t.Items = append(t.Items, APEItem{Flags: flagText, Key: key, Value: []byte(value)})
}

func (t *APETag) AddBinary(key string, value []byte) {
	if key == "" || len(value) == 0 {
		return
	}
	t.Items = append(t.Items, APEItem{Flags: flagBinary, Key: key, Value: value})
}

func (t *APETag) AddCoverArt(key, filename string, data []byte) {
	if key == "" || len(data) == 0 || filename == "" {
		return
	}
	v := append([]byte(filename), 0)
	v = append(v, data...)
	t.AddBinary(key, v)
}

func (t *APETag) Serialize() ([]byte, error) {
	var itemsBuf bytes.Buffer

	for _, item := range t.Items {
		if len(item.Key) > 255 {
			return nil, fmt.Errorf("key too long: %s", item.Key)
		}
		if err := binary.Write(&itemsBuf, binary.LittleEndian, uint32(len(item.Value))); err != nil {
			return nil, err
		}
		if err := binary.Write(&itemsBuf, binary.LittleEndian, item.Flags); err != nil {
			return nil, err
		}
		itemsBuf.WriteString(item.Key)
		itemsBuf.WriteByte(0)
		itemsBuf.Write(item.Value)
	}

	itemsData := itemsBuf.Bytes()
	itemCount := uint32(len(t.Items))
	tagSize := uint32(len(itemsData)) + 32 // items + footer

	header := buildAPEHeader(tagSize, itemCount, true)
	footer := buildAPEHeader(tagSize, itemCount, false)

	var out bytes.Buffer
	out.WriteString(apeTAGFILE)
	out.Write(header)
	out.Write(itemsData)
	out.Write(footer)
	return out.Bytes(), nil
}

func buildAPEHeader(tagSize, itemCount uint32, isHeader bool) []byte {
	buf := make([]byte, 32)
	copy(buf[0:8], apeSignature)
	binary.LittleEndian.PutUint32(buf[8:12], apeVersion)
	binary.LittleEndian.PutUint32(buf[12:16], tagSize)
	binary.LittleEndian.PutUint32(buf[16:20], itemCount)

	flags := tagFlagHasHeader
	if isHeader {
		flags |= tagFlagIsHeader
	}
	binary.LittleEndian.PutUint32(buf[20:24], flags)
	// bytes 24..31 reserved = zeros
	return buf
}
