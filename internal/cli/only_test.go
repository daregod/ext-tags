package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/formats/matroska"
	"github.com/daregod/ext-tags/internal/read"
	"github.com/daregod/ext-tags/internal/registry"
)

func TestImportOnlyMP3(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.mp3")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.mp3"), audio)
	tagPath := audio + ".tag"
	write := func(fields map[string]string) {
		tag := ape.NewTag()
		for key, value := range fields {
			tag.AddText(key, value)
		}
		data, err := tag.Serialize()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(tagPath, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	a := &app{opts: &options{}, registry: registry.NewDefault()}
	write(map[string]string{"Title": "Tag title", "Rating": "5", "Track": "1/10"})
	if err := a.runImport([]string{audio}, &importOptions{version: 3, verify: "off"}); err != nil {
		t.Fatal(err)
	}
	write(map[string]string{"Track": "2/10"})
	a.opts = &options{}
	if err := a.runImport([]string{audio}, &importOptions{version: 3, verify: "off", only: []string{"Rating"}}); err != nil {
		t.Fatal(err)
	}
	doc, err := read.Taglib(audio)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Fields["TIT2"] != "Tag title" {
		t.Fatalf("Title = %q", doc.Fields["TIT2"])
	}
	if _, ok := doc.Fields["Rating"]; ok {
		t.Fatalf("Rating was not removed: %v", doc.Fields)
	}
	if doc.Fields["TRCK"] != "1/10" {
		t.Fatalf("Track = %v", doc.Fields)
	}
}

func TestImportOnlyFLAC(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), audio)
	tagPath := audio + ".tag"
	write := func(fields map[string]string) {
		tag := ape.NewTag()
		for key, value := range fields {
			tag.AddText(key, value)
		}
		data, err := tag.Serialize()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(tagPath, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	a := &app{opts: &options{}, registry: registry.NewDefault()}
	write(map[string]string{"Title": "Tag title", "Rating": "5", "Track": "1/10"})
	if err := a.runImport([]string{audio}, &importOptions{version: 3, verify: "off"}); err != nil {
		t.Fatal(err)
	}
	write(map[string]string{"Track": "2/10"})
	a.opts = &options{}
	if err := a.runImport([]string{audio}, &importOptions{version: 3, verify: "off", only: []string{"Rating"}}); err != nil {
		t.Fatal(err)
	}
	doc, err := read.Taglib(audio)
	if err != nil {
		t.Fatal(err)
	}
	if read.FieldValue(doc.Fields, "TITLE") != "Tag title" {
		t.Fatalf("title = %v", doc.Fields)
	}
	if read.FieldValue(doc.Fields, "RATING") != "" {
		t.Fatalf("rating = %v", doc.Fields)
	}
	if doc.Fields["tracknumber"] != "1" || doc.Fields["tracktotal"] != "10" {
		t.Fatalf("track = %v", doc.Fields)
	}
}

func TestImportOnlyCoverRemoval(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.mp3")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.mp3"), audio)
	tag := ape.NewTag()
	tag.AddText("Title", "Tag title")
	data, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(audio+".tag", data, 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{}, registry: registry.NewDefault()}
	if err := a.runImport([]string{audio}, &importOptions{version: 3, verify: "off", only: []string{"Cover Art (Front)"}}); err != nil {
		t.Fatal(err)
	}
	doc, err := read.Taglib(audio)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Cover != nil {
		t.Fatalf("cover was not removed: %+v", doc.Cover)
	}
}

func TestImportOnlyCoverRemovalFLAC(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), audio)
	tag := ape.NewTag()
	tag.AddText("Title", "Tag title")
	data, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(audio+".tag", data, 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{}, registry: registry.NewDefault()}
	if err := a.runImport([]string{audio}, &importOptions{version: 3, verify: "off", only: []string{"Cover Art (Front)"}}); err != nil {
		t.Fatal(err)
	}
	doc, err := read.Taglib(audio)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Cover != nil {
		t.Fatalf("cover was not removed: %+v", doc.Cover)
	}
}

func TestImportOnlyMP4(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.m4a")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio-aac.m4a"), audio)
	tagPath := audio + ".tag"
	write := func(fields map[string]string) {
		tag := ape.NewTag()
		for key, value := range fields {
			tag.AddText(key, value)
		}
		data, err := tag.Serialize()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(tagPath, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	a := &app{opts: &options{}, registry: registry.NewDefault()}
	write(map[string]string{"Title": "Tag title", "Track": "1/10"})
	if err := a.runImport([]string{audio}, &importOptions{version: 3, verify: "off"}); err != nil {
		t.Fatal(err)
	}
	write(map[string]string{"Title": "Tag title"})
	a.opts = &options{}
	if err := a.runImport([]string{audio}, &importOptions{version: 3, verify: "off", only: []string{"Track"}}); err != nil {
		t.Fatal(err)
	}
	doc, err := read.Taglib(audio)
	if err != nil {
		t.Fatal(err)
	}
	if read.FieldValue(doc.Fields, "TITLE", "title") != "Tag title" {
		t.Fatalf("title = %v", doc.Fields)
	}
	if read.FieldValue(doc.Fields, "TRACKNUMBER", "tracknumber") != "" {
		t.Fatalf("track was not removed: %v", doc.Fields)
	}
}

func TestImportOnlyMatroska(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.mkv")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.mkv"), audio)
	tagPath := audio + ".tag"
	write := func(fields map[string]string) {
		tag := ape.NewTag()
		for key, value := range fields {
			tag.AddText(key, value)
		}
		data, err := tag.Serialize()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(tagPath, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	a := &app{opts: &options{}, registry: registry.NewDefault()}
	write(map[string]string{"Title": "Tag title", "Track": "1/10"})
	if err := a.runImport([]string{audio}, &importOptions{version: 3, verify: "off"}); err != nil {
		t.Fatal(err)
	}
	write(map[string]string{"Title": "Tag title"})
	a.opts = &options{}
	if err := a.runImport([]string{audio}, &importOptions{version: 3, verify: "off", only: []string{"Track"}}); err != nil {
		t.Fatal(err)
	}
	doc, err := matroska.Read(audio, nil)
	if err != nil {
		t.Fatal(err)
	}
	if read.FieldValue(doc.Fields, "TITLE", "title") != "Tag title" {
		t.Fatalf("title = %v", doc.Fields)
	}
	if read.FieldValue(doc.Fields, "PART_NUMBER", "TRACKNUMBER") != "" {
		t.Fatalf("track was not removed: %v", doc.Fields)
	}
}
