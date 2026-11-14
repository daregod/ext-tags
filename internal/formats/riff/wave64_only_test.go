package riff

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/metadata"
)

func TestWriteSelectiveWave64Info(t *testing.T) {
	info := []byte("INFO")
	info = append(info, makeChunkLE("INAM", []byte("Old title\x00"))...)
	info = append(info, makeChunkLE("ICMT", []byte("Keep comment\x00"))...)
	listGUID := make([]byte, 16)
	copy(listGUID, "LIST")
	body := wave64Chunk(listGUID, info)
	out := make([]byte, 40)
	copy(out[0:4], "riff")
	copy(out[24:28], "wave")
	out = append(out, body...)
	binary.LittleEndian.PutUint64(out[16:24], uint64(len(out)))
	path := filepath.Join(t.TempDir(), "track.w64")
	if err := os.WriteFile(path, out, 0644); err != nil {
		t.Fatal(err)
	}
	doc := metadata.New()
	doc.Fields["Title"] = "New title"
	if err := WriteSelective(path, doc, []string{"Title"}, false); err != nil {
		t.Fatal(err)
	}
	md, err := ReadWave64(path)
	if err != nil {
		t.Fatal(err)
	}
	if md.Info["INAM"] != "" {
		t.Fatalf("INAM = %q", md.Info["INAM"])
	}
	if md.Info["ICMT"] != "Keep comment" {
		t.Fatalf("ICMT = %q", md.Info["ICMT"])
	}
}
