package asf

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"unicode/utf16"
)

func utf16le(s string) []byte {
	units := utf16.Encode([]rune(s))
	out := make([]byte, len(units)*2)
	for i, u := range units {
		binary.LittleEndian.PutUint16(out[i*2:], u)
	}
	return out
}

func asfObject(guid [16]byte, payload []byte) []byte {
	out := append([]byte{}, guid[:]...)
	size := make([]byte, 8)
	binary.LittleEndian.PutUint64(size, uint64(24+len(payload)))
	out = append(out, size...)
	return append(out, payload...)
}

func asfHeader(objects ...[]byte) []byte {
	var body []byte
	for _, obj := range objects {
		body = append(body, obj...)
	}
	out := append([]byte{}, guidHeader[:]...)
	size := make([]byte, 8)
	binary.LittleEndian.PutUint64(size, uint64(30+len(body)))
	out = append(out, size...)
	count := make([]byte, 4)
	binary.LittleEndian.PutUint32(count, uint32(len(objects)))
	out = append(out, count...)
	out = append(out, 1, 2) // reserved
	return append(out, body...)
}

func contentDescription(title, author, copyright, desc, rating string) []byte {
	values := []string{title, author, copyright, desc, rating}
	var lens, data []byte
	for _, v := range values {
		enc := utf16le(v)
		l := make([]byte, 2)
		binary.LittleEndian.PutUint16(l, uint16(len(enc)))
		lens = append(lens, l...)
		data = append(data, enc...)
	}
	return asfObject(guidContentDesc, append(lens, data...))
}

func descriptor(name string, valueType uint16, value []byte) []byte {
	nameEnc := utf16le(name)
	out := make([]byte, 2)
	binary.LittleEndian.PutUint16(out, uint16(len(nameEnc)))
	out = append(out, nameEnc...)
	out = append(out, byte(valueType), byte(valueType>>8))
	out = append(out, byte(len(value)), byte(len(value)>>8))
	return append(out, value...)
}

func extendedContent(descriptors ...[]byte) []byte {
	payload := make([]byte, 2)
	binary.LittleEndian.PutUint16(payload, uint16(len(descriptors)))
	for _, d := range descriptors {
		payload = append(payload, d...)
	}
	return asfObject(guidExtendedContent, payload)
}

func pictureValue(mime string, data []byte) []byte {
	out := []byte{3}
	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(data)))
	out = append(out, size...)
	out = append(out, utf16le(mime)...)
	out = append(out, 0, 0)
	out = append(out, 0, 0) // empty description
	return append(out, data...)
}

func TestReadFile(t *testing.T) {
	image := []byte{0xFF, 0xD8, 0xFF}
	file := asfHeader(
		contentDescription("G-SPOTT", "G-Spott", "(c)", "LIVE @ TENDANCE", "2"),
		extendedContent(
			descriptor("WM/AlbumTitle", 0, utf16le("Album")),
			descriptor("WM/Genre", 0, utf16le("Drum & Bass")),
			descriptor("WM/Year", 0, utf16le("2024")),
			descriptor("WM/TrackNumber", 0, utf16le("5")),
			descriptor("WM/Picture", 1, pictureValue("image/jpeg", image)),
		),
	)
	path := filepath.Join(t.TempDir(), "sample.wmv")
	if err := os.WriteFile(path, file, 0644); err != nil {
		t.Fatal(err)
	}

	md, err := ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if md.Title != "G-SPOTT" || md.Author != "G-Spott" || md.Description != "LIVE @ TENDANCE" || md.Rating != "2" {
		t.Fatalf("content description = %+v", md)
	}
	if len(md.Descriptors) != 5 {
		t.Fatalf("descriptors = %+v", md.Descriptors)
	}

	var picture []byte
	for _, d := range md.Descriptors {
		if d.Name == "WM/Picture" {
			pic, ok := ParsePicture(d.Value)
			if !ok {
				t.Fatal("picture must parse")
			}
			if pic.MIME != "image/jpeg" {
				t.Fatalf("picture mime = %q", pic.MIME)
			}
			picture = pic.Data
		}
	}
	if len(picture) != len(image) {
		t.Fatalf("picture data = %d bytes", len(picture))
	}
}
