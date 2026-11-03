package ape

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// ParsedTag is a decoded APEv2 tag (optionally wrapped in a TAGFILE file).
type ParsedTag struct {
	Prefix    bool
	Version   uint32
	Items     []Item
	HasHeader bool
}

// Get returns the first item with the given key (case-insensitive).
func (t *ParsedTag) Get(key string) (Item, bool) {
	for _, item := range t.Items {
		if strings.EqualFold(item.Key, key) {
			return item, true
		}
	}
	return Item{}, false
}

// ReadFile parses an APEv2 tag file produced by ext-tags or foo_external_tags.
func ReadFile(path string) (*ParsedTag, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

// Parse decodes an APEv2 tag. Both bare tags (header/items/footer) and
// TAGFILE-wrapped tags (".tag" files) are supported.
func Parse(data []byte) (*ParsedTag, error) {
	if len(data) < 32 {
		return nil, fmt.Errorf("tag too short: %d bytes", len(data))
	}

	t := &ParsedTag{Prefix: bytes.HasPrefix(data, []byte(apeTAGFILE))}

	footer := data[len(data)-32:]
	if string(footer[0:8]) != apeSignature {
		return nil, fmt.Errorf("missing APETAGEX footer")
	}
	t.Version = binary.LittleEndian.Uint32(footer[8:12])
	tagSize := binary.LittleEndian.Uint32(footer[12:16])
	itemCount := binary.LittleEndian.Uint32(footer[16:20])
	headerStart := 0
	if t.Prefix {
		headerStart = len(apeTAGFILE)
	}
	t.HasHeader = len(data) >= headerStart+32 && string(data[headerStart:headerStart+8]) == apeSignature

	if tagSize < 32 || int(tagSize) > len(data) {
		return nil, fmt.Errorf("invalid tag size %d", tagSize)
	}
	itemsEnd := len(data) - 32
	offset := itemsEnd - int(tagSize) + 32

	for i := 0; i < int(itemCount); i++ {
		if offset+8 > itemsEnd {
			return nil, fmt.Errorf("item %d is truncated", i)
		}
		valueLen := binary.LittleEndian.Uint32(data[offset : offset+4])
		itemFlags := binary.LittleEndian.Uint32(data[offset+4 : offset+8])
		offset += 8

		keyStart := offset
		for offset < itemsEnd && data[offset] != 0 {
			offset++
		}
		if offset >= itemsEnd {
			return nil, fmt.Errorf("item %d key is not terminated", i)
		}
		key := string(data[keyStart:offset])
		offset++

		if int(valueLen) < 0 || offset+int(valueLen) > itemsEnd {
			return nil, fmt.Errorf("item %d value is truncated", i)
		}
		value := append([]byte(nil), data[offset:offset+int(valueLen)]...)
		offset += int(valueLen)

		t.Items = append(t.Items, Item{Flags: itemFlags, Key: key, Value: value})
	}
	if offset != itemsEnd {
		return nil, fmt.Errorf("%d trailing bytes after items", itemsEnd-offset)
	}
	return t, nil
}
