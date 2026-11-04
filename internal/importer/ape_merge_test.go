package importer_test

import (
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/importer"
)

func TestImportAPEv2Merge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.wv")
	copyFile(t, filepath.Join("..", "..", "testdata", "formats", "audio.wv"), path)
	tf := &importer.TagFile{
		Fields: map[string]string{"Title": "New", "Custom": "x"},
		Covers: map[string]*importer.Cover{
			"front": {Filename: "cover.jpg", MIME: "image/jpeg", Data: []byte{1}},
			"back":  {Filename: "", MIME: "image/jpeg", Data: []byte{2}},
			"bad":   {Filename: "bad.jpg", MIME: "image/jpeg", Data: []byte{3}},
		},
	}
	if err := importer.ImportAPEv2(path, tf, importer.Options{Replace: false}); err != nil {
		t.Fatal(err)
	}
	doc, err := ape.Read(path, "WAVPACK")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Fields["Title"] != "New" {
		t.Fatalf("fields = %v", doc.Fields)
	}
	if doc.Artwork["front"] == nil || doc.Artwork["back"] == nil {
		t.Fatalf("artwork = %+v", doc.Artwork)
	}
}
