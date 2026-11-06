package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/registry"
	"github.com/daregod/ext-tags/internal/tagformat"
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
func TestRunExtractVerbose(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), path)
	a := &app{opts: &options{verbose: true}, registry: registry.NewDefault()}
	if err := a.runExtract(a.extractCmd(), []string{path}); err != nil {
		t.Fatal(err)
	}
}

func TestRunImportDryRun(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), path)
	writeTestTag(t, path+".tag", "Dry")
	a := &app{opts: &options{quiet: true, dryRun: true}, registry: registry.NewDefault()}
	if err := a.runImport([]string{path}, &importOptions{version: 3}); err != nil {
		t.Fatal(err)
	}
}
func TestImportFileOggInFlacWarning(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.opus"), path)
	writeTestTag(t, path+".tag", "Ogg in FLAC")
	doc := metadata.New()
	doc.Fields["Title"] = "Ogg in FLAC"
	a := &app{opts: &options{quiet: true}, registry: tagformat.New(
		[]tagformat.Reader{&serviceReader{doc: doc}},
		[]tagformat.Writer{&serviceWriter{}},
	)}
	a.importFile(path, ".tag", nil, &importOptions{version: 3})
	if a.failed {
		t.Fatal("import failed")
	}
}
func TestRunImportValidateError(t *testing.T) {
	bad := filepath.Join(t.TempDir(), "bad.yaml")
	if err := os.WriteFile(bad, []byte("version: 2\n"), 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{quiet: true, configPath: bad}, registry: registry.NewDefault()}
	if err := a.runImport([]string{"x.flac"}, &importOptions{version: 3}); err == nil {
		t.Fatal("invalid config must fail")
	}
}
func TestRunExtractWarningsAndFixExtValidate(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "warn.yaml")
	cfg := "version: 1\nfiles:\n  on_existing: skip\n  merge_existing: true\n"
	if err := os.WriteFile(cfgPath, []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	txt := filepath.Join(dir, "track.txt")
	if err := os.WriteFile(txt, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{verbose: true, configPath: cfgPath}, registry: registry.NewDefault()}
	if err := a.runExtract(a.extractCmd(), []string{txt}); err != nil {
		t.Fatal(err)
	}

	bad := filepath.Join(t.TempDir(), "bad.yaml")
	if err := os.WriteFile(bad, []byte("version: 2\n"), 0644); err != nil {
		t.Fatal(err)
	}
	a = &app{opts: &options{quiet: true, configPath: bad}}
	if err := a.runFixExt([]string{"x.flac"}, &fixOptions{}); err == nil {
		t.Fatal("invalid config must fail")
	}
}
func TestExtractCommandRun(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), path)
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	root := a.rootCmd()
	root.SetArgs([]string{"extract", path})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".tag"); err != nil {
		t.Fatalf("tag missing: %v", err)
	}
}
