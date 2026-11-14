package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/registry"
	"github.com/daregod/ext-tags/internal/tagfile"
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

func TestExtractOutputAndSourceTag(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), audio)
	out := filepath.Join(t.TempDir(), "out.tag")

	tag := ape.NewTag()
	tag.AddText("Rating", "5")
	tag.AddText("Www", "https://example.test")
	raw, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(audio+".tag", raw, 0644); err != nil {
		t.Fatal(err)
	}

	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	cmd := a.extractCmd()
	a.opts.output = out
	if err := a.runExtract(cmd, []string{audio}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("output missing: %v", err)
	}
	parsed, err := ape.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if item, ok := parsed.Get("Rating"); !ok || string(item.Value) != "5" {
		t.Fatalf("source .tag Rating not merged: %v %v", item.Value, ok)
	}
	if item, ok := parsed.Get("Www"); !ok || string(item.Value) != "https://example.test" {
		t.Fatalf("source .tag Www not merged: %v %v", item.Value, ok)
	}
}

func TestExtractOutputConflict(t *testing.T) {
	a := &app{opts: &options{output: "x", stdout: true}}
	if err := a.validateExtractOutput(); err == nil {
		t.Fatal("output+stdout must conflict")
	}
	a = &app{opts: &options{stdout: true, json: true}}
	if err := a.validateExtractOutput(); err == nil {
		t.Fatal("stdout+json must conflict")
	}
}

func TestExtractStdout(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), audio)
	a := &app{opts: &options{}, registry: registry.NewDefault()}
	cmd := a.extractCmd()
	a.opts.stdout = true
	out := captureStdout(t, func() {
		if err := a.runExtract(cmd, []string{audio}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.HasPrefix(out, "TAGFILE") {
		t.Fatalf("stdout does not start with TAGFILE: %q", out[:minInt(16, len(out))])
	}
	if _, err := os.Stat(audio + ".tag"); !os.IsNotExist(err) {
		t.Fatal("stdout mode must not create a .tag")
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestExtractAlbumSeedsEmptyTags(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.wav")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio-empty.wav"), audio)
	if err := os.WriteFile(filepath.Join(dir, "album.yaml"), []byte("album:\n  Album: Seeded\n  Artist: Album Artist\n  Title: Album Title\n"), 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	if err := a.runExtract(a.extractCmd(), []string{audio}); err != nil {
		t.Fatal(err)
	}
	if a.failed {
		t.Fatal("extract failed")
	}
	tag, err := tagfile.ReadExisting(audio, ".tag")
	if err != nil {
		t.Fatal(err)
	}
	if tag == nil {
		t.Fatal("tag was not created")
	}
	if item, ok := tag.Get("Album"); !ok || string(item.Value) != "Seeded" {
		t.Fatalf("Album = %q, %v", item.Value, ok)
	}
}

func TestExtractDiffDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), audio)
	tag := ape.NewTag()
	tag.AddText("Title", "Old title")
	old, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	tagPath := audio + ".tag"
	if err := os.WriteFile(tagPath, old, 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{}, registry: registry.NewDefault()}
	cmd := a.extractCmd()
	a.opts.diff = true
	out := captureStdout(t, func() {
		if err := a.runExtract(cmd, []string{audio}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "DIFF") || !strings.Contains(out, "Old title") || !strings.Contains(out, "Fixture Title") {
		t.Fatalf("diff = %q", out)
	}
	if !strings.Contains(out, "cover front") || !strings.Contains(out, "cover.jpg") {
		t.Fatalf("cover diff = %q", out)
	}
	after, err := os.ReadFile(tagPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, old) {
		t.Fatal("diff modified the existing .tag")
	}
}

func TestExtractSummary(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), audio)
	a := &app{opts: &options{json: true, summary: true}, registry: registry.NewDefault()}
	cmd := a.extractCmd()
	out := captureStdout(t, func() {
		if err := a.runExtract(cmd, []string{audio}); err != nil {
			t.Fatal(err)
		}
	})
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		t.Fatalf("output = %q", out)
	}
	var summary map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &summary); err != nil {
		t.Fatal(err)
	}
	if summary["type"] != "summary" || summary["processed"].(float64) != 1 || summary["ok"].(float64) != 1 {
		t.Fatalf("summary = %v", summary)
	}
}

func TestExtractQuietSummaryOnly(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), audio)
	a := &app{opts: &options{quiet: true, json: true, summary: true}, registry: registry.NewDefault()}
	cmd := a.extractCmd()
	out := captureStdout(t, func() {
		if err := a.runExtract(cmd, []string{audio}); err != nil {
			t.Fatal(err)
		}
	})
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 || !strings.Contains(lines[0], `"type":"summary"`) {
		t.Fatalf("output = %q", out)
	}
}
