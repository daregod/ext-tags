package album

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPath(t *testing.T) {
	dir := t.TempDir()
	if cfg, err := Load(dir); err != nil || cfg != nil {
		t.Fatalf("missing = %+v %v", cfg, err)
	}
	path := filepath.Join(dir, "album.yaml")
	if err := os.WriteFile(path, []byte("album:\n  Title: X\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil || cfg == nil {
		t.Fatalf("load = %+v %v", cfg, err)
	}
	if cfg.Path() != path {
		t.Fatalf("path = %q", cfg.Path())
	}
	var nilCfg *Config
	if nilCfg.Path() != "" {
		t.Fatal("nil path")
	}
}
