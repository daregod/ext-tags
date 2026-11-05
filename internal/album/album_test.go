package album

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/metadata"
)

func TestLoadApply(t *testing.T) {
	dir := t.TempDir()
	config := `album:
  Album: Album Override
  Album artist: Artist Override
tracks:
  "3":
    Title: Track Override
    Artist: Track Artist
    Genre: Track Genre
artwork:
  front: cover.jpg
`
	if err := os.WriteFile(filepath.Join(dir, "album.yaml"), []byte(config), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cover.jpg"), []byte{0xFF, 0xD8}, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("album config not loaded")
	}
	doc := metadata.New()
	doc.Fields["Album"] = "Old"
	doc.Fields["Track"] = "3/8"
	cfg.Apply(doc, filepath.Join(dir, "3. track.flac"))
	if doc.Fields["Album"] != "Album Override" || doc.Fields["Title"] != "Track Override" || doc.Fields["Genre"] != "Track Genre" {
		t.Fatalf("fields = %v", doc.Fields)
	}
	if doc.Vars["tracknumber"] != "3" || doc.Vars["totaltracks"] != "8" {
		t.Fatalf("vars = %v", doc.Vars)
	}
	covers, err := cfg.ArtworkCovers()
	if err != nil {
		t.Fatal(err)
	}
	if covers["front"] == nil || covers["front"].MIME != "image/jpeg" {
		t.Fatalf("covers = %+v", covers)
	}
}

func TestLoadMissingAndBad(t *testing.T) {
	if cfg, err := Load(t.TempDir()); err != nil || cfg != nil {
		t.Fatalf("missing = %+v, %v", cfg, err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "album.yaml"), []byte("album: [\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir); err == nil {
		t.Fatal("bad yaml must fail")
	}
}
func TestLoadBOMAndCRLF(t *testing.T) {
	dir := t.TempDir()
	data := "\xEF\xBB\xBFalbum:\r\n  Album: BOM Album\r\n"
	if err := os.WriteFile(filepath.Join(dir, "album.yaml"), []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil || cfg.Album["Album"] != "BOM Album" {
		t.Fatalf("album = %+v", cfg)
	}
}
