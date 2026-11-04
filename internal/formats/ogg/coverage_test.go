package ogg

import (
	"encoding/base64"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/vorbis"
)

func TestExtension(t *testing.T) {
	cases := map[string]string{
		"audio.ogg":      ".ogg",
		"audio.opus":     ".opus",
		"audio.spx":      ".spx",
		"audio-flac.oga": ".oga",
	}
	for name, want := range cases {
		if got := Extension(filepath.Join("..", "..", "..", "testdata", "formats", name)); got != want {
			t.Errorf("%s: %q, want %q", name, got, want)
		}
	}
	if got := Extension(filepath.Join("..", "..", "..", "testdata", "formats", "audio.mp3")); got != "" {
		t.Errorf("mp3: %q", got)
	}
}

func TestMergeComments(t *testing.T) {
	existing := []vorbis.Comment{
		{Name: "TITLE", Value: "old"},
		{Name: "ALBUM", Value: "album"},
		{Name: "METADATA_BLOCK_PICTURE", Value: "pic"},
	}
	incoming := []vorbis.Comment{{Name: "TITLE", Value: "new"}}
	got := mergeComments(existing, incoming, false)
	if len(got) != 2 || got[0].Value != "album" || got[1].Value != "new" {
		t.Fatalf("merge = %+v", got)
	}
	got = mergeComments(existing, incoming, true)
	if len(got) != 1 || got[0].Value != "new" {
		t.Fatalf("replace = %+v", got)
	}
}

func TestMergePictures(t *testing.T) {
	pic := vorbis.Picture{Type: 3, MIME: "image/jpeg", Data: []byte{1}}
	encoded := base64.StdEncoding.EncodeToString(vorbis.MarshalPicture(pic))
	existing := []vorbis.Comment{{Name: "METADATA_BLOCK_PICTURE", Value: encoded}}
	incoming := []vorbis.Picture{{Type: 4, Data: []byte{2}}}
	got, err := mergePictures("vorbis", existing, incoming, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Type != 3 || got[1].Type != 4 {
		t.Fatalf("merge = %+v", got)
	}
	got, err = mergePictures("vorbis", existing, incoming, true)
	if err != nil || len(got) != 1 || got[0].Type != 4 {
		t.Fatalf("replace = %+v, %v", got, err)
	}
}

func TestMergeFLACPictures(t *testing.T) {
	existing := []vorbis.Picture{{Type: 3}, {Type: 4}}
	incoming := []vorbis.Picture{{Type: 3, Data: []byte{1}}}
	got := mergeFLACPictures(existing, incoming)
	if len(got) != 2 || got[0].Type != 4 || got[1].Type != 3 {
		t.Fatalf("merge = %+v", got)
	}
}

func TestParseFlacPicture(t *testing.T) {
	raw := vorbis.MarshalPicture(vorbis.Picture{Type: 3, MIME: "image/png", Data: []byte{1}})
	if pic, ok := parseFlacPicture(raw); !ok || pic.MIME != "image/png" {
		t.Fatalf("picture = %+v, %v", pic, ok)
	}
	if _, ok := parseFlacPicture([]byte{1}); ok {
		t.Fatal("short picture must fail")
	}
}
func TestOggHelpers(t *testing.T) {
	if isOggFLAC(nil) {
		t.Fatal("nil packets")
	}
	if !isOggFLAC([]packet{{data: append([]byte{0x7F, 'F', 'L', 'A', 'C', 1, 0, 0, 0, 'f', 'L', 'a', 'C'}, make([]byte, 4)...)}}) {
		t.Fatal("valid FLAC packet")
	}
	if got := stringsIndexByte("abc", 'b'); got != 1 {
		t.Fatalf("index = %d", got)
	}
	if got := stringsIndexByte("abc", 'z'); got != -1 {
		t.Fatalf("index = %d", got)
	}
	if _, err := readPrefix(filepath.Join(t.TempDir(), "missing"), 4); err == nil {
		t.Fatal("missing prefix must fail")
	}
	if _, _, err := ReadComments(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("missing comments must fail")
	}
}
