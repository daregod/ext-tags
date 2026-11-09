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
func TestGenerateFillsTotals(t *testing.T) {
	cfg := Generate([]Input{
		{AudioPath: "/music/a.flac", Fields: map[string]string{"Title": "A", "Track": "1"}},
		{AudioPath: "/music/b.flac", Fields: map[string]string{"Title": "B", "Track": "2"}},
	}, nil)
	if cfg.Tracks["1"]["Track"] != "1/2" || cfg.Tracks["2"]["Track"] != "2/2" {
		t.Fatalf("tracks = %v", cfg.Tracks)
	}
	cfg = Generate([]Input{{AudioPath: "/music/a.flac", Fields: map[string]string{"Title": "A"}}}, nil)
	if cfg.Album["Track"] != "1/1" {
		t.Fatalf("single = %v", cfg.Album)
	}
}

func TestMarshalUsesBlockScalarsForMultiline(t *testing.T) {
	cfg := &Config{
		Album:  map[string]string{"Www": "https://a\r\nhttps://b  \r\n"},
		Tracks: map[string]map[string]string{"01": {"track_url": "https://c\nhttps://d"}},
	}
	data, err := cfg.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, "Www: |") || !strings.Contains(text, "track_url: |") {
		t.Fatalf("block scalar missing:\n%s", text)
	}
	if strings.Contains(text, "\\r\\n") || strings.Contains(text, "https://b  ") {
		t.Fatalf("multiline value was not normalized:\n%s", text)
	}
	if !strings.Contains(text, "    https://a\n    https://b\n") {
		t.Fatalf("Www block =\n%s", text)
	}
}

func TestGenerateOptionsIgnoreAndOrder(t *testing.T) {
	inputs := []Input{
		{AudioPath: "/music/1.flac", Fields: map[string]string{"Album": "A", "Artist": "AA", "Www": "https://a\nhttps://b", "encodedby": "x", "Rating": "3", "Title": "One", "Track": "1/10"}},
		{AudioPath: "/music/2.flac", Fields: map[string]string{"Album": "A", "Artist": "AA", "Www": "https://a\nhttps://b", "encodedby": "x", "Rating": "3", "Title": "Two", "Track": "2/10"}},
		{AudioPath: "/music/10.flac", Fields: map[string]string{"Album": "A", "Artist": "AA", "Www": "https://a\nhttps://b", "encodedby": "x", "Rating": "3", "Title": "Ten", "Track": "10/10"}},
	}
	cfg := GenerateWithOptions(inputs, nil, GenerateOptions{
		Ignore:     []string{"encodedby", "rating"},
		AlbumOrder: []string{"Album", "Artist", "Www"},
		TrackOrder: []string{"Track", "Title"},
	})
	data, err := cfg.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, "encodedby") || strings.Contains(text, "Rating") {
		t.Fatalf("ignored fields present:\n%s", text)
	}
	ai := strings.Index(text, "Album: A")
	if ai < 0 || strings.Index(text, "Artist: AA") < ai {
		t.Fatalf("album order not applied:\n%s", text)
	}
	one := strings.Index(text, `"1":`)
	two := strings.Index(text, `"2":`)
	ten := strings.Index(text, `"10":`)
	if one >= two || two >= ten {
		t.Fatalf("track keys not numeric sorted: %d %d %d\n%s", one, two, ten, text)
	}
	if strings.Index(text, "Track: 01/10") > strings.Index(text, "Title: One") {
		t.Fatalf("track field order not applied:\n%s", text)
	}
}
