package registry_test

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/ape"
)

func wavChunk(id string, data []byte) []byte {
	out := []byte(id)
	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(data)))
	out = append(out, size...)
	out = append(out, data...)
	if len(data)%2 == 1 {
		out = append(out, 0)
	}
	return out
}

func syncSafe(n int) []byte {
	return []byte{byte(n >> 21 & 0x7F), byte(n >> 14 & 0x7F), byte(n >> 7 & 0x7F), byte(n & 0x7F)}
}

func id3Title(title string) []byte {
	payload := append([]byte{0}, []byte(title)...)
	frame := []byte("TIT2")
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(payload)))
	frame = append(frame, size...)
	frame = append(frame, 0, 0)
	frame = append(frame, payload...)

	out := []byte("ID3\x03\x00\x00")
	out = append(out, syncSafe(len(frame))...)
	return append(out, frame...)
}

func TestExtractWAV(t *testing.T) {
	info := []byte("INFO")
	info = append(info, wavChunk("INAM", []byte("Info Title\x00"))...)
	info = append(info, wavChunk("IART", []byte("Info Artist\x00"))...)

	body := wavChunk("fmt ", make([]byte, 16))
	body = append(body, wavChunk("LIST", info)...)
	body = append(body, wavChunk("data", make([]byte, 32))...)

	out := []byte("RIFF")
	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(4+len(body)))
	out = append(out, size...)
	out = append(out, []byte("WAVE")...)
	out = append(out, body...)
	out = append(out, id3Title("ID3 Title")...)

	path := filepath.Join(t.TempDir(), "sample.wav")
	if err := os.WriteFile(path, out, 0644); err != nil {
		t.Fatal(err)
	}

	res, err := readPath(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Format != "WAVE" {
		t.Fatalf("format = %q", res.Format)
	}
	if res.Fields["TITLE"] != "ID3 Title" {
		t.Fatalf("ID3 should take priority: %v", res.Fields)
	}
	if res.Fields["TIT2"] != "ID3 Title" {
		t.Fatalf("ID3 fields = %v", res.Fields)
	}
}

func TestExtractWavPack(t *testing.T) {
	tag := ape.NewTag()
	tag.AddText("Title", "WavPack Song")
	tag.AddText("Artist", "WavPack Artist")
	data, err := tag.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	file := append([]byte("wvpk"), make([]byte, 256)...)
	file = append(file, data...)

	path := filepath.Join(t.TempDir(), "sample.wv")
	if err := os.WriteFile(path, file, 0644); err != nil {
		t.Fatal(err)
	}

	res, err := readPath(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Format != "WAVPACK" {
		t.Fatalf("format = %q", res.Format)
	}
	if res.Fields["Title"] != "WavPack Song" || res.Fields["Artist"] != "WavPack Artist" {
		t.Fatalf("fields = %v", res.Fields)
	}
}
