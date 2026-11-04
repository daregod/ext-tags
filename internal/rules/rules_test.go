package rules

import (
	"testing"
	"time"

	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/metadata"
)

func build(t *testing.T, cfg *config.Config, fields map[string]string, vars map[string]string) map[string]string {
	t.Helper()
	engine, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res, err := engine.Build(&metadata.Document{Fields: fields, Vars: vars}, Meta{
		Path: "C:/music/track.flac", Filename: "track", Dir: "C:/music", Ext: ".flac",
		Now: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return res
}

func TestAliasesAndTrackTemplate(t *testing.T) {
	fields := build(t, config.Default(), map[string]string{
		"TITLE":       "Song",
		"artist":      "Artist",
		"vendor":      "Lavf",
		"TRACKNUMBER": "2",
		"TRACKTOTAL":  "10",
	}, map[string]string{"tracknumber": "2", "totaltracks": "10"})

	if fields["Title"] != "Song" {
		t.Fatalf("Title = %q", fields["Title"])
	}
	if fields["Artist"] != "Artist" {
		t.Fatalf("Artist = %q", fields["Artist"])
	}
	if fields["Track"] != "2/10" {
		t.Fatalf("Track = %q, want 2/10", fields["Track"])
	}
	if _, ok := fields["vendor"]; ok {
		t.Fatal("vendor must be ignored")
	}
	if _, ok := fields["TRACKNUMBER"]; ok {
		t.Fatal("raw track number must be ignored")
	}
}

func TestMP4Atoms(t *testing.T) {
	fields := build(t, config.Default(), map[string]string{
		"©nam": "Song",
		"©ART": "Artist",
		"aART": "Album Artist",
		"cpil": "0",
	}, nil)

	if fields["Title"] != "Song" || fields["Artist"] != "Artist" {
		t.Fatalf("unexpected fields: %v", fields)
	}
	if fields["Album artist"] != "Album Artist" {
		t.Fatalf("Album artist = %q", fields["Album artist"])
	}
	if fields["Itunescompilation"] != "0" {
		t.Fatalf("Itunescompilation = %q", fields["Itunescompilation"])
	}
	if len(fields) != 4 {
		t.Fatalf("unexpected extra fields: %v", fields)
	}
}

func TestCopyRenameDefaultRemove(t *testing.T) {
	cfg := config.Default()
	cfg.Fields.Rules = []config.Rule{
		{Copy: &config.CopyRule{From: "Label", To: "Publisher"}},
		{Rename: &config.RenameRule{From: "Comment", To: "Note"}},
		{Default: &config.DefaultRule{Field: "Album artist", From: "Artist"}},
		{Remove: []string{"Www"}},
	}
	fields := build(t, cfg, map[string]string{
		"LABEL":   "Label Co",
		"COMMENT": "hello",
		"ARTIST":  "Artist",
		"WWW":     "https://example.com",
	}, nil)

	if fields["Publisher"] != "Label Co" {
		t.Fatalf("Publisher = %q", fields["Publisher"])
	}
	if fields["Note"] != "hello" {
		t.Fatalf("Note = %q", fields["Note"])
	}
	if _, ok := fields["Comment"]; ok {
		t.Fatal("Comment must be renamed")
	}
	if fields["Album artist"] != "Artist" {
		t.Fatalf("Album artist = %q", fields["Album artist"])
	}
	if _, ok := fields["Www"]; ok {
		t.Fatal("Www must be removed")
	}
}

func TestIncludeExclude(t *testing.T) {
	cfg := config.Default()
	cfg.Fields.Include = []string{"Title", "Artist"}
	fields := build(t, cfg, map[string]string{
		"TITLE": "Song", "ARTIST": "A", "ALBUM": "B",
	}, nil)
	if len(fields) != 2 {
		t.Fatalf("include must keep two fields: %v", fields)
	}

	cfg = config.Default()
	cfg.Fields.Exclude = []string{"album"}
	fields = build(t, cfg, map[string]string{
		"TITLE": "Song", "ALBUM": "B",
	}, nil)
	if _, ok := fields["Album"]; ok {
		t.Fatal("Album must be excluded")
	}
}

func TestTemplateFuncsAndWhen(t *testing.T) {
	cfg := config.Default()
	cfg.Fields.Rules = []config.Rule{
		{Set: "Sort", Template: `{{ lower (replace " " "_" .title) }}`},
		{Set: "OnlyWithGenre", Template: "yes", When: "{{ .genre }}"},
		{Set: "Never", Template: "x", When: "{{ .missing }}"},
	}
	fields := build(t, cfg, map[string]string{"TITLE": "My Song"}, nil)
	if fields["Sort"] != "my_song" {
		t.Fatalf("Sort = %q", fields["Sort"])
	}
	if _, ok := fields["OnlyWithGenre"]; ok {
		t.Fatal("rule with empty when must not run")
	}
	if _, ok := fields["Never"]; ok {
		t.Fatal("rule with missing when must not run")
	}
}

func TestRulesSeePreviousResults(t *testing.T) {
	cfg := config.Default()
	cfg.Fields.Rules = []config.Rule{
		{Set: "Title", Template: `{{ upper .title }}`},
		{Set: "Sort", Template: `{{ .title }}`},
	}
	fields := build(t, cfg, map[string]string{"TITLE": "song"}, nil)
	if fields["Sort"] != "SONG" {
		t.Fatalf("Sort = %q, want SONG", fields["Sort"])
	}
}

func TestUnknownID3FrameSkipped(t *testing.T) {
	fields := build(t, config.Default(), map[string]string{"TSOA": "x", "MYKEY": "value"}, nil)
	if _, ok := fields["TSOA"]; ok {
		t.Fatal("unknown ID3v2 frame must be skipped")
	}
	if fields["MYKEY"] != "value" {
		t.Fatalf("custom field = %q", fields["MYKEY"])
	}
}

func TestInvalidTemplate(t *testing.T) {
	cfg := config.Default()
	cfg.Fields.Rules = []config.Rule{{Set: "X", Template: "{{ .title"}}
	if _, err := New(cfg); err == nil {
		t.Fatal("expected template parse error")
	}
}
func TestYearAndFormatTrack(t *testing.T) {
	cfg := config.Default()
	cfg.Fields.Rules = []config.Rule{
		{Set: "Year", Template: `{{ year .releasetime }}`},
		{Set: "Track", Template: `{{ formatTrack .track }}`},
	}
	fields := build(t, cfg, map[string]string{
		"Releasetime": "2016-05-16",
		"Track":       "1/12",
	}, nil)
	if fields["Year"] != "2016" {
		t.Fatalf("Year = %q", fields["Year"])
	}
	if fields["Track"] != "01/12" {
		t.Fatalf("Track = %q", fields["Track"])
	}
	if got := formatTrack("1/8"); got != "1/8" {
		t.Fatalf("formatTrack(1/8) = %q", got)
	}
	if got := formatTrack("7/12"); got != "07/12" {
		t.Fatalf("formatTrack(7/12) = %q", got)
	}
	if got := formatTrack("7"); got != "7" {
		t.Fatalf("formatTrack(7) = %q", got)
	}
	if got := yearOf("2025"); got != "2025" {
		t.Fatalf("yearOf = %q", got)
	}
}
