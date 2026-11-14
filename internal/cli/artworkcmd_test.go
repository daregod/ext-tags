package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/ape"
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
