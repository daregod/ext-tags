package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFixExt(t *testing.T) {
	dir := t.TempDir()
	opus, err := os.ReadFile(filepath.Join("..", "..", "testdata", "formats", "audio.opus"))
	if err != nil {
		t.Fatal(err)
	}
	flac, err := os.ReadFile(filepath.Join("..", "..", "testdata", "formats", "audio.flac"))
	if err != nil {
		t.Fatal(err)
	}
	broken := filepath.Join(dir, "broken.flac")
	if err := os.WriteFile(broken, opus, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(broken+".tag", []byte("tag"), 0644); err != nil {
		t.Fatal(err)
	}
	native := filepath.Join(dir, "native.flac")
	if err := os.WriteFile(native, flac, 0644); err != nil {
		t.Fatal(err)
	}

	a := &app{opts: &options{}}
	if err := a.runFixExt([]string{dir}, &fixOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(broken); err != nil {
		t.Fatal("dry run must not rename")
	}

	if err := a.runFixExt([]string{dir}, &fixOptions{apply: true}); err != nil {
		t.Fatal(err)
	}
	renamed := filepath.Join(dir, "broken.opus")
	if _, err := os.Stat(renamed); err != nil {
		t.Fatalf("renamed file missing: %v", err)
	}
	if _, err := os.Stat(renamed + ".tag"); err != nil {
		t.Fatalf("renamed tag missing: %v", err)
	}
	if _, err := os.Stat(broken); !os.IsNotExist(err) {
		t.Fatal("old file still exists")
	}
	if _, err := os.Stat(native); err != nil {
		t.Fatal("native FLAC must not be renamed")
	}
}

func TestFixExtCollision(t *testing.T) {
	dir := t.TempDir()
	opus, err := os.ReadFile(filepath.Join("..", "..", "testdata", "formats", "audio.opus"))
	if err != nil {
		t.Fatal(err)
	}
	broken := filepath.Join(dir, "broken.flac")
	if err := os.WriteFile(broken, opus, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "broken.opus"), []byte("occupied"), 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{}}
	if err := a.runFixExt([]string{broken}, &fixOptions{apply: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(broken); err != nil {
		t.Fatal("file must not be renamed on collision")
	}
}
