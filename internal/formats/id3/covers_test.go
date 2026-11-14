package id3

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func testRawFrame(id string, payload []byte) []byte {
	out := []byte(id)
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(payload)))
	out = append(out, size...)
	out = append(out, 0, 0)
	return append(out, payload...)
}

func testMakeTag(frames ...[]byte) []byte {
	var body []byte
	for _, frame := range frames {
		body = append(body, frame...)
	}
	out := []byte("ID3\x03\x00\x00")
	size := len(body)
	out = append(out, byte(size>>21&0x7f), byte(size>>14&0x7f), byte(size>>7&0x7f), byte(size&0x7f))
	return append(out, body...)
}

func testAPICPayload(mime string, pictureType byte, description string, data []byte) []byte {
	out := []byte{3}
	out = append(out, []byte(mime)...)
	out = append(out, 0, pictureType)
	out = append(out, []byte(description)...)
	out = append(out, 0)
	return append(out, data...)
}

func TestReadPicturesAllSlots(t *testing.T) {
	tag := testMakeTag(
		testRawFrame("APIC", testAPICPayload("image/jpeg", 3, "front", []byte{1})),
		testRawFrame("APIC", testAPICPayload("image/png", 4, "back", []byte{2})),
		testRawFrame("APIC", testAPICPayload("image/webp", 8, "artist", []byte{3})),
		testRawFrame("APIC", testAPICPayload("image/gif", 6, "disc", []byte{4})),
		testRawFrame("APIC", testAPICPayload("image/png", 1, "icon", []byte{5})),
	)
	pictures := ReadPictures(tag)
	if len(pictures) != 5 {
		t.Fatalf("pictures = %d", len(pictures))
	}
	for i, picture := range pictures {
		if int(picture.Type) != []int{3, 4, 8, 6, 1}[i] {
			t.Fatalf("type[%d] = %d", i, picture.Type)
		}
	}
	covers := ReadCovers(tag)
	for _, kind := range []string{"front", "back", "artist", "disc", "icon"} {
		if covers[kind] == nil {
			t.Fatalf("missing %s", kind)
		}
	}
}

func testMakeV22Tag(payload []byte) []byte {
	frame := []byte("PIC")
	size := len(payload)
	frame = append(frame, byte(size>>16), byte(size>>8), byte(size))
	frame = append(frame, payload...)
	out := []byte("ID3\x02\x00\x00")
	out = append(out, byte(len(frame)>>21&0x7f), byte(len(frame)>>14&0x7f), byte(len(frame)>>7&0x7f), byte(len(frame)&0x7f))
	return append(out, frame...)
}

func TestReadPicturesV22AndEncoding(t *testing.T) {
	pic := []byte{0, 'P', 'N', 'G', 3, 0, 0, 0, 9}
	tag := testMakeV22Tag(pic)
	pictures := ReadPictures(tag)
	if len(pictures) != 1 || pictures[0].MIME != "image/png" || string(pictures[0].Data) != "\x00\x00\x09" {
		t.Fatalf("pictures = %+v", pictures)
	}
	utf16 := []byte{1}
	utf16 = append(utf16, []byte("image/jpeg")...)
	utf16 = append(utf16, 0, 3)
	utf16 = append(utf16, 'd', 0, 0, 0)
	utf16 = append(utf16, 7)
	utf16Tag := testMakeTag(testRawFrame("APIC", utf16))
	if got := ReadPictures(utf16Tag); len(got) != 1 || string(got[0].Data) != "\x07" {
		t.Fatalf("utf16 = %+v", got)
	}
}

func TestReadPicturesMalformedAndFile(t *testing.T) {
	if got := ReadPictures([]byte("bad")); got != nil {
		t.Fatalf("bad = %v", got)
	}
	if got := ReadPictures(testMakeTag(testRawFrame("APIC", []byte{3, 'x'}))); len(got) != 0 {
		t.Fatalf("malformed = %v", got)
	}
	path := filepath.Join(t.TempDir(), "track.mp3")
	if err := os.WriteFile(path, append(testMakeTag(testRawFrame("APIC", testAPICPayload("image/jpeg", 3, "", []byte{1}))), 0, 1), 0644); err != nil {
		t.Fatal(err)
	}
	if got := ReadFilePictures(path); len(got) != 1 {
		t.Fatalf("file = %v", got)
	}
	if got := ReadFilePictures(filepath.Join(t.TempDir(), "missing")); got != nil {
		t.Fatalf("missing = %v", got)
	}
}

func TestCoverFilename(t *testing.T) {
	cases := map[string]string{
		"image/png":  "cover.png",
		"image/gif":  "cover.gif",
		"image/webp": "cover.webp",
		"image/jpeg": "cover.jpg",
	}
	for mime, want := range cases {
		if got := coverFilename(mime); got != want {
			t.Fatalf("%s = %q", mime, got)
		}
	}
}
