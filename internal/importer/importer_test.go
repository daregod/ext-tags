package importer_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dhowden/tag"

	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/importer"
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/registry"
	"github.com/daregod/ext-tags/internal/tagformat"
)

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func readTags(t *testing.T, path string) tag.Metadata {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	m, err := tag.ReadFrom(f)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestImportMP3(t *testing.T) {
	src := filepath.Join("..", "..", "testdata", "formats", "audio.mp3")
	coverData, err := os.ReadFile(filepath.Join("..", "..", "testdata", "formats", "cover.jpg"))
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	audio := filepath.Join(dir, "track.mp3")
	copyFile(t, src, audio)

	tagFile := ape.NewTag()
	tagFile.AddText("Title", "Imported Title")
	tagFile.AddText("Artist", "Imported Artist")
	tagFile.AddText("Track", "7")
	tagFile.AddText("Year", "1999")
	tagFile.AddText("Genre", "Rock")
	tagFile.AddText("Comment", "hello")
	tagFile.AddText("Catalog#", "CAT-1")
	tagFile.AddCoverArt("Cover Art (Front)", "cover.jpg", coverData)
	data, err := tagFile.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	tagPath := audio + ".tag"
	if err := os.WriteFile(tagPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	tf, err := importer.ReadTagFile(tagPath)
	if err != nil {
		t.Fatal(err)
	}
	if tf.Fields["Title"] != "Imported Title" || tf.Cover == nil || len(tf.Cover.Data) != len(coverData) {
		t.Fatalf("tag file = %+v", tf)
	}

	if err := importer.ImportMP3(audio, tf, importer.Options{Version: 3, Replace: true, Backup: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(audio + ".bak"); err != nil {
		t.Fatalf("backup missing: %v", err)
	}

	m := readTags(t, audio)
	if m.Title() != "Imported Title" || m.Artist() != "Imported Artist" || m.Comment() != "hello" {
		t.Fatalf("metadata = %q / %q / %q", m.Title(), m.Artist(), m.Comment())
	}
	if track, _ := m.Track(); track != 7 {
		t.Fatalf("track = %d", track)
	}
	if m.Year() != 1999 || m.Genre() != "Rock" {
		t.Fatalf("year/genre = %d / %q", m.Year(), m.Genre())
	}
	if pic := m.Picture(); pic == nil || len(pic.Data) != len(coverData) {
		t.Fatalf("picture = %+v", pic)
	}
	raw, ok := m.Raw()["TXXX"].(*tag.Comm)
	if !ok || raw.Description != "Catalog#" || raw.Text != "CAT-1" {
		t.Fatalf("TXXX = %+v (raw %v)", m.Raw()["TXXX"], m.Raw())
	}

	// Re-import as ID3v2.4 and change one field.
	tf.Fields["Title"] = "Updated Title"
	if err := importer.ImportMP3(audio, tf, importer.Options{Version: 4}); err != nil {
		t.Fatal(err)
	}
	if m := readTags(t, audio); m.Title() != "Updated Title" {
		t.Fatalf("title = %q", m.Title())
	}

	// ID3v2.2 and ID3v1.
	if err := importer.ImportMP3(audio, tf, importer.Options{Version: 2, V1: true, Replace: true}); err != nil {
		t.Fatal(err)
	}
	fileData, err := os.ReadFile(audio)
	if err != nil {
		t.Fatal(err)
	}
	if int(fileData[3]) != 2 {
		t.Fatalf("version = %d, want 2", fileData[3])
	}
	v1 := fileData[len(fileData)-128:]
	if string(v1[0:3]) != "TAG" {
		t.Fatalf("id3v1 marker = %q", v1[0:3])
	}
	if string(bytes.TrimRight(v1[3:33], "\x00")) != "Updated Title" {
		t.Fatalf("id3v1 title = %q", string(bytes.TrimRight(v1[3:33], "\x00")))
	}
	if m := readTags(t, audio); m.Title() != "Updated Title" {
		t.Fatalf("v2.2 title = %q", m.Title())
	}
}
func TestImportFLAC(t *testing.T) {
	src := filepath.Join("..", "..", "testdata", "formats", "audio.flac")
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	copyFile(t, src, audio)

	tf := &importer.TagFile{
		Fields: map[string]string{
			"Title": "Flac Title", "Artist": "Flac Artist", "Track": "3/12",
		},
		Covers: map[string]*importer.Cover{
			"front": {Filename: "front.jpg", MIME: "image/jpeg", Data: []byte{1}},
			"back":  {Filename: "back.jpg", MIME: "image/jpeg", Data: []byte{2}},
		},
	}
	if err := importer.ImportFLAC(audio, tf, importer.Options{Replace: true, Backup: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(audio + ".bak"); err != nil {
		t.Fatalf("backup missing: %v", err)
	}
	res, err := readAudio(t, audio)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ToLower(fieldValue(res.Fields, "title")) != "flac title" || strings.ToLower(fieldValue(res.Fields, "artist")) != "flac artist" {
		t.Fatalf("fields = %v", res.Fields)
	}
	if res.Cover == nil || len(res.Cover.Data) != 1 {
		t.Fatalf("front = %+v", res.Cover)
	}
}
func fieldValue(fields map[string]string, key string) string {
	for k, v := range fields {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}
func TestImportAPEv2(t *testing.T) {
	src := filepath.Join("..", "..", "testdata", "formats", "audio.wv")
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.wv")
	copyFile(t, src, audio)

	tf := &importer.TagFile{
		Fields: map[string]string{
			"Title": "Wv Title", "Artist": "Wv Artist", "Track": "3",
		},
		Covers: map[string]*importer.Cover{
			"front":  {Filename: "front.jpg", MIME: "image/jpeg", Data: []byte{1}},
			"back":   {Filename: "back.png", MIME: "image/png", Data: []byte{2}},
			"artist": {Filename: "artist.jpg", MIME: "image/jpeg", Data: []byte{3}},
			"disc":   {Filename: "disc.jpg", MIME: "image/jpeg", Data: []byte{4}},
			"icon":   {Filename: "icon.png", MIME: "image/png", Data: []byte{5}},
		},
	}
	if err := importer.ImportAPEv2(audio, tf, importer.Options{Replace: true, Backup: true}); err != nil {
		t.Fatal(err)
	}
	res, err := readAudio(t, audio)
	if err != nil {
		t.Fatal(err)
	}
	if res.Fields["Title"] != "Wv Title" || res.Fields["Artist"] != "Wv Artist" {
		t.Fatalf("fields = %v", res.Fields)
	}
	for _, kind := range []string{"front", "back", "artist", "disc", "icon"} {
		if res.Artwork[kind] == nil {
			t.Fatalf("missing %s artwork", kind)
		}
	}
}
func readAudio(t *testing.T, path string) (*metadata.Document, error) {
	t.Helper()
	reader, ok := registry.NewDefault().ReaderFor(path)
	if !ok {
		t.Fatalf("no reader for %s", path)
	}
	return reader.Read(path, tagformat.ReadOptions{})
}
