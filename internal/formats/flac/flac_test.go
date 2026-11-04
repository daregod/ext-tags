package flac

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/vorbis"
)

func TestWriteRead(t *testing.T) {
	src := filepath.Join("..", "..", "..", "testdata", "formats", "audio.flac")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "test.flac")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	_, audioBefore, err := parse(data)
	if err != nil {
		t.Fatal(err)
	}

	comments := []vorbis.Comment{
		{Name: "TITLE", Value: "Imported"},
		{Name: "ARTIST", Value: "Artist"},
	}
	var pictures []vorbis.Picture
	for _, typ := range []uint32{3, 4, 8, 6, 1} {
		pictures = append(pictures, vorbis.Picture{Type: typ, MIME: "image/jpeg", Data: []byte{byte(typ)}})
	}
	if err := Write(path, comments, pictures, true); err != nil {
		t.Fatal(err)
	}

	gotComments, gotPictures, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(gotComments) != 2 || gotComments[0].Name != "TITLE" || gotComments[0].Value != "Imported" {
		t.Fatalf("comments = %+v", gotComments)
	}
	if len(gotPictures) != 5 {
		t.Fatalf("pictures = %d", len(gotPictures))
	}
	for i, typ := range []uint32{3, 4, 8, 6, 1} {
		if gotPictures[i].Type != typ {
			t.Fatalf("picture %d type = %d", i, gotPictures[i].Type)
		}
	}

	_, audioAfter, err := parse(mustRead(t, path))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(audioBefore, audioAfter) {
		t.Fatal("audio frames changed")
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
