package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/registry"
	"github.com/daregod/ext-tags/internal/rules"
	"github.com/daregod/ext-tags/internal/tagformat"
)

func TestConfigShowAndPath(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(old) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	a := &app{opts: &options{}}
	root := a.rootCmd()
	root.SetArgs([]string{"config", "show"})
	out := captureStdout(t, func() {
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "files:") {
		t.Fatalf("show = %q", out)
	}

	a = &app{opts: &options{}}
	root = a.rootCmd()
	root.SetArgs([]string{"config", "path"})
	out = captureStdout(t, func() {
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "built-in defaults") {
		t.Fatalf("path = %q", out)
	}

	a = &app{opts: &options{}}
	root = a.rootCmd()
	root.SetArgs([]string{"--config", "custom.yaml", "config", "path"})
	out = captureStdout(t, func() {
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "explicit: custom.yaml") {
		t.Fatalf("explicit path = %q", out)
	}
}

func TestConfigValidateInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(path, []byte("files: [\n"), 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{}}
	root := a.rootCmd()
	root.SetArgs([]string{"--config", path, "config", "validate"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected invalid config error")
	}
}

func TestExecuteUnknownCommand(t *testing.T) {
	old := os.Args
	defer func() { os.Args = old }()
	os.Args = []string{"ext-tags", "--definitely-unknown-flag"}
	if code := Execute(); code == 0 {
		t.Fatal("expected non-zero exit")
	}
}

func TestProcessFileReaderErrorAndDryRun(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	if err := os.WriteFile(path, []byte("audio"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := testConfig()
	engine, err := rules.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	a := &app{opts: &options{quiet: true}, registry: tagformat.New([]tagformat.Reader{&serviceReader{err: errors.New("boom")}}, nil)}
	a.processFile(path, cfg, engine)
	if _, err := os.Stat(path + ".tag"); !os.IsNotExist(err) {
		t.Fatal("tag must not be written on reader error")
	}
	if !a.failed {
		t.Fatal("reader error must mark failure")
	}

	a = &app{opts: &options{quiet: true, dryRun: true}, registry: registry.NewDefault()}
	a.processFile(path, cfg, engine)
	if _, err := os.Stat(path + ".tag"); !os.IsNotExist(err) {
		t.Fatal("dry-run must not write")
	}
}

func TestImportFileEdgeCases(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	if err := os.WriteFile(audio, []byte("audio"), 0644); err != nil {
		t.Fatal(err)
	}

	a := &app{opts: &options{quiet: true}}
	a.importFile(audio, ".tag", &importOptions{})
	if a.failed {
		t.Fatal("missing .tag must be a skip")
	}

	if err := os.WriteFile(audio+".tag", []byte("not a tag"), 0644); err != nil {
		t.Fatal(err)
	}
	a = &app{opts: &options{quiet: true}}
	a.importFile(audio, ".tag", &importOptions{})
	if !a.failed {
		t.Fatal("invalid .tag must fail")
	}

	writeTestTag(t, audio+".tag", "Song")
	a = &app{opts: &options{quiet: true, dryRun: true}}
	a.importFile(audio, ".tag", &importOptions{})
	if a.failed {
		t.Fatal("dry-run must not fail")
	}

	writeTestTag(t, audio+".tag", "Song")
	writer := &serviceWriter{}
	a = &app{opts: &options{quiet: true}, registry: tagformat.New(nil, []tagformat.Writer{writer})}
	a.importFile(audio, ".tag", &importOptions{})
	if a.failed {
		t.Fatal("missing reader must be unsupported, not error")
	}

	unsupported := filepath.Join(dir, "track.txt")
	if err := os.WriteFile(unsupported, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	a = &app{opts: &options{quiet: true}}
	a.importFile(unsupported, ".tag", &importOptions{})
}
