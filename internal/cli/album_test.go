package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/registry"
	"github.com/daregod/ext-tags/internal/rules"
)

func TestAlbumOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "3. track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), path)
	cover := []byte{0xFF, 0xD8, 1, 2, 3}
	if err := os.WriteFile(filepath.Join(dir, "cover.jpg"), cover, 0644); err != nil {
		t.Fatal(err)
	}
	albumYAML := `album:
  Album: Album Override
  Album artist: Album Artist
tracks:
  "3":
    Title: Track Override
artwork:
  front: cover.jpg
`
	if err := os.WriteFile(filepath.Join(dir, "album.yaml"), []byte(albumYAML), 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	if err := a.runExtract(a.extractCmd(), []string{path}); err != nil {
		t.Fatal(err)
	}
	tag, err := ape.ReadFile(path + ".tag")
	if err != nil {
		t.Fatal(err)
	}
	if item, ok := tag.Get("Album"); !ok || string(item.Value) != "Album Override" {
		t.Fatalf("album = %+v", tag.Items)
	}
	if item, ok := tag.Get("Title"); !ok || string(item.Value) != "Track Override" {
		t.Fatalf("title = %+v", tag.Items)
	}
	if item, ok := tag.Get("Cover Art (Front)"); !ok || !bytes.Contains(item.Value, cover) {
		t.Fatalf("cover = % X", item.Value)
	}
}

func TestAlbumInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), path)
	if err := os.WriteFile(filepath.Join(dir, "album.yaml"), []byte("album: [\n"), 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	engine, err := rules.New(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	a.processFile(path, testConfig(), engine)
	if !a.failed {
		t.Fatal("invalid album.yaml must fail")
	}
}
