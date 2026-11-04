package ape

import (
	"bytes"
	"testing"
)

func TestParseRoundTrip(t *testing.T) {
	tag := NewTag()
	tag.AddText("Title", "Song")
	tag.AddText("Artist", "Artist")
	img := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	tag.AddCoverArt("Cover Art (Front)", "cover.jpg", img)

	data, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if !parsed.Prefix || !parsed.HasHeader {
		t.Fatalf("prefix=%v header=%v", parsed.Prefix, parsed.HasHeader)
	}
	if len(parsed.Items) != 3 {
		t.Fatalf("items = %d, want 3", len(parsed.Items))
	}
	if item, ok := parsed.Get("title"); !ok || string(item.Value) != "Song" {
		t.Fatalf("Title = %q/%v", item.Value, ok)
	}
	cover, ok := parsed.Get("Cover Art (Front)")
	if !ok {
		t.Fatal("cover missing")
	}
	if !bytes.HasPrefix(cover.Value, []byte("cover.jpg\x00")) || !bytes.HasSuffix(cover.Value, img) {
		t.Fatal("cover value malformed")
	}
}

func TestParseErrors(t *testing.T) {
	if _, err := Parse([]byte("short")); err == nil {
		t.Fatal("expected error for short data")
	}
	if _, err := Parse(make([]byte, 40)); err == nil {
		t.Fatal("expected error for missing footer")
	}
}
