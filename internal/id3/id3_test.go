package id3

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func writeSample(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sample.mp3")
	if err := os.WriteFile(path, []byte("AUDIODATA-0123456789"), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func fields() map[string]string {
	return map[string]string{
		"Title": "Title", "Artist": "Artist", "Album": "Album",
		"Year": "1999", "Track": "3", "Genre": "Rock", "Catalog#": "CAT-1",
	}
}

func TestWriteVersions(t *testing.T) {
	cover := &Cover{Filename: "cover.png", MIME: "image/png", Data: []byte{1, 2, 3}}
	for _, version := range []int{2, 3, 4} {
		t.Run(string(rune('0'+version)), func(t *testing.T) {
			path := writeSample(t)
			if err := Write(path, fields(), cover, Options{Version: version, Replace: true}); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(data[0:3]) != "ID3" || int(data[3]) != version {
				t.Fatalf("header = % X", data[0:10])
			}
			frames, gotVersion, _, has, err := parseTag(data)
			if err != nil || !has || gotVersion != version {
				t.Fatalf("parse: version=%d has=%v err=%v", gotVersion, has, err)
			}
			ids := map[string]bool{}
			for _, f := range frames {
				ids[f.id] = true
			}
			titleID := "TIT2"
			yearID := "TYER"
			picID := "APIC"
			if version == 2 {
				titleID, yearID, picID = "TT2", "TYE", "PIC"
			}
			if version == 4 {
				yearID = "TDRC"
			}
			if !ids[titleID] || !ids[yearID] || !ids[picID] || !ids[txxxID(version)] {
				t.Fatalf("frames = %v", ids)
			}
			// Audio data must be preserved after the tag.
			if !bytes.HasSuffix(data, []byte("AUDIODATA-0123456789")) {
				t.Fatal("audio payload was not preserved")
			}
		})
	}
}

func TestID3v1(t *testing.T) {
	path := writeSample(t)
	f := fields()
	if err := Write(path, f, nil, Options{Version: 3, V1: true, Replace: true}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	v1 := data[len(data)-128:]
	if string(v1[0:3]) != "TAG" {
		t.Fatalf("id3v1 = %q", v1[0:3])
	}
	if got := string(bytes.TrimRight(v1[3:33], "\x00")); got != "Title" {
		t.Fatalf("title = %q", got)
	}
	if got := string(bytes.TrimRight(v1[33:63], "\x00")); got != "Artist" {
		t.Fatalf("artist = %q", got)
	}
	if got := string(bytes.TrimRight(v1[93:97], "\x00")); got != "1999" {
		t.Fatalf("year = %q", got)
	}
	if v1[125] != 0 || v1[126] != 3 {
		t.Fatalf("track bytes = %d %d", v1[125], v1[126])
	}
	if v1[127] != 17 {
		t.Fatalf("genre = %d, want Rock=17", v1[127])
	}
}

func TestPreserveAndReplace(t *testing.T) {
	path := writeSample(t)
	if err := Write(path, map[string]string{"Album": "Album", "Artist": "Artist"}, nil, Options{Version: 3, Replace: true}); err != nil {
		t.Fatal(err)
	}
	if err := Write(path, map[string]string{"Title": "Title"}, nil, Options{Version: 3}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	frames, _, _, _, err := parseTag(data)
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, f := range frames {
		ids[f.id] = true
	}
	if !ids["TIT2"] || !ids["TALB"] || !ids["TPE1"] {
		t.Fatalf("preserved frames = %v", ids)
	}

	if err := Write(path, map[string]string{"Title": "Title"}, nil, Options{Version: 3, Replace: true}); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(path)
	frames, _, _, _, err = parseTag(data)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range frames {
		if f.id == "TALB" || f.id == "TPE1" {
			t.Fatalf("frame %s survived --replace", f.id)
		}
	}
}
