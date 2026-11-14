package importer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/matroska"
)

func copyMatroska(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "formats", "audio-cover.png.mkv"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "track.mkv")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestImportMatroskaArtworkOnly(t *testing.T) {
	path := copyMatroska(t)
	tf := &TagFile{Fields: map[string]string{"Title": "New"}, Covers: map[string]*Cover{
		"back": {Filename: "back.jpg", MIME: "image/jpeg", Data: []byte{2}},
	}}
	if err := ImportMatroska(path, tf, Options{Only: []string{"Cover Art (Back)"}}); err != nil {
		t.Fatal(err)
	}
	md, err := matroska.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(md.Attachments) != 2 {
		t.Fatalf("attachments = %+v", md.Attachments)
	}
	found := false
	for _, att := range md.Attachments {
		if att.Filename == "back.jpg" {
			found = true
		}
	}
	if !found {
		t.Fatalf("back missing: %+v", md.Attachments)
	}

	if err := ImportMatroska(path, &TagFile{Fields: map[string]string{"Title": "New"}}, Options{Only: []string{"Cover Art (Back)"}}); err != nil {
		t.Fatal(err)
	}
	md, err = matroska.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(md.Attachments) != 1 {
		t.Fatalf("back not removed: %+v", md.Attachments)
	}
}

func TestImportMatroskaFieldsKeepAttachments(t *testing.T) {
	path := copyMatroska(t)
	before, err := matroska.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	tf := &TagFile{Fields: map[string]string{"Title": "New title", "Track": "2/10"}}
	if err := ImportMatroska(path, tf, Options{Only: []string{"Title"}}); err != nil {
		t.Fatal(err)
	}
	md, err := matroska.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(md.Attachments) != len(before.Attachments) {
		t.Fatalf("attachments changed: %+v", md.Attachments)
	}
	doc, err := matroska.Read(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Fields["TITLE"] != "New title" {
		t.Fatalf("fields = %+v", doc.Fields)
	}
}
