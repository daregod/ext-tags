package ape_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/read"
)

func TestRead(t *testing.T) {
	tag := ape.NewTag()
	tag.AddText("Title", "APE Title")
	tag.AddText("Artist", "APE Artist")
	for _, key := range []string{"Cover Art (Front)", "Cover Art (Back)", "Artist Picture", "Cover Art (Disc)", "Cover Art (Icon)"} {
		tag.AddCoverArt(key, "cover.jpg", []byte{0xFF, 0xD8})
	}
	data, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "track.wv")
	if err := os.WriteFile(path, append([]byte("wvpk"), data...), 0644); err != nil {
		t.Fatal(err)
	}
	doc, err := ape.Read(path, "WAVPACK")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Fields["Title"] != "APE Title" || doc.Fields["Artist"] != "APE Artist" {
		t.Fatalf("fields = %v", doc.Fields)
	}
	for _, kind := range []string{"front", "back", "artist", "disc", "icon"} {
		if doc.Artwork[kind] == nil {
			t.Fatalf("missing %s artwork", kind)
		}
	}
}

func TestReadNoTag(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.wv")
	if err := os.WriteFile(path, []byte("wvpk"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := ape.Read(path, "WAVPACK"); err != read.ErrNoTags {
		t.Fatalf("err = %v", err)
	}
}

func TestSerializeEmbedded(t *testing.T) {
	tag := ape.NewTag()
	tag.AddText("Title", "Embedded")
	data, err := tag.SerializeEmbedded()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ape.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if item, ok := parsed.Get("Title"); !ok || string(item.Value) != "Embedded" {
		t.Fatalf("parsed = %+v", parsed.Items)
	}
}
func TestReadFile(t *testing.T) {
	tag := ape.NewTag()
	tag.AddText("Title", "File")
	data, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "track.tag")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	parsed, err := ape.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if item, ok := parsed.Get("Title"); !ok || string(item.Value) != "File" {
		t.Fatalf("items = %+v", parsed.Items)
	}
	if _, err := ape.ReadFile(filepath.Join(t.TempDir(), "missing.tag")); err == nil {
		t.Fatal("missing file must fail")
	}
}
