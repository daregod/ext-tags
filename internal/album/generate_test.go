package album

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	inputs := []Input{
		{AudioPath: "/music/1.flac", Fields: map[string]string{
			"Album": "A", "Album artist": "AA", "Title": "One", "Track": "1/2", "Encoder": "x",
		}},
		{AudioPath: "/music/2.flac", Fields: map[string]string{
			"Album": "A", "Album artist": "AA", "Title": "Two", "Track": "2/2", "Encoder": "x",
		}},
	}
	cfg := Generate(inputs, map[string]string{"front": "cover.jpg"})
	if cfg.Album["Album"] != "A" || cfg.Album["Album artist"] != "AA" {
		t.Fatalf("album = %v", cfg.Album)
	}
	if cfg.Album["Encoder"] != "" {
		t.Fatal("technical field leaked into album")
	}
	if len(cfg.Tracks) != 2 || cfg.Tracks["1"]["Title"] != "One" || cfg.Tracks["2"]["Track"] != "2/2" {
		t.Fatalf("tracks = %v", cfg.Tracks)
	}
	if cfg.Artwork["front"] != "cover.jpg" {
		t.Fatalf("artwork = %v", cfg.Artwork)
	}
}

func TestGenerateSingle(t *testing.T) {
	cfg := Generate([]Input{{AudioPath: "/music/single.flac", Fields: map[string]string{
		"Album": "Single", "Title": "Track", "Track": "1/1",
	}}}, nil)
	if cfg.Album["Track"] != "1/1" || len(cfg.Tracks) != 0 {
		t.Fatalf("album=%v tracks=%v", cfg.Album, cfg.Tracks)
	}
}

func TestSave(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{Album: map[string]string{"Album": "A"}}
	if _, err := cfg.Save(dir, false); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "album.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Album: A") {
		t.Fatalf("yaml = %s", data)
	}
	if _, err := cfg.Save(dir, false); err == nil {
		t.Fatal("overwrite must fail without force")
	}
	if _, err := cfg.Save(dir, true); err != nil {
		t.Fatal(err)
	}
}
