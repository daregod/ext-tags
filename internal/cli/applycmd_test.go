package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/registry"
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
