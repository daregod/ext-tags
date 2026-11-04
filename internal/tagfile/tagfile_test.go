package tagfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/ape"
	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/extractor"
)

func newConfig(policy string) *config.Config {
	cfg := config.Default()
	cfg.Files.OnExisting = policy
	return cfg
}

func newAudio(t *testing.T, dir string) string {
	t.Helper()
	audio := filepath.Join(dir, "track.flac")
	if err := os.WriteFile(audio, []byte("audio"), 0644); err != nil {
		t.Fatal(err)
	}
	return audio
}

func writeTag(t *testing.T, path string, fields map[string]string) {
	t.Helper()
	tag := ape.NewTag()
	for key, value := range fields {
		tag.AddText(key, value)
	}
	data, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func readTag(t *testing.T, path string) *ape.ParsedTag {
	t.Helper()
	tag, err := ape.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return tag
}

func TestWriteCreatesFile(t *testing.T) {
	audio := newAudio(t, t.TempDir())

	res, err := Write(audio, map[string]string{"Title": "Song"}, nil, newConfig(config.OnExistingBackup), false)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Written || res.Items != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}
	if _, err := os.Stat(audio + ".tag"); err != nil {
		t.Fatalf("tag file missing: %v", err)
	}
}

func TestWriteBackup(t *testing.T) {
	audio := newAudio(t, t.TempDir())
	cfg := newConfig(config.OnExistingBackup)

	if _, err := Write(audio, map[string]string{"Title": "One"}, nil, cfg, false); err != nil {
		t.Fatal(err)
	}
	res, err := Write(audio, map[string]string{"Title": "Two"}, nil, cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Backup == "" {
		t.Fatal("expected backup path")
	}
	if _, err := os.Stat(audio + ".tag.bak"); err != nil {
		t.Fatalf("backup missing: %v", err)
	}
}

func TestWriteSkip(t *testing.T) {
	audio := newAudio(t, t.TempDir())
	cfg := newConfig(config.OnExistingSkip)
	if _, err := Write(audio, map[string]string{"Title": "One"}, nil, cfg, false); err != nil {
		t.Fatal(err)
	}
	res, err := Write(audio, map[string]string{"Title": "Two"}, nil, cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Skipped || res.Reason != "already exists" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestWriteErrorPolicy(t *testing.T) {
	audio := newAudio(t, t.TempDir())
	cfg := newConfig(config.OnExistingError)
	if _, err := Write(audio, map[string]string{"Title": "One"}, nil, cfg, false); err != nil {
		t.Fatal(err)
	}
	if _, err := Write(audio, map[string]string{"Title": "Two"}, nil, cfg, false); err == nil {
		t.Fatal("expected error for existing tag file")
	}
}

func TestWriteDryRun(t *testing.T) {
	audio := newAudio(t, t.TempDir())
	res, err := Write(audio, map[string]string{"Title": "One"}, nil, newConfig(config.OnExistingBackup), true)
	if err != nil {
		t.Fatal(err)
	}
	if res.Written {
		t.Fatalf("dry run must not write: %+v", res)
	}
	if _, err := os.Stat(audio + ".tag"); !os.IsNotExist(err) {
		t.Fatal("tag file must not exist after dry run")
	}
}

func TestBuildCover(t *testing.T) {
	cfg := config.Default()
	cover := &extractor.Cover{Filename: "cover.jpg", MIME: "image/jpeg", Data: []byte{0xFF, 0xD8}}
	data, items, _, err := Build(map[string]string{"Title": "Song"}, cover, nil, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if items != 2 || len(data) == 0 {
		t.Fatalf("items = %d, len = %d", items, len(data))
	}
}

func TestMergePreservesFields(t *testing.T) {
	audio := newAudio(t, t.TempDir())
	writeTag(t, audio+".tag", map[string]string{"Title": "Old", "Rating": "3"})

	cfg := newConfig(config.OnExistingBackup)
	if _, err := Write(audio, map[string]string{"Title": "New"}, nil, cfg, false); err != nil {
		t.Fatal(err)
	}

	gen := readTag(t, audio+".tag")
	if item, ok := gen.Get("Title"); !ok || string(item.Value) != "New" {
		t.Fatalf("Title = %q", item.Value)
	}
	if item, ok := gen.Get("Rating"); !ok || string(item.Value) != "3" {
		t.Fatalf("Rating must be preserved, got %q/%v", item.Value, ok)
	}

	backup := readTag(t, audio+".tag.bak")
	if item, ok := backup.Get("Title"); !ok || string(item.Value) != "Old" {
		t.Fatalf("backup Title = %q", item.Value)
	}
}

func TestMergeDisabledDropsFields(t *testing.T) {
	audio := newAudio(t, t.TempDir())
	writeTag(t, audio+".tag", map[string]string{"Title": "Old", "Rating": "3"})

	cfg := newConfig(config.OnExistingBackup)
	cfg.Files.MergeExisting = false
	if _, err := Write(audio, map[string]string{"Title": "New"}, nil, cfg, false); err != nil {
		t.Fatal(err)
	}

	gen := readTag(t, audio+".tag")
	if _, ok := gen.Get("Rating"); ok {
		t.Fatal("Rating must be dropped when merge is disabled")
	}
}

func TestMergeKeepsCover(t *testing.T) {
	audio := newAudio(t, t.TempDir())
	old := ape.NewTag()
	old.AddText("Title", "Old")
	old.AddCoverArt("Cover Art (Front)", "old.jpg", []byte{0xFF, 0xD8, 0x01})
	data, err := old.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(audio+".tag", data, 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := Write(audio, map[string]string{"Title": "New"}, nil, newConfig(config.OnExistingBackup), false); err != nil {
		t.Fatal(err)
	}
	gen := readTag(t, audio+".tag")
	cover, ok := gen.Get("Cover Art (Front)")
	if !ok || !hasSuffix(cover.Value, []byte{0xFF, 0xD8, 0x01}) {
		t.Fatal("existing cover must be preserved when the audio has none")
	}
}

func TestMergeSkipLeavesFileUntouched(t *testing.T) {
	audio := newAudio(t, t.TempDir())
	writeTag(t, audio+".tag", map[string]string{"Title": "Old", "Rating": "3"})

	cfg := newConfig(config.OnExistingSkip)
	res, err := Write(audio, map[string]string{"Title": "New"}, nil, cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Skipped {
		t.Fatal("expected skip")
	}
	gen := readTag(t, audio+".tag")
	if item, _ := gen.Get("Title"); string(item.Value) != "Old" {
		t.Fatalf("skip must not rewrite the tag, Title = %q", item.Value)
	}
}

func TestMergeDryRun(t *testing.T) {
	audio := newAudio(t, t.TempDir())
	writeTag(t, audio+".tag", map[string]string{"Title": "Old", "Rating": "3"})

	res, err := Write(audio, map[string]string{"Title": "New"}, nil, newConfig(config.OnExistingBackup), true)
	if err != nil {
		t.Fatal(err)
	}
	if res.Written {
		t.Fatalf("dry run must not write: %+v", res)
	}
	if _, err := os.Stat(audio + ".tag.bak"); !os.IsNotExist(err) {
		t.Fatal("dry run must not create a backup")
	}
	gen := readTag(t, audio+".tag")
	if item, _ := gen.Get("Title"); string(item.Value) != "Old" {
		t.Fatalf("dry run must not rewrite the tag, Title = %q", item.Value)
	}
}

func hasSuffix(value, suffix []byte) bool {
	if len(value) < len(suffix) {
		return false
	}
	return string(value[len(value)-len(suffix):]) == string(suffix)
}
