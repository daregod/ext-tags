package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadBOMAndCRLF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bom.yaml")
	data := "\xEF\xBB\xBFversion: 1\r\nfiles:\r\n  on_existing: skip\r\n"
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path, "")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Files.OnExisting != OnExistingSkip {
		t.Fatalf("on_existing = %q", cfg.Files.OnExisting)
	}
}
