package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/album"
	"github.com/daregod/ext-tags/internal/formats/ape"
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

func TestAlbumInitMergesExistingTag(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), audio)
	tag := ape.NewTag()
	tag.AddText("Releasetime", "2015-08-30")
	tag.AddText("Year", "2015")
	data, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(audio+".tag", data, 0644); err != nil {
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
	if cfg.Album["Releasetime"] != "2015-08-30" {
		t.Fatalf("Releasetime = %q, want 2015-08-30", cfg.Album["Releasetime"])
	}
}

func TestAlbumInitMergeAlbum(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), audio)
	existing := `album:
  Album: Manual Album
tracks:
  "1":
    Comment: ""
artwork:
  front: manual.jpg
`
	if err := os.WriteFile(filepath.Join(dir, "album.yaml"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	if err := a.runAlbumInit([]string{dir}, &albumInitOptions{mergeAlbum: true}); err != nil {
		t.Fatal(err)
	}
	cfg, err := album.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Album["Album"] != "Manual Album" {
		t.Fatalf("Album = %q, want Manual Album", cfg.Album["Album"])
	}
	if _, ok := cfg.Tracks["1"]["Comment"]; !ok || cfg.Tracks["1"]["Comment"] != "" {
		t.Fatalf("explicit empty Comment not preserved: %#v", cfg.Tracks["1"])
	}
	if cfg.Album["Title"] != "Fixture Title" {
		t.Fatalf("generated album fields were not added: %#v", cfg.Album)
	}
	if cfg.Artwork["front"] != "manual.jpg" {
		t.Fatalf("artwork = %q", cfg.Artwork["front"])
	}

	a = &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	if err := a.runAlbumInit([]string{dir}, &albumInitOptions{mergeAlbum: true, force: true}); err == nil {
		t.Fatal("merge-album + force must fail")
	}
}
