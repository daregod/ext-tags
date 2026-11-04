package metadata

import "testing"

func TestNewAndSetArtwork(t *testing.T) {
	doc := New()
	if doc.Fields == nil || doc.Vars == nil || doc.Artwork == nil {
		t.Fatal("maps must be initialised")
	}
	front := &Cover{Filename: "cover.jpg", MIME: "image/jpeg", Data: []byte{1}}
	doc.SetArtwork(Front, front)
	if doc.Cover != front || doc.Artwork[Front] != front {
		t.Fatal("front alias not set")
	}
	back := &Cover{Filename: "back.png", MIME: "image/png", Data: []byte{2}}
	doc.SetArtwork(Back, back)
	if doc.Cover != front {
		t.Fatal("non-front artwork must not replace Cover")
	}
	if doc.Artwork[Back] != back {
		t.Fatal("back artwork missing")
	}
}
