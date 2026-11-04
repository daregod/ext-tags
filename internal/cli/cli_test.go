package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/metadata"
)

func TestExpandPathsRecursive(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{filepath.Join(dir, "a.flac"), filepath.Join(sub, "b.flac")} {
		if err := os.WriteFile(name, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	flat, err := expandPaths([]string{dir}, false, []string{".flac"})
	if err != nil {
		t.Fatal(err)
	}
	if len(flat) != 1 {
		t.Fatalf("flat = %v, want one file", flat)
	}

	deep, err := expandPaths([]string{dir}, true, []string{".flac"})
	if err != nil {
		t.Fatal(err)
	}
	if len(deep) != 2 {
		t.Fatalf("deep = %v, want two files", deep)
	}
}

func TestExpandPathsGlob(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.flac", "b.flac", "c.mp3"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	files, err := expandPaths([]string{filepath.Join(dir, "*.flac")}, false, []string{".flac"})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("glob = %v, want two files", files)
	}
}

func TestApplyOverrides(t *testing.T) {
	a := &app{opts: &options{recursive: true, onExisting: "skip", cover: false}}
	cfg := config.Default()
	changed := func(name string) bool {
		return name == "recursive" || name == "on-existing" || name == "cover"
	}
	if err := a.applyOverrides(cfg, changed); err != nil {
		t.Fatal(err)
	}

	if !cfg.Files.Recursive {
		t.Fatal("recursive override not applied")
	}
	if cfg.Files.OnExisting != "skip" {
		t.Fatalf("on_existing = %q", cfg.Files.OnExisting)
	}
	if cfg.Cover.Enabled {
		t.Fatal("cover override not applied")
	}
}

func TestConfigInitAndValidate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ext-tags.yaml")

	a := &app{opts: &options{}}
	root := a.rootCmd()
	root.SetArgs([]string{"config", "init", path})
	if err := root.Execute(); err != nil {
		t.Fatalf("config init: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file missing: %v", err)
	}

	a = &app{opts: &options{configPath: path}}
	root = a.rootCmd()
	root.SetArgs([]string{"config", "validate"})
	if err := root.Execute(); err != nil {
		t.Fatalf("config validate: %v", err)
	}
}

func TestConfigInitRefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ext-tags.yaml")
	if err := os.WriteFile(path, []byte("version: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{}}
	root := a.rootCmd()
	root.SetArgs([]string{"config", "init", path})
	if err := root.Execute(); err == nil {
		t.Fatal("expected refusal to overwrite")
	}
}

func TestConfigInitMinimalAndFull(t *testing.T) {
	dir := t.TempDir()
	minimalPath := filepath.Join(dir, "minimal.yaml")
	fullPath := filepath.Join(dir, "full.yaml")

	a := &app{opts: &options{}}
	root := a.rootCmd()
	root.SetArgs([]string{"config", "init", minimalPath})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	a = &app{opts: &options{}}
	root = a.rootCmd()
	root.SetArgs([]string{"config", "init", "--full", fullPath})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	minimal, err := os.ReadFile(minimalPath)
	if err != nil {
		t.Fatal(err)
	}
	full, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(minimal) >= len(full) {
		t.Fatalf("minimal config (%d B) must be smaller than full (%d B)", len(minimal), len(full))
	}
	if !strings.Contains(string(minimal), "minimal") {
		t.Fatal("minimal config must have a minimal header")
	}
	if !strings.Contains(string(full), "aliases:") {
		t.Fatal("full config must contain aliases")
	}
}

func TestIsOwnOutput(t *testing.T) {
	a := &app{opts: &options{}}
	cfg := config.Default()
	cases := map[string]bool{
		"track.flac.tag":     true,
		"track.flac.tag.bak": true,
		"TRACK.FLAC.TAG":     true,
		"track.flac":         false,
		"track.tagged.flac":  false,
	}
	for path, want := range cases {
		if got := a.isOwnOutput(path, cfg); got != want {
			t.Fatalf("isOwnOutput(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestApplySidecarCover(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "cover.jpg"), []byte{0xFF, 0xD8}, 0644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	a := &app{opts: &options{}}

	md := &metadata.Document{Fields: map[string]string{}, Vars: map[string]string{}}
	a.applySidecarArtwork(md, filepath.Join(dir, "track.flac"), cfg)
	if md.Cover == nil || md.Cover.Filename != "cover.jpg" || md.Cover.MIME != "image/jpeg" {
		t.Fatalf("cover = %+v", md.Cover)
	}

	embedded := &metadata.Document{Cover: &metadata.Cover{Filename: "embedded.png", MIME: "image/png"}}
	a.applySidecarArtwork(embedded, filepath.Join(dir, "track.flac"), cfg)
	if embedded.Cover.Filename != "embedded.png" {
		t.Fatal("embedded cover must win")
	}

	cfg.Cover.Directory = false
	md = &metadata.Document{Fields: map[string]string{}, Vars: map[string]string{}}
	a.applySidecarArtwork(md, filepath.Join(dir, "track.flac"), cfg)
	if md.Cover != nil {
		t.Fatal("sidecar must be disabled")
	}
}
func TestApplySidecarAllArtwork(t *testing.T) {
	dir := t.TempDir()
	for name, data := range map[string][]byte{
		"cover.jpg":  {0xFF, 0xD8},
		"back.png":   {0x89, 0x50},
		"artist.jpg": {0xFF, 0xD8},
		"disc.jpg":   {0xFF, 0xD8},
		"icon.gif":   {0x47, 0x49},
	} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	cfg := config.Default()
	a := &app{opts: &options{}}
	md := &metadata.Document{Fields: map[string]string{}, Vars: map[string]string{}}
	a.applySidecarArtwork(md, filepath.Join(dir, "track.flac"), cfg)
	for _, kind := range []string{"front", "back", "artist", "disc", "icon"} {
		if md.Artwork[kind] == nil {
			t.Fatalf("missing %s: %v", kind, md.Artwork)
		}
	}
	if md.Cover == nil || md.Cover.Filename != "cover.jpg" {
		t.Fatalf("front = %+v", md.Cover)
	}
	if got := md.Artwork["back"].Filename; got != "back.png" {
		t.Fatalf("back = %q", got)
	}
	if got := md.Artwork["icon"].Filename; got != "icon.gif" {
		t.Fatalf("icon = %q", got)
	}
}
func TestSniffContainer(t *testing.T) {
	dir := t.TempDir()
	cases := map[string]string{"x.flac": "OggS", "y.flac": "fLaC", "z.mp3": "ID3\x04"}
	for name, magic := range cases {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(magic), 0644); err != nil {
			t.Fatal(err)
		}
		want := ""
		switch magic[:4] {
		case "OggS":
			want = "ogg"
		case "fLaC":
			want = "flac"
		}
		if got := sniffContainer(path); got != want {
			t.Fatalf("%s: got %q want %q", name, got, want)
		}
	}
}
