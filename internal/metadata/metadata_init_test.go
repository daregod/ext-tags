package metadata

import "testing"

func TestSetArtworkInitialisesMap(t *testing.T) {
	var doc Document
	back := &Cover{Filename: "back.jpg"}
	doc.SetArtwork(Back, back)
	if doc.Artwork == nil || doc.Artwork[Back] != back {
		t.Fatal("artwork map not initialised")
	}
	front := &Cover{Filename: "front.jpg"}
	doc.SetArtwork(Front, front)
	if doc.Cover != front {
		t.Fatal("front alias not set")
	}
}
