package ape

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// ReadFromFileEnd reads an APEv2 tag stored at the end of a file
// (as used by Monkey's Audio and other formats with appended tags).
func ReadFromFileEnd(path string) (*ParsedTag, error) {
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
	if size < 32 {
		return nil, fmt.Errorf("file is too small for an APEv2 tag")
	}

	footer := make([]byte, 32)
	if _, err := f.ReadAt(footer, size-32); err != nil {
		return nil, err
	}
	if string(footer[0:8]) != apeSignature {
		return nil, fmt.Errorf("no APEv2 tag at the end of the file")
	}
	tagSize := int64(binary.LittleEndian.Uint32(footer[12:16]))
	if tagSize < 32 || tagSize > size || tagSize > 64<<20 {
		return nil, fmt.Errorf("invalid tag size %d", tagSize)
	}

	data := make([]byte, tagSize)
	if _, err := f.ReadAt(data, size-tagSize); err != nil && err != io.EOF {
		return nil, err
	}
	return Parse(data)
}
