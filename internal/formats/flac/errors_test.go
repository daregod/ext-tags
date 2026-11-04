package flac

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/vorbis"
)

func TestReadWriteErrors(t *testing.T) {
	dir := t.TempDir()
	if _, _, err := Read(filepath.Join(dir, "missing.flac")); err == nil {
		t.Fatal("Read missing must fail")
	}
	if err := Write(filepath.Join(dir, "missing.flac"), nil, nil, true); err == nil {
		t.Fatal("Write missing must fail")
	}
	bad := filepath.Join(dir, "bad.flac")
	if err := os.WriteFile(bad, []byte("not flac"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Read(bad); err == nil {
		t.Fatal("Read bad must fail")
	}
	if err := Write(bad, nil, nil, true); err == nil {
		t.Fatal("Write bad must fail")
	}
}

func TestWriteMergeKeepsExisting(t *testing.T) {
	src := filepath.Join("..", "..", "..", "testdata", "formats", "audio.flac")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "merge.flac")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	comments := []vorbis.Comment{{Name: "TITLE", Value: "New Title"}}
	pictures := []vorbis.Picture{{Type: 3, MIME: "image/jpeg", Data: []byte{1}}}
	if err := Write(path, comments, pictures, false); err != nil {
		t.Fatal(err)
	}
	gotComments, gotPictures, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	values := map[string]string{}
	for _, c := range gotComments {
		values[c.Name] = c.Value
	}
	upper := map[string]string{}
	for key, value := range values {
		upper[strings.ToUpper(key)] = value
	}
	if upper["TITLE"] != "New Title" {
		t.Fatalf("comments = %+v", gotComments)
	}
	if upper["ALBUM"] == "" {
		t.Fatalf("existing album lost: %+v", gotComments)
	}
	if len(gotPictures) == 0 {
		t.Fatal("pictures missing")
	}
}
