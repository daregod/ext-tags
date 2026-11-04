package riff

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/read"
)

func TestBuildDocumentNoID3(t *testing.T) {
	doc, err := buildDocument(&Metadata{Info: map[string]string{"INAM": "Title", "IART": "Artist"}}, "WAVE")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Fields["TITLE"] != "Title" || doc.Fields["ARTIST"] != "Artist" {
		t.Fatalf("fields = %v", doc.Fields)
	}
	if _, err := buildDocument(&Metadata{Info: map[string]string{}}, "WAVE"); !errors.Is(err, read.ErrNoTags) {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildAIFFInfo(t *testing.T) {
	doc, err := buildAIFF(&Metadata{Info: map[string]string{"NAME": "Title", "AUTH": "Artist", "ANNO": "Comment"}})
	if err != nil {
		t.Fatal(err)
	}
	if doc.Fields["TITLE"] != "Title" || doc.Fields["ARTIST"] != "Artist" || doc.Fields["COMMENT"] != "Comment" {
		t.Fatalf("fields = %v", doc.Fields)
	}
	if _, err := buildAIFF(&Metadata{Info: map[string]string{}}); !errors.Is(err, read.ErrNoTags) {
		t.Fatalf("err = %v", err)
	}
}

func TestReadEmptyAndBroken(t *testing.T) {
	dir := t.TempDir()
	empty := filepath.Join(dir, "empty.wav")
	if err := os.WriteFile(empty, append([]byte("RIFF\x04\x00\x00\x00WAVE"), chunk("fmt ", make([]byte, 16))...), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(empty); !errors.Is(err, read.ErrNoTags) {
		t.Fatalf("empty err = %v", err)
	}
	broken := filepath.Join(dir, "broken.wav")
	if err := os.WriteFile(broken, []byte("not riff"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(broken); !errors.Is(err, read.ErrNoTags) {
		t.Fatalf("broken err = %v", err)
	}
}

func TestRF64DataSize(t *testing.T) {
	body := chunk("ds64", make([]byte, 28))
	data := append([]byte("RF64"), []byte{0xFF, 0xFF, 0xFF, 0xFF}...)
	data = append(data, []byte("WAVE")...)
	data = append(data, body...)
	if got := rf64DataSize(data); got != 0 {
		t.Fatalf("dataSize = %d", got)
	}
}
