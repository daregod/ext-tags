package extractor_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"external-tags/internal/config"
	"external-tags/internal/extractor"
	"external-tags/internal/rules"
)

type fixtureCase struct {
	name   string
	title  string
	artist string
	cover  bool
}

func TestFixtures(t *testing.T) {
	cfg := config.Default()
	engine, err := rules.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	cases := []fixtureCase{
		{"audio.mp3", "Fixture Title", "Fixture Artist", true},
		{"audio.mp2", "Fixture Title", "Fixture Artist", true},
		{"audio.aac", "Fixture Title", "Fixture Artist", true},
		{"audio.flac", "Fixture Title", "Fixture Artist", true},
		{"audio.ogg", "Fixture Title", "Fixture Artist", false},
		{"audio.opus", "Fixture Title", "Fixture Artist", false},
		{"audio-aac.m4a", "Fixture Title", "Fixture Artist", true},
		{"audio-alac.m4a", "Fixture Title", "Fixture Artist", true},
		{"audio.wav", "Fixture Title", "Fixture Artist", false},
		{"audio.aiff", "Fixture Title", "Fixture Artist", true},
		{"audio.wv", "Fixture Title", "Fixture Artist", false},
		{"audio.wma", "Fixture Title", "Fixture Artist", false},
		{"audio.mkv", "Fixture Title", "Fixture Artist", true},
		{"audio-attachments.mkv", "Fixture Title", "Fixture Artist", true},
		{"audio.webm", "Fixture Title", "Fixture Artist", false},
		{"video.wmv", "Fixture Title", "Fixture Artist", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := filepath.Join("..", "..", "testdata", "formats", c.name)
			res, err := extractor.ExtractWithOptions(path, extractor.Options{CoverNames: cfg.Cover.Names})
			if err != nil {
				t.Fatalf("extract: %v", err)
			}
			if c.cover && res.Cover == nil {
				t.Fatal("cover art expected")
			}
			fields, err := engine.Build(res, rules.Meta{
				Path: path, Filename: c.name, Dir: "testdata", Ext: filepath.Ext(c.name),
				Now: time.Unix(0, 0),
			})
			if err != nil {
				t.Fatalf("rules: %v", err)
			}
			if fields["Title"] != c.title {
				t.Fatalf("Title = %q, want %q (fields %v)", fields["Title"], c.title, fields)
			}
			if fields["Artist"] != c.artist {
				t.Fatalf("Artist = %q, want %q", fields["Artist"], c.artist)
			}
			if fields["Track"] != "3" {
				t.Fatalf("Track = %q, want 3", fields["Track"])
			}
		})
	}
}

func TestCoverAttachmentPriority(t *testing.T) {
	formats := filepath.Join("..", "..", "testdata", "formats")
	folderJPEG, err := os.ReadFile(filepath.Join(formats, "folder.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	posterJPEG, err := os.ReadFile(filepath.Join(formats, "poster.jpg"))
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(formats, "audio-attachments.mkv")
	res, err := extractor.Extract(path)
	if err != nil {
		t.Fatal(err)
	}
	if res.Cover == nil || !bytes.Equal(res.Cover.Data, folderJPEG) {
		t.Fatal("folder.jpg must win over poster.jpg with default names")
	}

	res, err = extractor.ExtractWithOptions(path, extractor.Options{CoverNames: []string{"poster"}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Cover == nil || !bytes.Equal(res.Cover.Data, posterJPEG) {
		t.Fatal("poster must win when configured first")
	}
}
