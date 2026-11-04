package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLayers(t *testing.T) {
	if !(Layers{}).Empty() {
		t.Fatal("zero layers must be empty")
	}
	l := Layers{Explicit: "a", User: "b", Project: "c"}
	if l.Empty() {
		t.Fatal("layers must not be empty")
	}
	if got := l.Paths(); len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Fatalf("paths = %v", got)
	}
	if got := l.Label(); got != "a + b + c" {
		t.Fatalf("label = %q", got)
	}
	if got := ResolveLayers("explicit.yaml"); got.Explicit != "explicit.yaml" {
		t.Fatalf("explicit = %+v", got)
	}
}

func TestUserAndProjectPaths(t *testing.T) {
	dir := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", dir)
	} else {
		t.Setenv("XDG_CONFIG_HOME", dir)
	}
	userDir := filepath.Join(dir, "ext-tags")
	if err := os.MkdirAll(userDir, 0755); err != nil {
		t.Fatal(err)
	}
	userFile := filepath.Join(userDir, "config.yaml")
	if err := os.WriteFile(userFile, []byte("version: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if got, ok := UserPath(); !ok || got != userFile {
		t.Fatalf("UserPath = %q/%v", got, ok)
	}

	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(old) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	project := filepath.Join(dir, "ext-tags.yaml")
	if err := os.WriteFile(project, []byte("version: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if got, ok := ProjectPath(); !ok || got != project {
		t.Fatalf("ProjectPath = %q/%v", got, ok)
	}
}
