package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/registry"
)

func TestRenderTemplate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), path)
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	out := captureStdout(t, func() {
		if err := a.runRender([]string{path}, &renderOptions{template: `{{ .artist }} - {{ .title }}`}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "Fixture Artist - Fixture Title\n") {
		t.Fatalf("out = %q", out)
	}
}

func TestRenderJSONNullAndMissingTemplate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), path)
	a := &app{opts: &options{quiet: true, json: true}, registry: registry.NewDefault()}
	out := captureStdout(t, func() {
		if err := a.runRender([]string{path}, &renderOptions{template: `{{ .title }}`, null: true}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, `"value":"Fixture Title"`) || !strings.Contains(out, "\x00") {
		t.Fatalf("out = %q", out)
	}

	if err := a.runRender([]string{path}, &renderOptions{}); err == nil {
		t.Fatal("missing template must fail")
	}
}
