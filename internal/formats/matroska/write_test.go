package matroska_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/matroska"
)

func TestWrite(t *testing.T) {
	src := filepath.Join("..", "..", "..", "testdata", "formats", "audio.mkv")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "test.mkv")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	covers := map[string]*matroska.Cover{
		"front":  {Filename: "cover.jpg", MIME: "image/jpeg", Data: []byte{1}},
		"back":   {Filename: "back.png", MIME: "image/png", Data: []byte{2}},
		"artist": {Filename: "artist.jpg", MIME: "image/jpeg", Data: []byte{3}},
		"disc":   {Filename: "disc.jpg", MIME: "image/jpeg", Data: []byte{4}},
		"icon":   {Filename: "icon.png", MIME: "image/png", Data: []byte{5}},
	}
	fields := map[string]string{
		"Title": "MKV Title", "Artist": "MKV Artist", "Track": "3/12",
	}
	if err := matroska.Write(path, fields, covers); err != nil {
		t.Fatal(err)
	}
	md, err := matroska.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var hasTitle, hasPart bool
	for _, tag := range md.Tags {
		if tag.Name == "TITLE" && tag.Value == "MKV Title" {
			hasTitle = true
		}
		if tag.Name == "PART_NUMBER" && tag.Value == "3" {
			hasPart = true
		}
	}
	if !hasTitle || !hasPart {
		t.Fatalf("tags = %+v", md.Tags)
	}
	if len(md.Attachments) != 5 {
		t.Fatalf("attachments = %d", len(md.Attachments))
	}
	if md.Attachments[0].Filename != "cover.jpg" {
		t.Fatalf("attachment = %+v", md.Attachments[0])
	}
}
