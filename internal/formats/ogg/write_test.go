package ogg_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/ogg"
	"github.com/daregod/ext-tags/internal/formats/vorbis"
)

func TestWriteComments(t *testing.T) {
	for _, name := range []string{"audio.ogg", "audio.opus", "audio.spx"} {
		t.Run(name, func(t *testing.T) {
			src := filepath.Join("..", "..", "..", "testdata", "formats", name)
			data, err := os.ReadFile(src)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), name)
			if err := os.WriteFile(path, data, 0644); err != nil {
				t.Fatal(err)
			}
			comments := []vorbis.Comment{
				{Name: "TITLE", Value: "Ogg Title"},
				{Name: "ARTIST", Value: "Ogg Artist"},
			}
			var pictures []vorbis.Picture
			for _, typ := range []uint32{3, 4, 8, 6, 1} {
				pictures = append(pictures, vorbis.Picture{Type: typ, MIME: "image/jpeg", Data: []byte{byte(typ)}})
			}
			if err := ogg.WriteComments(path, comments, pictures, true); err != nil {
				t.Fatal(err)
			}

			_, got, err := ogg.ReadComments(path)
			if err != nil {
				t.Fatal(err)
			}
			var hasTitle, hasArtist bool
			pictureCount := 0
			for _, c := range got {
				switch strings.ToUpper(c.Name) {
				case "TITLE":
					hasTitle = c.Value == "Ogg Title"
				case "ARTIST":
					hasArtist = c.Value == "Ogg Artist"
				case "METADATA_BLOCK_PICTURE":
					pictureCount++
				}
			}
			if !hasTitle || !hasArtist || pictureCount != 5 {
				t.Fatalf("comments = %+v", got)
			}

			res, err := ogg.Read(path)
			if err != nil {
				t.Fatal(err)
			}
			if field(res.Fields, "title") != "Ogg Title" || field(res.Fields, "artist") != "Ogg Artist" {
				t.Fatalf("extracted fields = %v", res.Fields)
			}
		})
	}
}

func field(fields map[string]string, name string) string {
	for k, v := range fields {
		if strings.EqualFold(k, name) {
			return v
		}
	}
	return ""
}
