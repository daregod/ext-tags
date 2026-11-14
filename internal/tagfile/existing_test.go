package tagfile

import (
	"testing"

	"github.com/daregod/ext-tags/internal/formats/ape"
)

func TestExistingPicturesAndArtwork(t *testing.T) {
	tag := &ape.ParsedTag{Items: []ape.Item{
		{Flags: 0, Key: "Title", Value: []byte("text")},
		{Flags: 2, Key: "Cover Art (Front)", Value: []byte("front.png\x00\x01")},
		{Flags: 2, Key: "Cover Art (Back)", Value: []byte("back.jpg\x00\x02")},
		{Flags: 2, Key: "Custom Art", Value: []byte("custom.webp\x00\x03")},
		{Flags: 2, Key: "Broken", Value: []byte{0}},
	}}
	pictures := ExistingPictures(tag)
	if len(pictures) != 3 {
		t.Fatalf("pictures = %+v", pictures)
	}
	if pictures[0].Slot != "front" || pictures[0].MIME != "image/png" || string(pictures[0].Data) != "\x01" {
		t.Fatalf("front = %+v", pictures[0])
	}
	if pictures[2].Slot != "" || pictures[2].MIME != "image/webp" {
		t.Fatalf("custom = %+v", pictures[2])
	}
	art := ExistingArtwork(tag)
	if art["front"] == nil || art["back"] == nil || len(art) != 2 {
		t.Fatalf("artwork = %+v", art)
	}
	if ExistingPictures(nil) != nil {
		t.Fatal("nil pictures")
	}
	if len(ExistingArtwork(nil)) != 0 {
		t.Fatal("nil artwork")
	}
}

func TestCoverValueAndMIME(t *testing.T) {
	name, data := splitCoverValue([]byte("cover.gif\x00abc"))
	if name != "cover.gif" || string(data) != "abc" {
		t.Fatalf("split = %q %q", name, data)
	}
	name, data = splitCoverValue([]byte("raw"))
	if name != "" || string(data) != "raw" {
		t.Fatalf("raw = %q %q", name, data)
	}
	cases := map[string]string{".png": "image/png", ".webp": "image/webp", ".gif": "image/gif", ".jpg": "image/jpeg"}
	for ext, want := range cases {
		if got := coverMIME("cover" + ext); got != want {
			t.Fatalf("%s = %s", ext, got)
		}
	}
}
