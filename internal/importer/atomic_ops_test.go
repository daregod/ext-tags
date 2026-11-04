package importer

import (
	"errors"
	"io"
	"io/fs"
	"strings"
	"testing"
	"time"

	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/tagformat"
)

type fakeInfo struct{}

func (fakeInfo) Name() string       { return "track.wav" }
func (fakeInfo) Size() int64        { return 1 }
func (fakeInfo) Mode() fs.FileMode  { return 0644 }
func (fakeInfo) ModTime() time.Time { return time.Time{} }
func (fakeInfo) IsDir() bool        { return false }
func (fakeInfo) Sys() interface{}   { return nil }

type fakeOps struct {
	statErr        error
	copyErr        error
	createErr      error
	removeErr      error
	chmodErr       error
	renameAlways   error
	renameFailOnce bool
	renameCalls    int
	removed        int
}

func (o *fakeOps) Stat(string) (fs.FileInfo, error) { return fakeInfo{}, o.statErr }
func (o *fakeOps) CreateTemp(string, string) (string, io.Closer, error) {
	if o.createErr != nil {
		return "", nil, o.createErr
	}
	return "/tmp/track.wav", io.NopCloser(strings.NewReader("")), nil
}
func (o *fakeOps) Copy(string, string) error { return o.copyErr }
func (o *fakeOps) Remove(string) error {
	o.removed++
	return o.removeErr
}
func (o *fakeOps) Rename(string, string) error {
	o.renameCalls++
	if o.renameAlways != nil {
		return o.renameAlways
	}
	if o.renameFailOnce && o.renameCalls == 1 {
		return errors.New("rename failed")
	}
	return nil
}
func (o *fakeOps) Chmod(string, fs.FileMode) error { return o.chmodErr }

type fakeWriter struct{ err error }

func (w *fakeWriter) Name() string                                                   { return "fake" }
func (w *fakeWriter) Extensions() []string                                           { return []string{".wav"} }
func (w *fakeWriter) Write(string, *metadata.Document, tagformat.WriteOptions) error { return w.err }

type fakeReader struct {
	doc *metadata.Document
	err error
}

func (r *fakeReader) Name() string         { return "fake" }
func (r *fakeReader) Extensions() []string { return []string{".wav"} }
func (r *fakeReader) Read(string, tagformat.ReadOptions) (*metadata.Document, error) {
	return r.doc, r.err
}

func TestWriteAtomicFileOps(t *testing.T) {
	doc := metadata.New()
	doc.Fields["Title"] = "Song"
	reader := &fakeReader{doc: doc}
	writer := &fakeWriter{}
	base := &fakeOps{}

	tests := []struct {
		name string
		ops  *fakeOps
		opts tagformat.WriteOptions
		r    *fakeReader
		w    *fakeWriter
	}{
		{name: "stat error", ops: &fakeOps{statErr: errors.New("stat")}},
		{name: "backup error", ops: &fakeOps{copyErr: errors.New("copy")}, opts: tagformat.WriteOptions{Backup: true}},
		{name: "temp error", ops: &fakeOps{createErr: errors.New("temp")}},
		{name: "writer error", ops: base, w: &fakeWriter{err: errors.New("write")}},
		{name: "reader error", ops: base, r: &fakeReader{err: errors.New("read")}},
		{name: "verify error", ops: base, r: &fakeReader{doc: metadata.New()}},
		{name: "chmod error", ops: &fakeOps{chmodErr: errors.New("chmod")}},
		{name: "rename fallback error", ops: &fakeOps{renameAlways: errors.New("rename"), removeErr: errors.New("remove")}},
		{name: "rename fallback success", ops: &fakeOps{renameFailOnce: true}},
		{name: "success", ops: base},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := tt.w
			if w == nil {
				w = writer
			}
			r := tt.r
			if r == nil {
				r = reader
			}
			err := writeAtomic("/tmp/track.wav", w, r, doc, tt.opts, tt.ops)
			if tt.name == "success" || tt.name == "rename fallback success" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
