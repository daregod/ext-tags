package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/registry"
	"github.com/daregod/ext-tags/internal/tagfile"
)

func TestApplyTagOnly(t *testing.T) {
	dir := t.TempDir()
	tagPath := filepath.Join(dir, "track.flac.tag")
	tag := ape.NewTag()
	tag.AddText("TITLE", "Song")
	tag.AddText("ARTIST", "Artist")
	tag.AddText("Track", "1/4")
	data, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tagPath, data, 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "out.tag")
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	if err := a.runApply(&applyOptions{tag: tagPath, output: out}); err != nil {
		t.Fatal(err)
	}
	parsed, err := ape.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if item, ok := parsed.Get("Title"); !ok || string(item.Value) != "Song" {
		t.Fatalf("Title = %q, %v", item.Value, ok)
	}
}

func TestApplyDiffDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	tagPath := filepath.Join(dir, "track.flac.tag")
	tag := ape.NewTag()
	tag.AddText("Title", "Old")
	data, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tagPath, data, 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{}, registry: registry.NewDefault()}
	out := captureStdout(t, func() {
		if err := a.runApply(&applyOptions{tag: tagPath, diff: true}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "DIFF") {
		t.Fatalf("diff = %q", out)
	}
	after, err := os.ReadFile(tagPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, data) {
		t.Fatal("apply --diff modified the tag")
	}
}

func TestApplyAlbumOverlay(t *testing.T) {
	dir := t.TempDir()
	tagPath := filepath.Join(dir, "track.flac.tag")
	tag := ape.NewTag()
	tag.AddText("Title", "Tag title")
	data, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tagPath, data, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "album.yaml"), []byte("album:\n  Album: Album title\n  Title: Album track\n"), 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "out.tag")
	a := &app{opts: &options{}, registry: registry.NewDefault()}
	if err := a.runApply(&applyOptions{tag: tagPath, output: out}); err != nil {
		t.Fatal(err)
	}
	parsed, err := ape.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	fields := tagfile.ExistingFields(parsed)
	if fields["Album"] != "Album title" || fields["Title"] != "Album track" {
		t.Fatalf("fields = %v", fields)
	}
}
