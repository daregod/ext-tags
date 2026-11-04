package riff

import (
	"encoding/binary"
	"path/filepath"
	"testing"
)

func TestRF64DataSizeAndPatch(t *testing.T) {
	payload := make([]byte, 28)
	binary.LittleEndian.PutUint64(payload[0:8], 1000)
	binary.LittleEndian.PutUint64(payload[8:16], 123)
	data := append([]byte("RF64"), []byte{0xFF, 0xFF, 0xFF, 0xFF}...)
	data = append(data, []byte("WAVE")...)
	data = append(data, chunk("ds64", payload)...)
	if got := rf64DataSize(data); got != 123 {
		t.Fatalf("dataSize = %d", got)
	}
	patchDS64(data)
	if got := binary.LittleEndian.Uint64(data[12+8 : 12+16]); got != uint64(len(data)-8) {
		t.Fatalf("riffSize = %d", got)
	}
}

func TestReadSpecificFixtures(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "testdata", "formats")
	if _, err := ReadRF64(filepath.Join(dir, "audio.rf64")); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadWave64(filepath.Join(dir, "audio.w64")); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadAIFF(filepath.Join(dir, "audio.aiff")); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadFile(filepath.Join(dir, "audio.wav")); err != nil {
		t.Fatal(err)
	}
}
