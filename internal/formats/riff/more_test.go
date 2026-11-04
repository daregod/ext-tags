package riff

import (
	"encoding/binary"
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
func TestMalformedContainerChunks(t *testing.T) {
	dir := t.TempDir()
	rf64 := filepath.Join(dir, "bad.rf64")
	data := append([]byte("RF64"), []byte{0xFF, 0xFF, 0xFF, 0xFF}...)
	data = append(data, []byte("WAVE")...)
	data = append(data, []byte{0xFF, 0xFF, 0xFF, 0xFF, 1, 2, 3, 4}...)
	if err := os.WriteFile(rf64, data, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadRF64(rf64); err != nil {
		t.Fatal(err)
	}
	w64 := filepath.Join(dir, "bad.w64")
	head := make([]byte, 40)
	copy(head[0:4], "riff")
	copy(head[24:28], "wave")
	chunk := make([]byte, 24)
	copy(chunk[0:4], "data")
	binary.LittleEndian.PutUint64(chunk[16:24], 1)
	if err := os.WriteFile(w64, append(head, chunk...), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadWave64(w64); err != nil {
		t.Fatal(err)
	}
}

func TestDS64Short(t *testing.T) {
	patchDS64(nil)
	rf64DataSize(nil)
}
