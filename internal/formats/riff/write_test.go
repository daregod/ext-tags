package riff_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/riff"
	"github.com/daregod/ext-tags/internal/metadata"
)

func TestWriteRead(t *testing.T) {
	cases := []string{"audio.wav", "audio.rf64", "audio.w64", "audio.aiff"}
	for _, fixture := range cases {
		t.Run(fixture, func(t *testing.T) {
			src := filepath.Join("..", "..", "..", "testdata", "formats", fixture)
			path := filepath.Join(t.TempDir(), fixture)
			data, err := os.ReadFile(src)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, data, 0644); err != nil {
				t.Fatal(err)
			}
			doc := metadata.New()
			doc.Fields["Title"] = "RIFF Title"
			doc.Fields["Artist"] = "RIFF Artist"
			doc.Fields["Album"] = "RIFF Album"
			doc.Fields["Genre"] = "RIFF Genre"
			doc.Fields["Year"] = "2026"
			front := &metadata.Cover{Filename: "cover.jpg", MIME: "image/jpeg", Data: []byte{0xFF, 0xD8, 0xFF}}
			doc.Cover = front
			doc.Artwork[metadata.Front] = front

			if err := riff.Write(path, doc); err != nil {
				t.Fatal(err)
			}
			md, err := riff.Read(path)
			if err != nil {
				t.Fatal(err)
			}
			for _, field := range []string{"Title", "Artist", "Album", "Genre", "Year"} {
				if got := fieldValue(md, field); !strings.EqualFold(got, doc.Fields[field]) {
					t.Errorf("%s = %q, want %q", field, got, doc.Fields[field])
				}
			}
			if md.Cover == nil || len(md.Cover.Data) == 0 {
				t.Fatal("cover missing")
			}
		})
	}
}

func fieldValue(doc *metadata.Document, name string) string {
	for key, value := range doc.Fields {
		if strings.EqualFold(key, name) {
			return value
		}
	}
	return ""
}
