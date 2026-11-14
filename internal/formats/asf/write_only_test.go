package asf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func copyASF(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "formats", "audio.wma"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "track.wma")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestWriteOnly(t *testing.T) {
	path := copyASF(t)
	if err := WriteOnly(path, map[string]string{"Title": "New", "Track": "2/10"}, map[string]*Cover{
		"front": {Filename: "front.jpg", MIME: "image/jpeg", Data: []byte{1}},
		"back":  {Filename: "back.jpg", MIME: "image/jpeg", Data: []byte{2}},
	}, []string{"Title", "Track", "Cover Art (Front)"}); err != nil {
		t.Fatal(err)
	}
	doc, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(doc.Fields["TITLE"], "New") || doc.Fields["TRACKNUMBER"] != "2/10" {
		t.Fatalf("fields = %+v", doc.Fields)
	}
	if doc.Artwork["front"] == nil || doc.Artwork["back"] == nil {
		t.Fatalf("artwork = %+v", doc.Artwork)
	}
	if err := WriteOnly(path, map[string]string{"Title": "New"}, nil, []string{"Track", "Cover Art (Back)"}); err != nil {
		t.Fatal(err)
	}
	doc, err = Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Fields["TRACKNUMBER"] != "" || doc.Artwork["back"] != nil {
		t.Fatalf("removal = %+v %+v", doc.Fields, doc.Artwork)
	}
	if err := WriteOnly(path, nil, nil, []string{"NoSuchField"}); err == nil {
		t.Fatal("expected unsupported field")
	}
}

func TestASFArtworkKind(t *testing.T) {
	cases := map[string]string{
		"Cover Art (Front)": "front",
		"Cover Art (Back)":  "back",
		"Artist Picture":    "artist",
		"Cover Art (Disc)":  "disc",
		"Cover Art (Icon)":  "icon",
		"other":             "",
	}
	for in, want := range cases {
		if got := asfArtworkKind(in); got != want {
			t.Fatalf("%s = %s", in, got)
		}
	}
}
