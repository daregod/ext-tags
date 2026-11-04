package read

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/metadata"
)

func TestTrackNumbers(t *testing.T) {
	num, total := TrackNumbers(map[string]interface{}{
		"TRACKNUMBER": "02",
		"TRACKTOTAL":  "16",
	})
	if num != "02" || total != "16" {
		t.Fatalf("TrackNumbers = %q/%q, want 02/16", num, total)
	}
}

func TestTrackNumbersSlash(t *testing.T) {
	num, total := TrackNumbers(map[string]interface{}{"TRCK": "5/12"})
	if num != "5" || total != "12" {
		t.Fatalf("TrackNumbers = %q/%q, want 5/12", num, total)
	}
}

func TestDiscNumbers(t *testing.T) {
	num, total := DiscNumbers(map[string]interface{}{"TPOS": "1/2"})
	if num != "1" || total != "2" {
		t.Fatalf("DiscNumbers = %q/%q, want 1/2", num, total)
	}
}

func TestNormalizeKeyMP4(t *testing.T) {
	key, ok := NormalizeKey("\xa9nam")
	if !ok || key != "©nam" {
		t.Fatalf("NormalizeKey = %q/%v, want ©nam/true", key, ok)
	}
}

func TestStringify(t *testing.T) {
	val, ok := Stringify([]string{" a ", "b", "a", ""})
	if !ok || val != "a; b" {
		t.Fatalf("Stringify = %q/%v, want a; b/true", val, ok)
	}
}

func TestIsSupported(t *testing.T) {
	exts := []string{".mp3", ".flac"}
	if !IsSupported("x/y/track.MP3", exts) {
		t.Fatal("MP3 must be supported")
	}
	if IsSupported("x/y/track.wmv", exts) {
		t.Fatal("wmv must not be supported")
	}
}
func TestTaglibFixtures(t *testing.T) {
	for _, name := range []string{"audio.mp3", "audio.flac", "audio-aac.m4a", "audio-alac.m4a", "audio.aac"} {
		t.Run(name, func(t *testing.T) {
			doc, err := Taglib(filepath.Join("..", "..", "testdata", "formats", name))
			if err != nil {
				t.Fatal(err)
			}
			if field(doc, "title") == "" {
				t.Fatalf("missing title: %v", doc.Fields)
			}
			if doc.Format == "" {
				t.Fatal("missing format")
			}
		})
	}
}

func TestTaglibNoTags(t *testing.T) {
	_, err := Taglib(filepath.Join("..", "..", "testdata", "formats", "audio-empty.wav"))
	if !errors.Is(err, ErrNoTags) {
		t.Fatalf("err = %v", err)
	}
}

func TestHelpers(t *testing.T) {
	if got := CoverName("", "image/png"); got != "cover.png" {
		t.Fatalf("CoverName = %q", got)
	}
	if got := FieldValue(map[string]string{"A": "1", "B": "2"}, "A", "B"); got != "1" {
		t.Fatalf("FieldValue = %q", got)
	}
	if num, total := SplitPair("3/12"); num != "3" || total != "12" {
		t.Fatalf("SplitPair = %q/%q", num, total)
	}
	if magic, err := ReadMagic(filepath.Join("..", "..", "testdata", "formats", "audio.mp3"), 4); err != nil || len(magic) != 4 {
		t.Fatalf("ReadMagic = % X, %v", magic, err)
	}
	if !IsArtworkField("METADATA_BLOCK_PICTURE") {
		t.Fatal("artwork key not detected")
	}
}

func field(doc *metadata.Document, name string) string {
	for key, value := range doc.Fields {
		if strings.EqualFold(key, name) {
			return value
		}
	}
	return ""
}
