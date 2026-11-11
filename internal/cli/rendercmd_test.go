package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/ape"
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

func TestRenderExposesArtwork(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), path)
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	out := captureStdout(t, func() {
		if err := a.runRender([]string{path}, &renderOptions{template: `{{ .artwork }}|{{ .artworkcount }}|{{ .artworkfront }}`}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "front:cover.jpg") || !strings.Contains(out, "|1|cover.jpg") {
		t.Fatalf("artwork template = %q", out)
	}
}

func TestRenderRaw(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), path)
	tag := ape.NewTag()
	tag.AddText("Custom Tag", "raw value")
	tag.AddText("Title", "Tag title")
	data, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path+".tag", data, 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	out := captureStdout(t, func() {
		if err := a.runRender([]string{path}, &renderOptions{
			raw:      true,
			template: `{{ range $k, $v := .raw }}{{ println $k $v }}{{ end }}artworkcount={{ .artworkcount }} container={{ .containerartworkcount }}`,
		}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "Custom Tag raw value") || !strings.Contains(out, "Title Tag title") {
		t.Fatalf("raw = %q", out)
	}
	if !strings.Contains(out, "artworkcount=1") || !strings.Contains(out, "container=1") {
		t.Fatalf("raw artwork = %q", out)
	}
}

func TestRenderFieldsMap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), path)
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	out := captureStdout(t, func() {
		if err := a.runRender([]string{path}, &renderOptions{template: `{{ sanitize .title }}|{{ index .fields "Title" }}|{{ .vars }}`}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.HasPrefix(out, "Fixture Title|Fixture Title|") {
		t.Fatalf("fields = %q", out)
	}
}
