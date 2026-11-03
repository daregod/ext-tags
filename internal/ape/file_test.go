package ape

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadFromFileEnd(t *testing.T) {
	tag := NewTag()
	tag.AddText("Artist", "Noisia")
	tag.AddText("Title", "Stigma")
	tag.AddCoverArt("Cover Art (Front)", "cover.jpg", []byte{0xFF, 0xD8})
	data, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "sample.ape")
	file := append([]byte("MAC "), make([]byte, 512)...)
	file = append(file, data...)
	if err := os.WriteFile(path, file, 0644); err != nil {
		t.Fatal(err)
	}

	parsed, err := ReadFromFileEnd(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.Items) != 3 {
		t.Fatalf("items = %d, want 3", len(parsed.Items))
	}
	if item, ok := parsed.Get("Artist"); !ok || string(item.Value) != "Noisia" {
		t.Fatalf("Artist = %q", item.Value)
	}
}

func TestReadFromFileEndNoTag(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plain.ape")
	if err := os.WriteFile(path, []byte("MAC "), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadFromFileEnd(path); err == nil {
		t.Fatal("expected error for a file without APEv2 tag")
	}
}
