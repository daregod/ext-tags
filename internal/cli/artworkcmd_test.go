package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/registry"
)

func TestArtworkListAndExtract(t *testing.T) {
	dir := t.TempDir()
	audio := filepath.Join(dir, "track.flac")
	copyFixture(t, filepath.Join("..", "..", "testdata", "formats", "audio.flac"), audio)
	a := &app{opts: &options{quiet: true}, registry: registry.NewDefault()}
	out := captureStdout(t, func() {
		if err := a.runArtworkList([]string{audio}); err != nil {
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
