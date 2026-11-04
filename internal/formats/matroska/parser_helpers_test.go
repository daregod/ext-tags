package matroska

import (
	"bytes"
	"testing"
)

func TestParserReadHelpers(t *testing.T) {
	data := []byte{0x18, 0x53, 0x80, 0x67, 0xFF, 'a', 0, 0}
	p := &parser{r: bytes.NewReader(data), size: int64(len(data)), meta: &Metadata{}}
	if got := p.readString(5, 3); got != "a" {
		t.Fatalf("readString = %q", got)
	}
	if p.readString(0, 2<<20) != "" {
		t.Fatal("oversized string must be empty")
	}
	if p.readBytes(0, 65<<20) != nil {
		t.Fatal("oversized bytes must be nil")
	}
	id, size, n, err := p.readHeader(0)
	if err != nil || id != 0x18538067 || size != 127 || n != 5 {
		t.Fatalf("header = %#x %d %d %v", id, size, n, err)
	}
	if _, _, _, err := p.readHeader(0); err != nil {
		t.Fatal(err)
	}
}

func TestParserHeaderErrors(t *testing.T) {
	for _, data := range [][]byte{{0x00}, {0x80}, {0x18, 0x53, 0x80, 0x67}} {
		p := &parser{r: bytes.NewReader(data), size: int64(len(data)), meta: &Metadata{}}
		if _, _, _, err := p.readHeader(0); err == nil {
			t.Fatalf("data % X must fail", data)
		}
	}
}
