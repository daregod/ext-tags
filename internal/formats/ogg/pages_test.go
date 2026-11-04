package ogg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/vorbis"
)

func TestPageSequences(t *testing.T) {
	for _, name := range []string{"audio.ogg", "audio.opus", "audio.spx"} {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "formats", name))
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), name)
			if err := os.WriteFile(path, data, 0644); err != nil {
				t.Fatal(err)
			}
			if err := WriteComments(path, []vorbis.Comment{{Name: "TITLE", Value: "x"}}, nil, true); err != nil {
				t.Fatal(err)
			}
			out, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			pages, err := parsePages(out)
			if err != nil {
				t.Fatal(err)
			}
			for i, p := range pages {
				if p.seq != uint32(i) {
					t.Fatalf("page %d seq = %d", i, p.seq)
				}
			}
		})
	}
}
