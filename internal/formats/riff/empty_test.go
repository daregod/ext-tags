package riff

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/read"
)

func TestRF64AndWave64NoTags(t *testing.T) {
	dir := t.TempDir()
	rf64 := filepath.Join(dir, "empty.rf64")
	data := append([]byte("RF64"), []byte{0xFF, 0xFF, 0xFF, 0xFF}...)
	data = append(data, []byte("WAVE")...)
	if err := os.WriteFile(rf64, data, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadRF64(rf64); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(rf64); !errors.Is(err, read.ErrNoTags) {
		t.Fatalf("rf64 read = %v", err)
	}

	w64 := filepath.Join(dir, "empty.w64")
	head := make([]byte, 40)
	copy(head[0:4], "riff")
	copy(head[24:28], "wave")
	if err := os.WriteFile(w64, head, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadWave64(w64); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(w64); !errors.Is(err, read.ErrNoTags) {
		t.Fatalf("w64 read = %v", err)
	}
}
