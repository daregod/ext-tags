package ogg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/vorbis"
)

func TestOggFLACWrite(t *testing.T) {
	src := filepath.Join("..", "..", "..", "testdata", "formats", "audio-flac.oga")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "test.oga")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	var pictures []vorbis.Picture
	for _, typ := range []uint32{3, 4, 8, 6, 1} {
		pictures = append(pictures, vorbis.Picture{Type: typ, MIME: "image/jpeg", Data: []byte{byte(typ)}})
	}
	if err := WriteComments(path, []vorbis.Comment{{Name: "TITLE", Value: "Ogg FLAC"}}, pictures, true); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	pages, err := parsePages(out)
	if err != nil {
		t.Fatal(err)
	}
	packets, err := collectPackets(pages)
	if err != nil {
		t.Fatal(err)
	}
	count := int(packets[0].data[7])<<8 | int(packets[0].data[8])
	if count != 6 {
		t.Fatalf("header count = %d, want 6", count)
	}
	seen := 0
	for i := 1; i <= count; i++ {
		if packets[i].data[0]&0x7F == 6 {
			seen++
		}
	}
	if seen != 5 {
		t.Fatalf("picture packets = %d", seen)
	}
	if len(pages[0].payload) != len(packets[0].data) {
		t.Fatalf("first page must contain only the first packet")
	}
}
