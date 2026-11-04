package importer_test

import (
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/importer"
	"github.com/daregod/ext-tags/internal/registry"
	"github.com/daregod/ext-tags/internal/tagformat"
)

func TestImportCoverFallbacks(t *testing.T) {
	cases := []struct {
		name    string
		fixture string
		fn      func(string, *importer.TagFile, importer.Options) error
	}{
		{"asf", "audio.wma", importer.ImportASF},
		{"matroska", "audio.mkv", importer.ImportMatroska},
		{"mp4", "audio-aac.m4a", importer.ImportMP4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, tc.fixture)
			copyFile(t, filepath.Join("..", "..", "testdata", "formats", tc.fixture), path)
			tf := &importer.TagFile{
				Fields: map[string]string{"Title": "Cover Title"},
				Covers: map[string]*importer.Cover{
					"front": {Filename: "cover.jpg", MIME: "image/jpeg", Data: []byte{1}},
					"bad":   nil,
				},
			}
			if err := tc.fn(path, tf, importer.Options{Replace: true}); err != nil {
				t.Fatal(err)
			}
			reader, ok := registry.NewDefault().ReaderFor(path)
			if !ok {
				t.Fatal("no reader")
			}
			doc, err := reader.Read(path, tagformat.ReadOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if doc.Fields["Title"] != "Cover Title" && doc.Fields["TITLE"] != "Cover Title" {
				t.Fatalf("fields = %v", doc.Fields)
			}
			if doc.Cover == nil {
				t.Fatalf("cover missing: %+v", doc.Artwork)
			}
		})
	}
}

func TestImportASFFrontFallback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.wma")
	copyFile(t, filepath.Join("..", "..", "testdata", "formats", "audio.wma"), path)
	tf := &importer.TagFile{
		Fields: map[string]string{"Title": "Fallback"},
		Cover:  &importer.Cover{Filename: "front.jpg", MIME: "image/jpeg", Data: []byte{1}},
	}
	if err := importer.ImportASF(path, tf, importer.Options{Replace: true}); err != nil {
		t.Fatal(err)
	}
}
