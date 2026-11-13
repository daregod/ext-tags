package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/daregod/ext-tags/internal/config"
)

func TestPreferExisting(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	tag := audio + ".tag"
	if err := os.WriteFile(audio, []byte("audio"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tag, []byte("tag"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Files.Prefer = "tag"
	if !preferExisting(cfg, audio) {
		t.Fatal("prefer tag")
	}
	cfg.Files.Prefer = "container"
	if preferExisting(cfg, audio) {
		t.Fatal("prefer container")
	}
	cfg.Files.Prefer = "newest"
	old := time.Now().Add(-time.Hour)
	now := time.Now()
	if err := os.Chtimes(audio, old, old); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(tag, now, now); err != nil {
		t.Fatal(err)
	}
	if !preferExisting(cfg, audio) {
		t.Fatal("newest tag")
	}
	if err := os.Chtimes(audio, now.Add(time.Hour), now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if preferExisting(cfg, audio) {
		t.Fatal("newest audio")
	}
}
