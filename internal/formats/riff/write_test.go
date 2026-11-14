package riff_test

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daregod/ext-tags/internal/formats/id3"
	"github.com/daregod/ext-tags/internal/formats/riff"
	"github.com/daregod/ext-tags/internal/metadata"
)

func TestWriteRead(t *testing.T) {
	cases := []string{"audio.wav", "audio.rf64", "audio.w64", "audio.aiff"}
	for _, fixture := range cases {
		t.Run(fixture, func(t *testing.T) {
			src := filepath.Join("..", "..", "..", "testdata", "formats", fixture)
			path := filepath.Join(t.TempDir(), fixture)
			data, err := os.ReadFile(src)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, data, 0644); err != nil {
				t.Fatal(err)
			}
			doc := metadata.New()
			doc.Fields["Title"] = "RIFF Title"
			doc.Fields["Artist"] = "RIFF Artist"
			doc.Fields["Album"] = "RIFF Album"
			doc.Fields["Genre"] = "RIFF Genre"
			doc.Fields["Year"] = "2026"
			front := &metadata.Cover{Filename: "cover.jpg", MIME: "image/jpeg", Data: []byte{0xFF, 0xD8, 0xFF}}
			doc.Cover = front
			doc.Artwork[metadata.Front] = front

			if err := riff.Write(path, doc); err != nil {
				t.Fatal(err)
			}
			md, err := riff.Read(path)
			if err != nil {
				t.Fatal(err)
			}
			for _, field := range []string{"Title", "Artist", "Album", "Genre", "Year"} {
				if got := fieldValue(md, field); !strings.EqualFold(got, doc.Fields[field]) {
					t.Errorf("%s = %q, want %q", field, got, doc.Fields[field])
				}
			}
			if md.Cover == nil || len(md.Cover.Data) == 0 {
				t.Fatal("cover missing")
			}
		})
	}
}

func fieldValue(doc *metadata.Document, name string) string {
	for key, value := range doc.Fields {
		if strings.EqualFold(key, name) {
			return value
		}
	}
	return ""
}

func TestWritePreservesInfoAndID3(t *testing.T) {
	info := []byte("INFO")
	info = append(info, riffChunk("ICMT", []byte("Keep comment\x00"))...)
	id3tag, err := id3.Build(map[string]string{"Comment": "Old ID3 comment"}, nil, 3)
	if err != nil {
		t.Fatal(err)
	}
	body := riffChunk("fmt ", make([]byte, 16))
	body = append(body, riffChunk("LIST", info)...)
	body = append(body, riffChunk("data", make([]byte, 64))...)
	body = append(body, riffChunk("id3 ", id3tag)...)
	out := []byte("RIFF")
	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(4+len(body)))
	out = append(out, size...)
	out = append(out, []byte("WAVE")...)
	out = append(out, body...)
	path := filepath.Join(t.TempDir(), "track.wav")
	if err := os.WriteFile(path, out, 0644); err != nil {
		t.Fatal(err)
	}
	doc := metadata.New()
	doc.Fields["Title"] = "New title"
	if err := riff.Write(path, doc); err != nil {
		t.Fatal(err)
	}
	md, err := riff.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fieldValue(md, "Title"); got != "New title" {
		t.Fatalf("Title = %q", got)
	}
	if got := fieldValue(md, "Comment"); got == "" {
		t.Fatalf("comment lost: %v", md.Fields)
	}
}

func riffChunk(id string, data []byte) []byte {
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

func TestWriteReadAllCovers(t *testing.T) {
	src := filepath.Join("..", "..", "..", "testdata", "formats", "audio.wav")
	path := filepath.Join(t.TempDir(), "audio.wav")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	doc := metadata.New()
	for _, kind := range []string{"front", "back", "artist", "disc", "icon"} {
		cover := &metadata.Cover{Filename: kind + ".jpg", MIME: "image/jpeg", Data: []byte{1, 2, 3}}
		doc.Artwork[kind] = cover
	}
	doc.Cover = doc.Artwork["front"]
	if err := riff.Write(path, doc); err != nil {
		t.Fatal(err)
	}
	md, err := riff.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"front", "back", "artist", "disc", "icon"} {
		if md.Artwork[kind] == nil {
			t.Fatalf("missing %s artwork: %+v", kind, md.Artwork)
		}
	}
}
