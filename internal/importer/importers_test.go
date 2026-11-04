package importer_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/importer"
	"github.com/daregod/ext-tags/internal/registry"
	"github.com/daregod/ext-tags/internal/tagformat"
)

func TestImportWriters(t *testing.T) {
	cases := []struct {
		fixture  string
		importFn func(string, *importer.TagFile, importer.Options) error
	}{
		{"audio-aac.m4a", importer.ImportMP4},
		{"audio.ogg", importer.ImportOgg},
		{"audio.opus", importer.ImportOgg},
		{"audio.mkv", importer.ImportMatroska},
		{"audio.wma", importer.ImportASF},
		{"audio.wav", importer.ImportRIFF},
	}
	for _, tc := range cases {
		t.Run(tc.fixture, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, tc.fixture)
			copyFile(t, filepath.Join("..", "..", "testdata", "formats", tc.fixture), path)
			tf := &importer.TagFile{Fields: map[string]string{
				"Title": "Imported Title", "Artist": "Imported Artist",
				"Album": "Imported Album", "Genre": "Imported Genre", "Year": "2026",
			}}
			if err := tc.importFn(path, tf, importer.Options{Replace: true, Backup: true}); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(path + ".bak"); err != nil {
				t.Fatalf("backup missing: %v", err)
			}
			reader, ok := registry.NewDefault().ReaderFor(path)
			if !ok {
				t.Fatal("no reader")
			}
			doc, err := reader.Read(path, tagformat.ReadOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.EqualFold(fieldValue(doc.Fields, "Title"), "Imported Title") {
				t.Fatalf("fields = %v", doc.Fields)
			}
		})
	}
}
