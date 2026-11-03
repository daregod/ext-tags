package extractor

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"unicode/utf16"
)

func testUTF16(s string) []byte {
	units := utf16.Encode([]rune(s))
	out := make([]byte, len(units)*2)
	for i, u := range units {
		binary.LittleEndian.PutUint16(out[i*2:], u)
	}
	return out
}

func testASFObject(guid []byte, payload []byte) []byte {
	out := append([]byte{}, guid...)
	size := make([]byte, 8)
	binary.LittleEndian.PutUint64(size, uint64(24+len(payload)))
	out = append(out, size...)
	return append(out, payload...)
}

func TestExtractASF(t *testing.T) {
	headerGUID := []byte{0x30, 0x26, 0xB2, 0x75, 0x8E, 0x66, 0xCF, 0x11, 0xA6, 0xD9, 0x00, 0xAA, 0x00, 0x62, 0xCE, 0x6C}
	contentGUID := []byte{0x33, 0x26, 0xB2, 0x75, 0x8E, 0x66, 0xCF, 0x11, 0xA6, 0xD9, 0x00, 0xAA, 0x00, 0x62, 0xCE, 0x6C}
	extendedGUID := []byte{0x40, 0xA4, 0xD0, 0xD2, 0x07, 0xE3, 0xD2, 0x11, 0x97, 0xF0, 0x00, 0xA0, 0xC9, 0x5E, 0xA8, 0x50}

	values := []string{"G-SPOTT", "G-Spott", "(c)", "LIVE @ TENDANCE", "2"}
	var lens, data []byte
	for _, v := range values {
		enc := testUTF16(v)
		l := make([]byte, 2)
		binary.LittleEndian.PutUint16(l, uint16(len(enc)))
		lens = append(lens, l...)
		data = append(data, enc...)
	}
	content := testASFObject(contentGUID, append(lens, data...))

	name := testUTF16("WM/AlbumTitle")
	value := testUTF16("Album")
	desc := make([]byte, 2)
	binary.LittleEndian.PutUint16(desc, uint16(len(name)))
	desc = append(desc, name...)
	desc = append(desc, 0, 0) // type 0
	desc = append(desc, byte(len(value)), byte(len(value)>>8))
	desc = append(desc, value...)
	extPayload := []byte{1, 0}
	extPayload = append(extPayload, desc...)
	extended := testASFObject(extendedGUID, extPayload)

	body := append(content, extended...)
	out := append([]byte{}, headerGUID...)
	size := make([]byte, 8)
	binary.LittleEndian.PutUint64(size, uint64(30+len(body)))
	out = append(out, size...)
	count := make([]byte, 4)
	binary.LittleEndian.PutUint32(count, 2)
	out = append(out, count...)
	out = append(out, 1, 2)
	out = append(out, body...)

	path := filepath.Join(t.TempDir(), "sample.wmv")
	if err := os.WriteFile(path, out, 0644); err != nil {
		t.Fatal(err)
	}

	res, err := Extract(path)
	if err != nil {
		t.Fatal(err)
	}
	if res.Format != "ASF" {
		t.Fatalf("format = %q", res.Format)
	}
	if res.Fields["TITLE"] != "G-SPOTT" || res.Fields["COMMENT"] != "LIVE @ TENDANCE" {
		t.Fatalf("fields = %v", res.Fields)
	}
	if res.Fields["RATING"] != "2" || res.Fields["ALBUM"] != "Album" {
		t.Fatalf("fields = %v", res.Fields)
	}
}
