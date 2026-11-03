package tagfile

import (
	"os"
	"path/filepath"
	"testing"

	"external-tags/internal/config"
	"external-tags/internal/extractor"
)

func newConfig(policy string) *config.Config {
	cfg := config.Default()
	cfg.Files.OnExisting = policy
	return cfg
}

func TestWriteCreatesFile(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	if err := os.WriteFile(audio, []byte("audio"), 0644); err != nil {
		t.Fatal(err)
	}

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
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	if err := os.WriteFile(audio, []byte("audio"), 0644); err != nil {
		t.Fatal(err)
	}
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
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	if err := os.WriteFile(audio, []byte("audio"), 0644); err != nil {
		t.Fatal(err)
	}
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
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	if err := os.WriteFile(audio, []byte("audio"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := newConfig(config.OnExistingError)
	if _, err := Write(audio, map[string]string{"Title": "One"}, nil, cfg, false); err != nil {
		t.Fatal(err)
	}
	if _, err := Write(audio, map[string]string{"Title": "Two"}, nil, cfg, false); err == nil {
		t.Fatal("expected error for existing tag file")
	}
}

func TestWriteDryRun(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	if err := os.WriteFile(audio, []byte("audio"), 0644); err != nil {
		t.Fatal(err)
	}
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
	data, items, err := Build(map[string]string{"Title": "Song"}, cover, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if items != 2 || len(data) == 0 {
		t.Fatalf("items = %d, len = %d", items, len(data))
	}
}
