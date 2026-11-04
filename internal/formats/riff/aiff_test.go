package riff

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func aiffChunk(id string, data []byte) []byte {
	out := []byte(id)
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(data)))
	out = append(out, size...)
	out = append(out, data...)
	if len(data)%2 == 1 {
		out = append(out, 0)
	}
	return out
}

func TestReadAIFF(t *testing.T) {
	id3 := []byte("ID3\x03\x00\x00\x00\x00\x00\x00")
	body := aiffChunk("NAME", []byte("Title\x00"))
	body = append(body, aiffChunk("AUTH", []byte("Artist\x00"))...)
	body = append(body, aiffChunk("ID3 ", id3)...)

	out := []byte("FORM")
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(4+len(body)))
	out = append(out, size...)
	out = append(out, []byte("AIFF")...)
	out = append(out, body...)

	path := filepath.Join(t.TempDir(), "sample.aiff")
	if err := os.WriteFile(path, out, 0644); err != nil {
		t.Fatal(err)
	}

	md, err := ReadAIFF(path)
	if err != nil {
		t.Fatal(err)
	}
	if md.Info["NAME"] != "Title" || md.Info["AUTH"] != "Artist" {
		t.Fatalf("info = %v", md.Info)
	}
	if len(md.ID3) != len(id3) {
		t.Fatalf("ID3 = %d bytes, want %d", len(md.ID3), len(id3))
	}
}
