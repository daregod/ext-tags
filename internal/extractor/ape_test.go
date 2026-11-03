package extractor

import (
	"os"
	"path/filepath"
	"testing"

	"external-tags/internal/ape"
)

func TestExtractAPE(t *testing.T) {
	tag := ape.NewTag()
	for key, value := range map[string]string{
		"Title": "Jedi Knight", "Artist": "Ruff Engine", "Album": "LP",
		"Year": "2003", "Genre": "Drum'n'Bass", "Track": "11", "Rating": "3",
	} {
		tag.AddText(key, value)
	}
	tag.AddCoverArt("Cover Art (Front)", "hello.jpg", []byte{0xFF, 0xD8})
	data, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "sample.ape")
	file := append([]byte("MAC "), make([]byte, 256)...)
	file = append(file, data...)
	if err := os.WriteFile(path, file, 0644); err != nil {
		t.Fatal(err)
	}

	res, err := Extract(path)
	if err != nil {
		t.Fatal(err)
	}
	if res.Format != "APE" {
		t.Fatalf("format = %q", res.Format)
	}
	if res.Fields["Title"] != "Jedi Knight" || res.Fields["Year"] != "2003" {
		t.Fatalf("fields = %v", res.Fields)
	}
	if res.Vars["tracknumber"] != "11" {
		t.Fatalf("vars = %v", res.Vars)
	}
	if res.Cover == nil || len(res.Cover.Data) != 2 {
		t.Fatalf("cover = %+v", res.Cover)
	}
}
