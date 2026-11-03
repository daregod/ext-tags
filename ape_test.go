package main

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestSerializeEmpty(t *testing.T) {
	data, err := NewAPETag().Serialize()
	if err != nil {
		t.Fatal(err)
	}
	// TAGFILE(7) + header(32) + footer(32) = 71
	if len(data) != 71 {
		t.Errorf("size = %d, want 71", len(data))
	}
	if !bytes.HasPrefix(data, []byte("TAGFILE")) {
		t.Error("missing TAGFILE prefix")
	}
	if !bytes.HasPrefix(data[7:], []byte("APETAGEX")) {
		t.Error("missing header signature")
	}
	if !bytes.HasPrefix(data[len(data)-32:], []byte("APETAGEX")) {
		t.Error("missing footer signature")
	}
}

func TestTagSizeField(t *testing.T) {
	tag := NewAPETag()
	tag.AddText("Title", "Test")
	data, _ := tag.Serialize()

	// itemsSize = 4(size)+4(flags)+5("Title")+1(nul)+4("Test") = 18
	// tagSize   = 18 + 32 = 50
	got := binary.LittleEndian.Uint32(data[7+12 : 7+16])
	if got != 50 {
		t.Errorf("tagSize = %d, want 50", got)
	}
}

func TestTextItem(t *testing.T) {
	tag := NewAPETag()
	tag.AddText("Title", "Hello")
	data, _ := tag.Serialize()

	if !bytes.Contains(data, []byte("Title\x00Hello")) {
		t.Error("missing Title item")
	}
}

func TestCoverArtItem(t *testing.T) {
	tag := NewAPETag()
	img := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	tag.AddCoverArt("Cover Art (Front)", "cover.jpg", img)
	data, _ := tag.Serialize()

	// Формат: <filename>\0<imageData>
	if !bytes.Contains(data, append([]byte("cover.jpg\x00"), img...)) {
		t.Error("cover art value malformed")
	}
}

func TestHeaderAndFooterFlags(t *testing.T) {
	data, _ := NewAPETag().Serialize()
	headerFlags := binary.LittleEndian.Uint32(data[7+20 : 7+24])
	footerFlags := binary.LittleEndian.Uint32(data[len(data)-12 : len(data)-8])
	if headerFlags != 0xA0000000 {
		t.Errorf("header flags = %#x, want 0xA0000000", headerFlags)
	}
	if footerFlags != 0x80000000 {
		t.Errorf("footer flags = %#x, want 0x80000000", footerFlags)
	}
}
