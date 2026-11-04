package vorbis

import (
	"bytes"
	"testing"
)

func TestCommentsRoundTrip(t *testing.T) {
	in := []Comment{{Name: "TITLE", Value: "Title"}, {Name: "ARTIST", Value: "Artist"}}
	data := MarshalComments("vendor", in, true)
	vendor, got, err := ParseComments(data, true)
	if err != nil {
		t.Fatal(err)
	}
	if vendor != "vendor" || len(got) != 2 || got[0].Name != "TITLE" || got[1].Value != "Artist" {
		t.Fatalf("vendor=%q comments=%+v", vendor, got)
	}
}

func TestParseCommentsErrors(t *testing.T) {
	if _, _, err := ParseComments([]byte{1, 2}, false); err == nil {
		t.Fatal("short comments must fail")
	}
	if _, _, err := ParseComments([]byte{20, 0, 0, 0}, false); err == nil {
		t.Fatal("bad vendor length must fail")
	}
}

func TestPictureRoundTrip(t *testing.T) {
	in := Picture{Type: 3, MIME: "image/jpeg", Description: "front", Width: 32, Height: 32, Depth: 24, Data: []byte{0xFF, 0xD8}}
	got, err := ParsePicture(MarshalPicture(in))
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != in.Type || got.MIME != in.MIME || got.Description != in.Description || !bytes.Equal(got.Data, in.Data) {
		t.Fatalf("got = %+v", got)
	}
}

func TestParsePictureErrors(t *testing.T) {
	if _, err := ParsePicture([]byte{1, 2}); err == nil {
		t.Fatal("short picture must fail")
	}
	if _, err := ParsePicture(MarshalPicture(Picture{Type: 3, MIME: "image/jpeg", Data: []byte{1}})[:10]); err == nil {
		t.Fatal("truncated picture must fail")
	}
}

func TestSplit(t *testing.T) {
	name, value, ok := Split("TITLE=Song")
	if !ok || name != "TITLE" || value != "Song" {
		t.Fatalf("Split = %q/%q/%v", name, value, ok)
	}
	if _, _, ok := Split("no-equals"); ok {
		t.Fatal("line without = must not split")
	}
}
