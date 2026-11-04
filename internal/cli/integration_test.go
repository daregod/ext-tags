package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/registry"
	"github.com/daregod/ext-tags/internal/tagformat"
)

func TestRunExtractAndImport(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), path)

	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	if err := a.runExtract(a.extractCmd(), []string{path}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".tag"); err != nil {
		t.Fatalf("tag missing: %v", err)
	}
	if err := a.runImport([]string{path}, &importOptions{replace: true, backup: true, version: 3}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Fatalf("backup missing: %v", err)
	}
	reader, _ := registry.NewDefault().ReaderFor(path)
	doc, err := reader.Read(path, tagformat.ReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if fieldValue(doc, "Title") == "" {
		t.Fatalf("title missing: %v", doc.Fields)
	}
}

func TestRunImportRejectsVersion(t *testing.T) {
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	if err := a.runImport([]string{"x.flac"}, &importOptions{version: 5}); err == nil {
		t.Fatal("expected version error")
	}
}

func TestExecuteVersion(t *testing.T) {
	old := os.Args
	defer func() { os.Args = old }()
	os.Args = []string{"ext-tags", "version"}
	if code := Execute(); code != 0 {
		t.Fatalf("exit = %d", code)
	}
}

func TestReportStatuses(t *testing.T) {
	out := captureStdout(t, func() {
		a := &app{opts: &options{}}
		for _, r := range []fileResult{
			{Status: "ok", File: "a", TagFile: "a.tag", Items: 1},
			{Status: "imported", File: "b", TagFile: "b.tag", Items: 1, Backup: "b.bak", Warning: "warn"},
			{Status: "skipped", File: "c", Reason: "exists"},
			{Status: "no-tags", File: "d"},
			{Status: "unsupported", File: "e", Reason: "fmt"},
			{Status: "error", File: "f", Error: "boom"},
			{Status: "rename-dry", File: "g", Target: "g.flac"},
			{Status: "renamed", File: "h", Target: "h.opus"},
			{Status: "import-dry", File: "i", TagFile: "i.tag", Items: 1},
		} {
			a.report(r)
		}
	})
	for _, want := range []string{"OK", "IMPORT", "SKIP", "DRY-RUN", "RENAME"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in report output: %q", want, out)
		}
	}
}

func TestReportJSON(t *testing.T) {
	out := captureStdout(t, func() {
		a := &app{opts: &options{json: true}}
		a.report(fileResult{Status: "ok", File: "a", TagFile: "a.tag"})
	})
	if !strings.Contains(out, `"status":"ok"`) {
		t.Fatalf("json = %q", out)
	}
}

func copyFixture(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func fieldValue(doc *metadata.Document, name string) string {
	for key, value := range doc.Fields {
		if strings.EqualFold(key, name) {
			return value
		}
	}
	return ""
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()
	fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
