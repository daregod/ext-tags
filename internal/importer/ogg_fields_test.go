package importer_test

import (
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/ogg"
	"github.com/daregod/ext-tags/internal/importer"
)

func TestImportOggFieldsAndArtwork(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.spx")
	copyFile(t, filepath.Join("..", "..", "testdata", "formats", "audio.spx"), path)
	tf := &importer.TagFile{
		Fields: map[string]string{
			"Title": "T", "Track": "3/12", "Disc": "1/2", "Custom Field": "x",
		},
		Covers: map[string]*importer.Cover{
			"front": {Filename: "cover.jpg", MIME: "image/jpeg", Data: []byte{1}},
			"back":  {Filename: "back.jpg", MIME: "image/jpeg", Data: []byte{2}},
			"bad":   {Filename: "bad.jpg", MIME: "image/jpeg", Data: []byte{3}},
		},
	}
	if err := importer.ImportOgg(path, tf, importer.Options{Replace: true}); err != nil {
		t.Fatal(err)
	}
	comments, _, err := ogg.ReadTags(path)
	if err != nil {
		t.Fatal(err)
	}
	values := map[string]string{}
	for _, c := range comments {
		values[c.Name] = c.Value
	}
	if values["TRACKNUMBER"] != "3" || values["TRACKTOTAL"] != "12" || values["DISCTOTAL"] != "2" {
		t.Fatalf("comments = %v", values)
	}
}
