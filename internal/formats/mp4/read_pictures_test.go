package mp4

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func pictureBox(dataType uint32, data []byte) []byte {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, dataType)
	payload = append(payload, data...)
	return makeBox("covr", makeBox("data", payload))
}

func mp4WithCovr(items ...[]byte) []byte {
	var ilst []byte
	for _, item := range items {
		ilst = append(ilst, item...)
	}
	meta := make([]byte, 4)
	meta = append(meta, makeBox("hdlr", make([]byte, 24))...)
	meta = append(meta, makeBox("ilst", ilst)...)
	moov := makeBox("moov", makeBox("udta", makeBox("meta", meta)))
	return append(makeBox("ftyp", []byte("M4A ")), moov...)
}

func TestReadAllPictures(t *testing.T) {
	data := mp4WithCovr(
		pictureBox(0, []byte{1}),
		pictureBox(14, []byte{2}),
		pictureBox(12, []byte{3}),
	)
	pictures, err := ReadPictures(writeTemp(t, data))
	if err != nil {
		t.Fatal(err)
	}
	if len(pictures) != 3 {
		t.Fatalf("pictures = %+v", pictures)
	}
	if pictures[1].MIME != "image/png" || pictures[2].MIME != "image/gif" {
		t.Fatalf("mimes = %+v", pictures)
	}
}

func TestReadPicturesMissingAndMalformed(t *testing.T) {
	if _, err := ReadPictures(filepath.Join(t.TempDir(), "missing.m4a")); err == nil {
		t.Fatal("expected missing error")
	}
	path := writeTemp(t, append(makeBox("ftyp", []byte("M4A ")), []byte("bad")...))
	if _, err := ReadPictures(path); err == nil {
		t.Fatal("expected malformed error")
	}
	path = writeTemp(t, mp4WithCovr())
	if pictures, err := ReadPictures(path); err != nil || len(pictures) != 0 {
		t.Fatalf("empty = %+v %v", pictures, err)
	}
}

func TestWriteOnlyReadPictures(t *testing.T) {
	src := filepath.Join("..", "..", "..", "testdata", "formats", "audio-aac.m4a")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	path := writeTemp(t, data)
	if err := WriteOnly(path, map[string]string{"Title": "New"}, &Cover{Filename: "c.png", MIME: "image/png", Data: []byte{9}}, false, []string{"Title", "Cover Art (Front)"}); err != nil {
		t.Fatal(err)
	}
	pictures, err := ReadPictures(path)
	if err != nil || len(pictures) != 1 || pictures[0].MIME != "image/png" {
		t.Fatalf("pictures = %+v %v", pictures, err)
	}
	if err := WriteOnly(path, nil, nil, false, []string{"Unknown"}); err == nil {
		t.Fatal("expected unsupported field error")
	}
}

func writeTemp(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "track.m4a")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestMP4ReadAndWriteErrors(t *testing.T) {
	if pictures, err := ReadPictures(writeTemp(t, makeBox("ftyp", []byte("M4A ")))); err != nil || len(pictures) != 0 {
		t.Fatalf("no moov = %+v %v", pictures, err)
	}
	bad := makeBox("moov", []byte("bad"))
	if _, err := ReadPictures(writeTemp(t, bad)); err == nil {
		t.Fatal("expected malformed moov")
	}
	shortMeta := makeBox("moov", makeBox("udta", makeBox("meta", []byte{1, 2, 3})))
	if pictures, err := ReadPictures(writeTemp(t, shortMeta)); err != nil || len(pictures) != 0 {
		t.Fatalf("short meta = %+v %v", pictures, err)
	}
	if err := Write(writeTemp(t, makeBox("ftyp", []byte("M4A "))), nil, nil, false); err == nil {
		t.Fatal("expected missing moov")
	}
}
