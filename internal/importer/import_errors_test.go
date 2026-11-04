package importer_test

import (
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/importer"
)

func TestImportersMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.wav")
	tf := &importer.TagFile{Fields: map[string]string{"Title": "x"}}
	cases := []struct {
		name string
		fn   func(string, *importer.TagFile, importer.Options) error
	}{
		{"mp3", importer.ImportMP3},
		{"flac", importer.ImportFLAC},
		{"ogg", importer.ImportOgg},
		{"apev2", importer.ImportAPEv2},
		{"mp4", importer.ImportMP4},
		{"matroska", importer.ImportMatroska},
		{"asf", importer.ImportASF},
		{"riff", importer.ImportRIFF},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(path, tf, importer.Options{Backup: true}); err == nil {
				t.Fatal("expected error for missing file")
			}
		})
	}
}
