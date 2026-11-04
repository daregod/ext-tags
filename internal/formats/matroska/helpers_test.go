package matroska

import (
	"bytes"
	"testing"
)

func TestBuildHelpers(t *testing.T) {
	if data, ok := buildTags(nil); ok || len(data) != 0 {
		t.Fatal("empty tags")
	}
	data, ok := buildTags(map[string]string{"Title": "T", "Track": "3/12", "Disc": "1/2", "Custom": "x"})
	if !ok || len(data) == 0 {
		t.Fatal("tags missing")
	}
	if data, ok := buildAttachments(nil); ok || len(data) != 0 {
		t.Fatal("empty attachments")
	}
	covers := map[string]*Cover{
		"front": {Filename: "cover.jpg", MIME: "image/jpeg", Data: []byte{1}},
		"back":  {Filename: "", MIME: "image/jpeg", Data: []byte{2}},
		"bad":   {Data: nil},
	}
	data, ok = buildAttachments(covers)
	if !ok || len(data) == 0 {
		t.Fatal("attachments missing")
	}
	if !bytes.Contains(data, []byte("cover.jpg")) {
		t.Fatal("filename missing")
	}
	if a, b := split("3/12"); a != "3" || b != "12" {
		t.Fatalf("split = %q/%q", a, b)
	}
	if a, b := split("3"); a != "3" || b != "" {
		t.Fatalf("split = %q/%q", a, b)
	}
}
