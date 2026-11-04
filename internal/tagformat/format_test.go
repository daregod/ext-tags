package tagformat

import (
	"errors"
	"testing"

	"github.com/daregod/ext-tags/internal/metadata"
)

type fakeReader struct {
	name string
	exts []string
	doc  *metadata.Document
	err  error
}

func (r *fakeReader) Name() string         { return r.name }
func (r *fakeReader) Extensions() []string { return r.exts }
func (r *fakeReader) Read(string, ReadOptions) (*metadata.Document, error) {
	return r.doc, r.err
}

type fakeWriter struct {
	name string
	exts []string
	doc  *metadata.Document
	opts WriteOptions
	err  error
}

func (w *fakeWriter) Name() string         { return w.name }
func (w *fakeWriter) Extensions() []string { return w.exts }
func (w *fakeWriter) Write(_ string, doc *metadata.Document, opts WriteOptions) error {
	w.doc = doc
	w.opts = opts
	return w.err
}

func TestReaderForOrderAndCase(t *testing.T) {
	first := &fakeReader{name: "first", exts: []string{".flac"}}
	second := &fakeReader{name: "second", exts: []string{".flac", ".mp3"}}
	reg := New([]Reader{first, second}, nil)

	got, ok := reg.ReaderFor("track.FLAC")
	if !ok || got.Name() != "first" {
		t.Fatalf("got %v ok=%v", got, ok)
	}
	got, ok = reg.ReaderFor("track.mp3")
	if !ok || got.Name() != "second" {
		t.Fatalf("got %v ok=%v", got, ok)
	}
	if _, ok := reg.ReaderFor("track.ogg"); ok {
		t.Fatal("unexpected reader")
	}
	if _, ok := reg.ReaderFor("noext"); ok {
		t.Fatal("unexpected reader for empty extension")
	}
}

func TestWriterFor(t *testing.T) {
	writer := &fakeWriter{name: "w", exts: []string{".m4a"}}
	reg := New(nil, []Writer{writer})
	got, ok := reg.WriterFor("track.M4A")
	if !ok || got.Name() != "w" {
		t.Fatalf("got %v ok=%v", got, ok)
	}
	if _, ok := reg.WriterFor("track.mp3"); ok {
		t.Fatal("unexpected writer")
	}
}

func TestNewCopiesSlices(t *testing.T) {
	readers := []Reader{&fakeReader{name: "a", exts: []string{".a"}}}
	writers := []Writer{&fakeWriter{name: "b", exts: []string{".b"}}}
	reg := New(readers, writers)
	readers[0] = &fakeReader{name: "changed", exts: []string{".a"}}
	writers[0] = &fakeWriter{name: "changed", exts: []string{".b"}}
	if got := reg.Readers()[0].Name(); got != "a" {
		t.Fatalf("reader mutated: %q", got)
	}
	if got := reg.Writers()[0].Name(); got != "b" {
		t.Fatalf("writer mutated: %q", got)
	}
	reg.Readers()[0] = nil
	if reg.Readers()[0] == nil {
		t.Fatal("Readers must return a copy")
	}
}

func TestReadErrorPropagates(t *testing.T) {
	want := errors.New("boom")
	reader := &fakeReader{name: "x", exts: []string{".x"}, err: want}
	reg := New([]Reader{reader}, nil)
	got, ok := reg.ReaderFor("a.x")
	if !ok {
		t.Fatal("reader not found")
	}
	if _, err := got.Read("a.x", ReadOptions{}); !errors.Is(err, want) {
		t.Fatalf("err = %v", err)
	}
}
