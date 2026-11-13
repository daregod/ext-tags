package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/daregod/ext-tags/internal/registry"
)

func TestFixTagCase(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("case-insensitive filesystem")
	}
	dir := t.TempDir()
	audio := filepath.Join(dir, "Track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), audio)
	oldTag := filepath.Join(dir, "track.flac.TAG")
	if err := os.WriteFile(oldTag, []byte("tag"), 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	if err := a.runFixTag([]string{audio}, &fixTagOptions{apply: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(audio + ".tag"); err != nil {
		t.Fatalf("target tag: %v", err)
	}
	if _, err := os.Stat(oldTag); !os.IsNotExist(err) {
		t.Fatalf("old tag still exists: %v", err)
	}
}

func TestFixTagChangedExtension(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "Track.mp3")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), audio)
	oldTag := filepath.Join(dir, "Track.flac.tag")
	if err := os.WriteFile(oldTag, []byte("tag"), 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	if err := a.runFixTag([]string{audio}, &fixTagOptions{apply: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(audio + ".tag"); err != nil {
		t.Fatalf("target tag: %v", err)
	}
	if _, err := os.Stat(oldTag); !os.IsNotExist(err) {
		t.Fatalf("old tag still exists: %v", err)
	}
}

func TestFixTagDoesNotStealExistingSource(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "Track.mp3")
	other := filepath.Join(dir, "Track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), audio)
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), other)
	oldTag := other + ".tag"
	if err := os.WriteFile(oldTag, []byte("tag"), 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	if err := a.runFixTag([]string{audio}, &fixTagOptions{apply: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(audio + ".tag"); !os.IsNotExist(err) {
		t.Fatal("tag for another audio file was moved")
	}
	if _, err := os.Stat(oldTag); err != nil {
		t.Fatal("source tag is missing")
	}
}
