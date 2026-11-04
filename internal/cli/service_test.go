package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/format"
	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/read"
	"github.com/daregod/ext-tags/internal/rules"
)

type serviceReader struct {
	doc     *metadata.Document
	err     error
	gotOpts format.ReadOptions
}

func (r *serviceReader) Name() string         { return "fake" }
func (r *serviceReader) Extensions() []string { return []string{".flac"} }
func (r *serviceReader) Read(_ string, opts format.ReadOptions) (*metadata.Document, error) {
	r.gotOpts = opts
	return r.doc, r.err
}

type serviceWriter struct {
	called bool
	doc    *metadata.Document
	opts   format.WriteOptions
	err    error
}

func (w *serviceWriter) Name() string         { return "fake" }
func (w *serviceWriter) Extensions() []string { return []string{".flac"} }
func (w *serviceWriter) Write(_ string, doc *metadata.Document, opts format.WriteOptions) error {
	w.called = true
	w.doc = doc
	w.opts = opts
	return w.err
}

func testConfig() *config.Config {
	cfg := config.Default()
	cfg.Files.OnExisting = config.OnExistingOverwrite
	return cfg
}

func TestProcessFileUsesInjectedReader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	if err := os.WriteFile(path, []byte("audio"), 0644); err != nil {
		t.Fatal(err)
	}
	doc := metadata.New()
	doc.Fields["Title"] = "Injected"
	reader := &serviceReader{doc: doc}
	a := &app{opts: &options{quiet: true}, registry: format.New([]format.Reader{reader}, nil)}
	cfg := testConfig()
	engine, err := rules.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	a.processFile(path, cfg, engine)

	tag, err := ape.ReadFile(path + ".tag")
	if err != nil {
		t.Fatal(err)
	}
	item, ok := tag.Get("Title")
	if !ok || string(item.Value) != "Injected" {
		t.Fatalf("Title = %q, ok=%v", item.Value, ok)
	}
	if len(reader.gotOpts.CoverNames) == 0 {
		t.Fatal("reader did not receive cover names")
	}
}

func TestProcessFileUnknownReader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	if err := os.WriteFile(path, []byte("audio"), 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{quiet: true}, registry: format.New(nil, nil)}
	cfg := testConfig()
	engine, err := rules.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	a.processFile(path, cfg, engine)
	if _, err := os.Stat(path + ".tag"); !os.IsNotExist(err) {
		t.Fatal("tag must not be written without a reader")
	}
}

func TestProcessFileNoTags(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	if err := os.WriteFile(path, []byte("audio"), 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{quiet: true}, registry: format.New([]format.Reader{&serviceReader{err: read.ErrNoTags}}, nil)}
	cfg := testConfig()
	engine, err := rules.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	a.processFile(path, cfg, engine)
	if _, err := os.Stat(path + ".tag"); !os.IsNotExist(err) {
		t.Fatal("tag must not be written when reader reports no tags")
	}
}

func TestImportFileUsesInjectedWriter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	if err := os.WriteFile(path, []byte("audio"), 0644); err != nil {
		t.Fatal(err)
	}
	writeTestTag(t, path+".tag", "Imported")
	writer := &serviceWriter{}
	doc := metadata.New()
	doc.Fields["Title"] = "Imported"
	a := &app{opts: &options{quiet: true}, registry: format.New([]format.Reader{&serviceReader{doc: doc}}, []format.Writer{writer})}
	a.importFile(path, ".tag", &importOptions{replace: true, backup: true, version: 4, v1: true})

	if !writer.called {
		t.Fatal("writer was not called")
	}
	if writer.doc.Fields["Title"] != "Imported" {
		t.Fatalf("fields = %v", writer.doc.Fields)
	}
	if !writer.opts.Replace || writer.opts.Backup || writer.opts.ID3Version != 4 || !writer.opts.ID3V1 {
		t.Fatalf("opts = %+v", writer.opts)
	}
	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Fatalf("backup missing: %v", err)
	}
}

func TestImportFileUnknownWriter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	if err := os.WriteFile(path, []byte("audio"), 0644); err != nil {
		t.Fatal(err)
	}
	writeTestTag(t, path+".tag", "Imported")
	a := &app{opts: &options{quiet: true}, registry: format.New(nil, nil)}
	a.importFile(path, ".tag", &importOptions{})
	if a.failed {
		t.Fatal("unknown writer must not mark failure")
	}
}

func writeTestTag(t *testing.T, path, title string) {
	t.Helper()
	tag := ape.NewTag()
	tag.AddText("Title", title)
	data, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}
