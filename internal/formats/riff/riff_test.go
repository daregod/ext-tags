package riff

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func chunk(id string, data []byte) []byte {
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

func TestReadInfoAndAppendedID3(t *testing.T) {
	info := []byte("INFO")
	info = append(info, chunk("INAM", []byte("Song\x00"))...)
	info = append(info, chunk("IART", []byte("Artist\x00"))...)

	body := chunk("fmt ", make([]byte, 16))
	body = append(body, chunk("LIST", info)...)
	body = append(body, chunk("data", make([]byte, 64))...)

	riff := []byte("RIFF")
	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(4+len(body)))
	riff = append(riff, size...)
	riff = append(riff, []byte("WAVE")...)
	riff = append(riff, body...)

	id3 := append([]byte("ID3\x03\x00\x00\x00\x00\x00\x0a"), make([]byte, 10)...)
	riff = append(riff, id3...)

	path := filepath.Join(t.TempDir(), "sample.wav")
	if err := os.WriteFile(path, riff, 0644); err != nil {
		t.Fatal(err)
	}

	md, err := ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if md.Info["INAM"] != "Song" || md.Info["IART"] != "Artist" {
		t.Fatalf("info = %v", md.Info)
	}
	if len(md.ID3) != len(id3) {
		t.Fatalf("appended ID3 = %d bytes, want %d", len(md.ID3), len(id3))
	}
}
