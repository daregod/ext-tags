package registry_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/format"
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/registry"
)

func TestRoundTrip(t *testing.T) {
	cases := []struct {
		fixture string
		cover   bool
	}{
		{"audio.mp3", true},
		{"audio.flac", true},
		{"audio.ogg", false},
		{"audio.opus", false},
		{"audio.spx", false},
		{"audio-flac.oga", true},
		{"audio.wv", true},
		{"audio.ac3", true},
		{"audio.dts", true},
		{"audio-aac.m4a", true},
		{"audio-alac.m4a", true},
		{"audio.mkv", true},
		{"audio.wma", true},
		{"audio.wav", true},
		{"audio.aiff", true},
		{"audio.w64", true},
		{"audio.rf64", true},
	}
	for _, tc := range cases {
		t.Run(tc.fixture, func(t *testing.T) {
			src := filepath.Join("..", "..", "testdata", "formats", tc.fixture)
			path := filepath.Join(t.TempDir(), tc.fixture)
			data, err := os.ReadFile(src)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, data, 0644); err != nil {
				t.Fatal(err)
			}
			doc := metadata.New()
			doc.Fields["Title"] = "Roundtrip Title"
			doc.Fields["Artist"] = "Roundtrip Artist"
			doc.Fields["Album"] = "Roundtrip Album"
			doc.Fields["Genre"] = "Roundtrip Genre"
			doc.Fields["Year"] = "2026"
			if tc.cover {
				cover := &metadata.Cover{Filename: "cover.jpg", MIME: "image/jpeg", Data: []byte{0xFF, 0xD8, 0xFF}}
				doc.Cover = cover
				doc.Artwork[metadata.Front] = cover
			}

			reg := registry.NewDefault()
			writer, ok := reg.WriterFor(path)
			if !ok {
				t.Fatalf("no writer for %s", path)
			}
			if err := writer.Write(path, doc, format.WriteOptions{Replace: true}); err != nil {
				t.Fatal(err)
			}
			reader, ok := reg.ReaderFor(path)
			if !ok {
				t.Fatalf("no reader for %s", path)
			}
			got, err := reader.Read(path, format.ReadOptions{})
			if err != nil {
				t.Fatal(err)
			}
			for _, field := range []string{"Title", "Artist", "Album", "Genre", "Year"} {
				want := doc.Fields[field]
				if !matchesField(got, field, want) {
					t.Errorf("%s not %q (fields=%v)", field, want, got.Fields)
				}
			}
			if tc.cover {
				if got.Cover == nil || len(got.Cover.Data) == 0 {
					t.Errorf("front cover missing (artwork=%v)", got.Artwork)
				}
			}
		})
	}
}
func matchesField(doc *metadata.Document, field, want string) bool {
	aliases := map[string][]string{
		"Artist": {"Artist", "AlbumArtist"},
		"Year":   {"Year", "Date"},
	}
	names := aliases[field]
	if len(names) == 0 {
		names = []string{field}
	}
	for _, name := range names {
		for key, value := range doc.Fields {
			if strings.EqualFold(key, name) && strings.EqualFold(value, want) {
				return true
			}
		}
	}
	return false
}
