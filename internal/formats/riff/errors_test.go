package riff

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/metadata"
)

func TestWriteErrors(t *testing.T) {
	dir := t.TempDir()
	doc := metadata.New()
	if err := Write(filepath.Join(dir, "missing.wav"), doc); err == nil {
		t.Fatal("missing file must fail")
	}
	bad := filepath.Join(dir, "bad.wav")
	if err := os.WriteFile(bad, []byte("not riff"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Write(bad, doc); err == nil {
		t.Fatal("bad container must fail")
	}
}

func TestReadBadContainers(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"bad.rf64", "bad.w64", "bad.aiff"} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("garbage"), 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := Read(path); err == nil {
			t.Fatalf("%s must fail", name)
		}
	}
}

func TestPatchDS64NoChunk(t *testing.T) {
	data := []byte("RIFF\x04\x00\x00\x00WAVE")
	patchDS64(data)
}
