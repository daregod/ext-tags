package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/formats/ape"
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/registry"
)

func TestArtworkListAndExtract(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), audio)
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	out := captureStdout(t, func() {
		if err := a.runArtworkList([]string{audio}, false, false); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "front") || !strings.Contains(out, "cover.jpg") {
		t.Fatalf("list = %q", out)
	}

	outPath := filepath.Join(dir, "front.jpg")
	if err := a.runArtworkExtract(audio, &artworkExtractOptions{slot: "front", output: outPath}); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(outPath); err != nil || info.Size() == 0 {
		t.Fatalf("extracted file: %v %v", info, err)
	}
	if err := a.runArtworkExtract(audio, &artworkExtractOptions{slot: "front", output: outPath}); err == nil {
		t.Fatal("existing output must fail without --force")
	}
	a = &app{opts: &options{quiet: true, json: true}, registry: registry.NewDefault()}
	captureStdout(t, func() {
		if err := a.runArtworkExtract(audio, &artworkExtractOptions{slot: "front", output: outPath, force: true}); err != nil {
			t.Fatal(err)
		}
	})
}

func TestArtworkIndexListAndExtract(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), audio)
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	out := captureStdout(t, func() {
		if err := a.runArtworkList([]string{audio}, false, true); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "container:0") {
		t.Fatalf("index list = %q", out)
	}
	outPath := filepath.Join(dir, "cover.jpg")
	if err := a.runArtworkExtract(audio, &artworkExtractOptions{slot: "front", output: outPath, index: "container:0"}); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(outPath); err != nil || info.Size() == 0 {
		t.Fatalf("extracted index: %v %v", info, err)
	}
}

func TestArtworkIndexTagSource(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), audio)
	tag := ape.NewTag()
	tag.AddCoverArt("Cover Art (Front)", "front.png", []byte{1, 2, 3})
	tag.AddCoverArt("Cover Art (Back)", "back.png", []byte{4, 5, 6})
	data, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(audio+".tag", data, 0644); err != nil {
		t.Fatal(err)
	}
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	out := captureStdout(t, func() {
		if err := a.runArtworkList([]string{audio}, false, true); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "tag:0") || !strings.Contains(out, "tag:1") {
		t.Fatalf("index list = %q", out)
	}
	outPath := filepath.Join(dir, "tag-cover.png")
	if err := a.runArtworkExtract(audio, &artworkExtractOptions{slot: "front", output: outPath, index: "tag:1"}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string([]byte{4, 5, 6}) {
		t.Fatalf("tag:1 = %v", got)
	}
}

func TestArtworkIndexErrors(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), audio)
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	if err := a.runArtworkExtract(audio, &artworkExtractOptions{slot: "front", index: "container:99", output: filepath.Join(dir, "x")}); err == nil {
		t.Fatal("expected missing index")
	}
	if err := a.runArtworkExtract(audio, &artworkExtractOptions{slot: "back", index: "container:0", output: filepath.Join(dir, "x")}); err == nil {
		t.Fatal("expected slot conflict")
	}
	if err := a.runArtworkExtract(audio, &artworkExtractOptions{slot: "front", index: "bad", output: filepath.Join(dir, "x")}); err == nil {
		t.Fatal("expected bad key")
	}
	if err := a.runArtworkExtract(audio, &artworkExtractOptions{slot: "front"}); err == nil {
		t.Fatal("expected output required")
	}
	if err := a.runArtworkExtract(filepath.Join(dir, "missing.flac"), &artworkExtractOptions{slot: "front", output: filepath.Join(dir, "x")}); err == nil {
		t.Fatal("expected read error")
	}
	if err := a.runArtworkList([]string{filepath.Join(dir, "missing.flac")}, false, true); err != nil {
		t.Fatal(err)
	}
}

func TestArtworkWarningsAndSlotValidation(t *testing.T) {
	a := &app{opts: &options{quiet: true}}
	a.warnArtworkSkips("track.flac", []artworkSkip{{Slot: "front", Cover: &metadata.Cover{Filename: "big.jpg", Data: make([]byte, 10)}, Hard: 1}}, "--all")
	if a.stats.warnings != 1 {
		t.Fatalf("warnings = %d", a.stats.warnings)
	}
	if !validArtworkSlot("front") || validArtworkSlot("bad") {
		t.Fatal("slot validation")
	}
}

func TestLoadArtworkErrors(t *testing.T) {
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	cfg := config.Default()
	if _, _, err := a.loadArtworkWithOptions("track.unknown", cfg, false); err == nil {
		t.Fatal("expected unsupported")
	}
	if _, _, err := a.loadArtworkWithOptions(filepath.Join(t.TempDir(), "missing.flac"), cfg, false); err == nil {
		t.Fatal("expected read error")
	}
	empty := filepath.Join(t.TempDir(), "empty.wav")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio-empty.wav"), empty)
	if covers, skips, err := a.loadArtworkWithOptions(empty, cfg, false); err != nil || len(covers) != 0 || len(skips) != 0 {
		t.Fatalf("empty = %+v %+v %v", covers, skips, err)
	}
}
