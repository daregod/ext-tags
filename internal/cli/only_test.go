package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/formats/asf"
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

func TestImportOnlyMatroskaFrontCoverRemoval(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.mkv")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio-attachments.mkv"), audio)
	tag := ape.NewTag()
	data, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(audio+".tag", data, 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{}, registry: registry.NewDefault()}
	before, err := matroska.ReadFile(audio)
	if err != nil {
		t.Fatal(err)
	}
	if matroska.FrontIndex(before.Attachments) < 0 {
		t.Skip("fixture has no front attachment")
	}
	if err := a.runImport([]string{audio}, &importOptions{version: 3, verify: "off", only: []string{"Cover Art (Front)"}}); err != nil {
		t.Fatal(err)
	}
	md, err := matroska.ReadFile(audio)
	if err != nil {
		t.Fatal(err)
	}
	if len(md.Attachments) != len(before.Attachments)-1 {
		t.Fatalf("attachments = %d, before %d", len(md.Attachments), len(before.Attachments))
	}
}

func TestImportOnlyMatroskaBackCoverRemoval(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.mkv")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.mkv"), audio)
	if err := matroska.Write(audio, map[string]string{"Title": "Title"}, map[string]*matroska.Cover{
		"front": {Filename: "cover.jpg", MIME: "image/jpeg", Data: []byte{1}},
		"back":  {Filename: "back.jpg", MIME: "image/jpeg", Data: []byte{2}},
	}); err != nil {
		t.Fatal(err)
	}
	tag := ape.NewTag()
	data, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(audio+".tag", data, 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{}, registry: registry.NewDefault()}
	if err := a.runImport([]string{audio}, &importOptions{version: 3, verify: "off", only: []string{"Cover Art (Back)"}}); err != nil {
		t.Fatal(err)
	}
	md, err := matroska.ReadFile(audio)
	if err != nil {
		t.Fatal(err)
	}
	if len(md.Attachments) != 1 || md.Attachments[0].Filename != "cover.jpg" {
		t.Fatalf("attachments = %+v", md.Attachments)
	}
}

func TestImportOnlyASF(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.wma")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.wma"), audio)
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
	doc, err := asf.Read(audio)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(read.FieldValue(doc.Fields, "TITLE", "title"), "Tag title") {
		t.Fatalf("title = %v", doc.Fields)
	}
	if read.FieldValue(doc.Fields, "RATING", "rating") != "" {
		t.Fatalf("rating = %v", doc.Fields)
	}
	if read.FieldValue(doc.Fields, "TRACKNUMBER", "tracknumber") != "1/10" {
		t.Fatalf("track = %v", doc.Fields)
	}
}

func TestCustomFieldWarnings(t *testing.T) {
	cfg := config.Default()
	warnings := customFieldWarnings("track.mp3", map[string]string{"Titel": "typo"}, cfg, []string{"Titel"})
	if len(warnings) != 1 || !strings.Contains(warnings[0], `custom field "Titel" written as ID3 TXXX`) {
		t.Fatalf("warnings = %v", warnings)
	}
	if got := customFieldWarnings("track.mp3", map[string]string{"Rating": "5"}, cfg, []string{"Rating"}); len(got) != 0 {
		t.Fatalf("Rating warning = %v", got)
	}
	if got := customFieldWarnings("track.mp4", map[string]string{"Titel": "typo"}, cfg, []string{"Titel"}); len(got) != 0 {
		t.Fatalf("MP4 warning = %v", got)
	}
}
