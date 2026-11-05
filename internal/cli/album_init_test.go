package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/album"
	"github.com/daregod/ext-tags/internal/registry"
)

func TestAlbumInit(t *testing.T) {
	dir := t.TempDir()
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), filepath.Join(dir, "track.flac"))
	if err := os.WriteFile(filepath.Join(dir, "cover.jpg"), []byte{1, 2, 3}, 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	if err := a.runAlbumInit([]string{dir}, &albumInitOptions{force: true}); err != nil {
		t.Fatal(err)
	}
	cfg, err := album.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil || cfg.Album["Album"] != "Fixture Album" {
		t.Fatalf("album = %+v", cfg)
	}
	if cfg.Artwork["front"] != "cover.jpg" {
		t.Fatalf("artwork = %+v", cfg.Artwork)
	}
}

func TestAlbumInitDryRunAndForce(t *testing.T) {
	dir := t.TempDir()
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), filepath.Join(dir, "track.flac"))
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	if err := a.runAlbumInit([]string{dir}, &albumInitOptions{dryRun: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "album.yaml")); !os.IsNotExist(err) {
		t.Fatal("dry-run wrote album.yaml")
	}
	if err := a.runAlbumInit([]string{dir}, &albumInitOptions{force: true}); err != nil {
		t.Fatal(err)
	}
	a = &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	if err := a.runAlbumInit([]string{dir}, &albumInitOptions{}); err == nil {
		t.Fatal("overwrite without force must fail")
	}
}
