package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/registry"
)

func TestRunImportSkipsMissingTag(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), path)
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	if err := a.runImport([]string{path}, &importOptions{version: 3}); err != nil {
		t.Fatal(err)
	}
	if a.failed {
		t.Fatal("missing .tag must skip")
	}
}

func TestRunExtractUnsupported(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.txt")
	if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	if err := a.runExtract(a.extractCmd(), []string{path}); err != nil {
		t.Fatal(err)
	}
}
func TestRunExtractValidateError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.yaml")
	if err := os.WriteFile(path, []byte("version: 2\n"), 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{quiet: true, configPath: path}, registry: registry.NewDefault()}
	if err := a.runExtract(a.extractCmd(), []string{"x.flac"}); err == nil {
		t.Fatal("invalid config must fail")
	}
}
