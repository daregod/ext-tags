package matroska

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteBroken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.mkv")
	if err := os.WriteFile(path, []byte("not matroska"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Write(path, map[string]string{"Title": "T"}, nil); err == nil {
		t.Fatal("broken write must fail")
	}
}
