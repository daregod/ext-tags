package mp4_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/mp4"
	"github.com/daregod/ext-tags/internal/read"
)

func TestWrite(t *testing.T) {
	src := filepath.Join("..", "..", "..", "testdata", "formats", "audio-aac.m4a")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "test.m4a")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	cover := &mp4.Cover{Filename: "cover.jpg", MIME: "image/jpeg", Data: []byte{0xFF, 0xD8}}
	fields := map[string]string{
		"Title": "MP4 Title", "Artist": "MP4 Artist", "Album": "MP4 Album",
		"Track": "3/12", "Year": "2026", "Genre": "Test",
	}
	if err := mp4.Write(path, fields, cover, true); err != nil {
		t.Fatal(err)
	}
	res, err := read.Taglib(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(fieldValue(res.Fields, "title"), "MP4 Title") || !strings.EqualFold(fieldValue(res.Fields, "artist"), "MP4 Artist") {
		t.Fatalf("fields = %v", res.Fields)
	}
	if res.Cover == nil || len(res.Cover.Data) != 2 {
		t.Fatalf("cover = %+v", res.Cover)
	}
}
func fieldValue(fields map[string]string, key string) string {
	for k, v := range fields {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}
