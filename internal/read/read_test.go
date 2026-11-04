package read

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dhowden/tag"

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
func TestCoverFilenameAndName(t *testing.T) {
	cases := []struct {
		ext, mime, want string
	}{
		{"jpg", "", "cover.jpg"},
		{"png", "", "cover.png"},
		{"", "image/jpeg", "cover.jpg"},
		{"", "image/png", "cover.png"},
		{"webp", "", "cover.webp"},
	}
	for _, tc := range cases {
		if got := CoverFilename(&tag.Picture{Ext: tc.ext, MIMEType: tc.mime}); got != tc.want {
			t.Errorf("CoverFilename(%q,%q) = %q, want %q", tc.ext, tc.mime, got, tc.want)
		}
	}
	if got := CoverName("folder.png", ""); got != "cover.png" {
		t.Errorf("CoverName = %q", got)
	}
	if got := CoverName("", "image/webp"); got != "cover.webp" {
		t.Errorf("CoverName webp = %q", got)
	}
	if got := CoverName("", "application/octet-stream"); got != "cover.octet-stream" {
		t.Errorf("CoverName unknown = %q", got)
	}
}

func TestGetStr(t *testing.T) {
	cases := []struct {
		in   interface{}
		want string
	}{
		{"x", "x"},
		{[]string{"x", "y"}, "x"},
		{int(3), "3"},
		{int64(4), "4"},
		{true, ""},
	}
	for _, tc := range cases {
		if got := getStr(tc.in); got != tc.want {
			t.Errorf("getStr(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeKeyInvalid(t *testing.T) {
	if _, ok := NormalizeKey("\xff"); ok {
		t.Fatal("invalid UTF-8 must be rejected")
	}
}
func TestReadHelpersExtra(t *testing.T) {
	for _, key := range []string{"DISCTOTAL", "TOTALDISCS", "DISCS", "TOTALDISC"} {
		if !isDiscTotalKey(key) {
			t.Errorf("%s must be disc total", key)
		}
	}
	if isDiscTotalKey("X") {
		t.Fatal("X is not disc total")
	}
	if got := CoverName("x.jpeg", ""); got != "cover.jpeg" {
		t.Fatalf("CoverName = %q", got)
	}
	if got := FieldValue(map[string]string{"A": "1"}, "B", "A"); got != "1" {
		t.Fatalf("FieldValue = %q", got)
	}
	if a, b := SplitPair(""); a != "" || b != "" {
		t.Fatalf("empty SplitPair = %q/%q", a, b)
	}
	if a, b := SplitPair("3"); a != "3" || b != "" {
		t.Fatalf("plain SplitPair = %q/%q", a, b)
	}
	if _, err := ReadMagic(filepath.Join(t.TempDir(), "missing"), 4); err == nil {
		t.Fatal("missing magic must fail")
	}
}
