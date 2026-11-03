package ape

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestSerializeEmpty(t *testing.T) {
	data, err := NewTag().Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 71 { // TAGFILE(7) + header(32) + footer(32)
		t.Errorf("size = %d, want 71", len(data))
	}
	if !bytes.HasPrefix(data, []byte("TAGFILE")) {
		t.Error("missing TAGFILE prefix")
	}
}

func TestTagSizeField(t *testing.T) {
	tag := NewTag()
	tag.AddText("Title", "Test")
	data, _ := tag.Serialize()

	got := binary.LittleEndian.Uint32(data[7+12 : 7+16])
	if got != 50 {
		t.Errorf("tagSize = %d, want 50", got)
	}
}

func TestCoverArtItem(t *testing.T) {
	tag := NewTag()
	img := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	tag.AddCoverArt("Cover Art (Front)", "cover.jpg", img)
	data, _ := tag.Serialize()

	if !bytes.Contains(data, append([]byte("cover.jpg\x00"), img...)) {
		t.Error("cover art value malformed")
	}
}
