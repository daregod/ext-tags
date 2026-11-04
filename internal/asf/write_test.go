package asf_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/asf"
	"github.com/daregod/ext-tags/internal/extractor"
)

func TestWrite(t *testing.T) {
	src := filepath.Join("..", "..", "testdata", "formats", "audio.wma")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "test.wma")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	fields := map[string]string{
		"Title": "WMA Title", "Artist": "WMA Artist", "Album": "WMA Album", "Track": "3",
	}
	covers := map[string]*asf.Cover{
		"front":  {Filename: "front.jpg", MIME: "image/jpeg", Data: []byte{1}},
		"back":   {Filename: "back.png", MIME: "image/png", Data: []byte{2}},
		"artist": {Filename: "artist.jpg", MIME: "image/jpeg", Data: []byte{3}},
		"disc":   {Filename: "disc.jpg", MIME: "image/jpeg", Data: []byte{4}},
		"icon":   {Filename: "icon.png", MIME: "image/png", Data: []byte{5}},
	}
	if err := asf.Write(path, fields, covers, true); err != nil {
		t.Fatal(err)
	}
	md, err := asf.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	pictures := 0
	for _, d := range md.Descriptors {
		if d.Name == "WM/Picture" {
			pictures++
		}
	}
	if pictures != 5 {
		t.Fatalf("pictures = %d", pictures)
	}
	res, err := extractor.Extract(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(field(res.Fields, "title"), "WMA Title") {
		t.Fatalf("fields = %v", res.Fields)
	}
	if res.Cover == nil {
		t.Fatal("missing cover")
	}
}
func field(fields map[string]string, key string) string {
	for k, v := range fields {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}
