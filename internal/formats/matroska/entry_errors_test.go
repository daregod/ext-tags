package matroska

import (
	"path/filepath"
	"testing"
)

func TestReadFileMissing(t *testing.T) {
	if _, err := ReadFile(filepath.Join(t.TempDir(), "missing.mkv")); err == nil {
		t.Fatal("missing must fail")
	}
}

func TestWriteMissing(t *testing.T) {
	if err := Write(filepath.Join(t.TempDir(), "missing.mkv"), nil, nil); err == nil {
		t.Fatal("missing write must fail")
	}
}
