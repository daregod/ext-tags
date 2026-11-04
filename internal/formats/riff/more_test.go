package riff

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/metadata"
)

func TestReadSpecificMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "missing.bin")
	for _, fn := range []func(string) (*Metadata, error){ReadRF64, ReadWave64, ReadAIFF} {
		if _, err := fn(path); err == nil {
			t.Fatal("missing file must fail")
		}
	}
	if _, err := Read(path); err == nil {
		t.Fatal("Read missing must fail")
	}
}

func TestIsPrintableFourCC(t *testing.T) {
	for _, id := range []string{"INAM", "IART", "ICMT"} {
		if !isPrintableFourCC(id) {
			t.Errorf("%s must be printable", id)
		}
	}
	for _, id := range []string{"\x00NAM", "IN\x01M"} {
		if isPrintableFourCC(id) {
			t.Errorf("%q must not be printable", id)
		}
	}
}

func TestWriteWithoutCover(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "track.wav")
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "formats", "audio.wav"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	if err := Write(path, metadata.New()); err != nil {
		t.Fatal(err)
	}
}
func TestBuildID3CoverFallback(t *testing.T) {
	doc := metadata.New()
	doc.Cover = &metadata.Cover{Filename: "cover.jpg", MIME: "image/jpeg", Data: []byte{1}}
	tag, err := buildID3(doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(tag) == 0 {
		t.Fatal("empty tag")
	}
}

func TestReadAtLimits(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "readat")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if readAt(f, 0, 0) != nil || readAt(f, 0, maxTagSize+1) != nil {
		t.Fatal("invalid sizes must return nil")
	}
}
