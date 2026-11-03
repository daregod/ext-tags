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
	name        string
	title       string
	artist      string
	albumArtist string
	track       string
	disc        string
	cover       bool
}

func TestFixtures(t *testing.T) {
	cfg := config.Default()
	engine, err := rules.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	cases := []fixtureCase{
		{name: "audio.mp3", title: "Fixture Title", artist: "Fixture Artist", track: "3", cover: true},
		{name: "audio.mp2", title: "Fixture Title", artist: "Fixture Artist", track: "3", cover: true},
		{name: "audio.aac", title: "Fixture Title", artist: "Fixture Artist", track: "3", cover: true},
		{name: "audio.flac", title: "Fixture Title", artist: "Fixture Artist", track: "3", cover: true},
		{name: "audio-unicode.flac", title: "Тестовый заголовок", artist: "Исполнитель", track: "3", cover: true},
		{name: "audio.ogg", title: "Fixture Title", artist: "Fixture Artist", track: "3"},
		{name: "audio.opus", title: "Fixture Title", artist: "Fixture Artist", track: "3"},
		{name: "audio-aac.m4a", title: "Fixture Title", artist: "Fixture Artist", track: "3", cover: true},
		{name: "audio-alac.m4a", title: "Fixture Title", artist: "Fixture Artist", track: "3", cover: true},
		{name: "audio.wav", title: "Fixture Title", artist: "Fixture Artist", track: "3"},
		{name: "audio.aiff", title: "Fixture Title", artist: "Fixture Artist", track: "3", cover: true},
		{name: "audio.wv", title: "Fixture Title", artist: "Fixture Artist", track: "3"},
		{name: "audio.wma", title: "Fixture Title", artist: "Fixture Artist", track: "3"},
		{name: "video.wmv", title: "Fixture Title", artist: "Fixture Artist", track: "3"},
		// Matroska tags written by ffmpeg are album-level: foobar2000 maps
		// ARTIST to Album artist and PART_NUMBER to Disc.
		{name: "audio.mkv", title: "Fixture Title", albumArtist: "Fixture Artist", disc: "3", cover: true},
		{name: "audio-attachments.mkv", title: "Fixture Title", albumArtist: "Fixture Artist", disc: "3", cover: true},
		{name: "audio-track.mkv", title: "Album Title", artist: "Track Artist", disc: "3"},
		{name: "audio-both.mkv", title: "Album Title", artist: "Track Artist", albumArtist: "Album Artist", disc: "3"},
		{name: "audio-cover.png.mkv", title: "Fixture Title", albumArtist: "Fixture Artist", disc: "3", cover: true},
		{name: "audio.webm", title: "Fixture Title", albumArtist: "Fixture Artist", disc: "3"},
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
			checks := map[string]string{
				"Title":        c.title,
				"Artist":       c.artist,
				"Album artist": c.albumArtist,
				"Track":        c.track,
				"Disc":         c.disc,
			}
			for field, want := range checks {
				if fields[field] != want {
					t.Fatalf("%s = %q, want %q (fields %v)", field, fields[field], want, fields)
				}
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

func TestEmptyWAV(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "formats", "audio-empty.wav")
	if _, err := extractor.Extract(path); err == nil {
		t.Fatal("expected ErrNoTags for a tagless WAV")
	}
}

func TestPNGCoverName(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "formats", "audio-cover.png.mkv")
	res, err := extractor.Extract(path)
	if err != nil {
		t.Fatal(err)
	}
	if res.Cover == nil || res.Cover.Filename != "cover.png" {
		t.Fatalf("cover = %+v", res.Cover)
	}
}
