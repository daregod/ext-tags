package matroska

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadFixtures(t *testing.T) {
	cases := []struct {
		name  string
		cover string
	}{
		{"audio.mkv", "cover.jpg"},
		{"audio-cover.png.mkv", "cover.png"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := Read(filepath.Join("..", "..", "..", "testdata", "formats", tc.name), nil)
			if err != nil {
				t.Fatal(err)
			}
			if doc.Fields["TITLE"] == "" {
				t.Fatalf("fields = %v", doc.Fields)
			}
			if doc.Cover == nil || doc.Cover.Filename != tc.cover {
				t.Fatalf("cover = %+v", doc.Cover)
			}
		})
	}
}

func TestReadNoAttachments(t *testing.T) {
	doc, err := Read(filepath.Join("..", "..", "..", "testdata", "formats", "audio-track.mkv"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Fields) == 0 {
		t.Fatalf("fields = %v", doc.Fields)
	}
}

func TestPickCover(t *testing.T) {
	atts := []Attachment{
		{Filename: "random.png", MIME: "image/png", Data: []byte{1}},
		{Filename: "back.jpg", MIME: "image/jpeg", Data: []byte{2}},
		{Filename: "poster.jpg", MIME: "image/jpeg", Data: []byte{3}},
	}
	if got := pickCover(atts, []string{"poster"}); got == nil || got.Filename != "poster.jpg" {
		t.Fatalf("poster = %+v", got)
	}
	if got := pickCover(atts, []string{"missing"}); got == nil || got.Filename != "random.png" {
		t.Fatalf("fallback = %+v", got)
	}
	if got := pickCover(nil, nil); got != nil {
		t.Fatalf("empty = %+v", got)
	}
}

func TestReadBroken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.mkv")
	if err := os.WriteFile(path, []byte("not matroska"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(path, nil); err == nil {
		t.Fatal("expected error")
	}
}
