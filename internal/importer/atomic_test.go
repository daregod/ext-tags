package importer_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/format"
	"github.com/daregod/ext-tags/internal/importer"
	"github.com/daregod/ext-tags/internal/metadata"
)

type atomicWriter struct {
	path string
	err  error
}

func (w *atomicWriter) Name() string         { return "fake" }
func (w *atomicWriter) Extensions() []string { return []string{".wav"} }
func (w *atomicWriter) Write(path string, _ *metadata.Document, _ format.WriteOptions) error {
	w.path = path
	if w.err != nil {
		return w.err
	}
	return os.WriteFile(path, []byte("mutated"), 0644)
}

type atomicReader struct {
	err error
	doc *metadata.Document
}

func (r *atomicReader) Name() string         { return "fake" }
func (r *atomicReader) Extensions() []string { return []string{".wav"} }
func (r *atomicReader) Read(string, format.ReadOptions) (*metadata.Document, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.doc != nil {
		return r.doc, nil
	}
	return metadata.New(), nil
}

func TestWriteAtomicVerifyFailureKeepsOriginal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.wav")
	if err := os.WriteFile(path, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	writer := &atomicWriter{}
	reader := &atomicReader{err: errors.New("verify failed")}
	doc := metadata.New()
	doc.Fields["Title"] = "x"
	if err := importer.WriteAtomic(path, writer, reader, doc, format.WriteOptions{}); err == nil {
		t.Fatal("expected verify error")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "original" {
		t.Fatalf("original changed: %q", data)
	}
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		if entry.Name() != "track.wav" {
			t.Fatalf("temporary file left behind: %s", entry.Name())
		}
	}
}

func TestWriteAtomicSuccessReplacesOriginal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.wav")
	if err := os.WriteFile(path, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	writer := &atomicWriter{}
	doc := metadata.New()
	doc.Fields["Title"] = "x"
	reader := &atomicReader{doc: doc}
	if err := importer.WriteAtomic(path, writer, reader, doc, format.WriteOptions{}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "mutated" {
		t.Fatalf("data = %q", data)
	}
}
