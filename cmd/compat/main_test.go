package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/tagfile"
)

func TestRunDirectories(t *testing.T) {
	if code := run(filepath.Join(t.TempDir(), "missing"), "", false); code != 1 {
		t.Fatalf("missing dir code = %d", code)
	}
	if code := run(t.TempDir(), "", false); code != 0 {
		t.Fatalf("empty dir code = %d", code)
	}

	dir := t.TempDir()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "formats", "audio.flac"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "track.flac"), data, 0644); err != nil {
		t.Fatal(err)
	}
	refDir := t.TempDir()
	if code := run(dir, refDir, false); code != 1 {
		t.Fatalf("missing references code = %d", code)
	}
}

func TestDumpTag(t *testing.T) {
	data, _, _, err := tagfile.Build(map[string]string{"Title": "Song"}, nil, nil, config.Default())
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "track.flac.tag")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	dumpTag(path)
}

func TestCompareAndHelpers(t *testing.T) {
	ref := ape.NewTag()
	ref.AddText("Title", " Song ")
	ref.AddBinary("Cover Art (Front)", []byte("cover.jpg\x00image"))
	gen := ape.NewTag()
	gen.AddText("Title", "Song")
	gen.AddBinary("Cover Art (Front)", []byte("other.jpg\x00image"))
	refData, err := ref.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	refParsed, err := ape.Parse(refData)
	if err != nil {
		t.Fatal(err)
	}
	genData, err := gen.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	genParsed, err := ape.Parse(genData)
	if err != nil {
		t.Fatal(err)
	}
	changed, cosmetic, refOnly, genOnly := compare(refParsed, genParsed)
	if len(changed) != 0 || len(refOnly) != 0 || len(genOnly) != 0 {
		t.Fatalf("changed=%v ref=%v gen=%v", changed, refOnly, genOnly)
	}
	if len(cosmetic) != 2 {
		t.Fatalf("cosmetic = %v", cosmetic)
	}
	if got := normalizeText("ReplayGain_Track_Gain", []byte(" -6.0 dB ")); got != "-6.0" {
		t.Fatalf("normalized = %q", got)
	}
	if got := keys(refParsed); len(got) != 2 {
		t.Fatalf("keys = %v", got)
	}
	if got := describe(refParsed, []string{"Title"}); got == "" {
		t.Fatal("describe empty")
	}
	if got := coverImage([]byte("name\x00image")); !bytes.Equal(got, []byte("image")) {
		t.Fatalf("cover = %q", got)
	}
	if got := index(refParsed); len(got) != 2 {
		t.Fatalf("index = %v", got)
	}
}
func TestRunWithReference(t *testing.T) {
	dir := t.TempDir()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "formats", "audio.flac"))
	if err != nil {
		t.Fatal(err)
	}
	audio := filepath.Join(dir, "track.flac")
	if err := os.WriteFile(audio, data, 0644); err != nil {
		t.Fatal(err)
	}
	ref := ape.NewTag()
	ref.AddText("Title", "wrong")
	tagData, err := ref.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(audio+".tag", tagData, 0644); err != nil {
		t.Fatal(err)
	}
	if code := run(dir, "", false); code != 1 {
		t.Fatalf("diff code = %d", code)
	}
}
