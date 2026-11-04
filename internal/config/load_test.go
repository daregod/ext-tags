package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadErrors(t *testing.T) {
	if _, err := Load("", "prod"); err == nil {
		t.Fatal("profile without file must fail")
	}
	if cfg, err := Load("", ""); err != nil || cfg == nil {
		t.Fatalf("default load = %v, %v", cfg, err)
	}
	if _, _, err := LoadLayers("", filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Fatal("missing explicit config must fail")
	}
	bad := filepath.Join(t.TempDir(), "bad.yaml")
	if err := os.WriteFile(bad, []byte("files: [\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(bad, ""); err == nil {
		t.Fatal("bad yaml must fail")
	}
	if _, _, err := LoadLayers("", bad); err == nil {
		t.Fatal("bad layers config must fail")
	}
}

func TestProjectPathMissing(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(old) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if _, ok := ProjectPath(); ok {
		t.Fatal("empty dir must not have project config")
	}
}
