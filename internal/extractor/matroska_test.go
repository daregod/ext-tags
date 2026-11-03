package extractor

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func ebmlVint(n int) []byte {
	for length := 1; length <= 8; length++ {
		if n < (1<<(7*length))-1 {
			buf := make([]byte, length)
			v := n
			for i := length - 1; i >= 0; i-- {
				buf[i] = byte(v & 0xFF)
				v >>= 8
			}
			buf[0] |= 0x80 >> (length - 1)
			return buf
		}
	}
	return nil
}

func ebmlElem(id []byte, parts ...[]byte) []byte {
	var body bytes.Buffer
	for _, part := range parts {
		body.Write(part)
	}
	out := append([]byte{}, id...)
	out = append(out, ebmlVint(body.Len())...)
	out = append(out, body.Bytes()...)
	return out
}

func ebmlText(n string) []byte { return []byte(n) }

func ebmlSimple(name, value string) []byte {
	return ebmlElem([]byte{0x67, 0xC8},
		ebmlElem([]byte{0x45, 0xA3}, ebmlText(name)),
		ebmlElem([]byte{0x44, 0x87}, ebmlText(value)),
	)
}

func TestExtractWebM(t *testing.T) {
	info := ebmlElem([]byte{0x15, 0x49, 0xA9, 0x66},
		ebmlElem([]byte{0x7B, 0xA9}, ebmlText("Container Title")),
	)
	tags := ebmlElem([]byte{0x12, 0x54, 0xC3, 0x67},
		ebmlElem([]byte{0x73, 0x73},
			ebmlSimple("ARTIST", "Noisia"),
			ebmlSimple("TITLE", "Stigma"),
			ebmlSimple("GENRE", "Drum & Bass"),
			ebmlSimple("PART_NUMBER", "3"),
			ebmlSimple("TOTAL_PARTS", "12"),
		),
	)
	image := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	attachments := ebmlElem([]byte{0x19, 0x41, 0xA4, 0x69},
		ebmlElem([]byte{0x61, 0xA7},
			ebmlElem([]byte{0x46, 0x6E}, ebmlText("cover.jpg")),
			ebmlElem([]byte{0x46, 0x60}, ebmlText("image/jpeg")),
			ebmlElem([]byte{0x46, 0x5C}, image),
		),
	)
	segment := ebmlElem([]byte{0x18, 0x53, 0x80, 0x67}, info, tags, attachments)
	header := ebmlElem([]byte{0x1A, 0x45, 0xDF, 0xA3}, ebmlElem([]byte{0x42, 0x86}, []byte{1}))

	path := filepath.Join(t.TempDir(), "sample.webm")
	if err := os.WriteFile(path, append(header, segment...), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := Extract(path)
	if err != nil {
		t.Fatal(err)
	}
	if res.Format != "MATROSKA" {
		t.Fatalf("format = %q", res.Format)
	}
	if res.Fields["TITLE"] != "Stigma" || res.Fields["ALBUMARTIST"] != "Noisia" {
		t.Fatalf("fields = %v", res.Fields)
	}
	if res.Fields["GENRE"] != "Drum & Bass" {
		t.Fatalf("genre = %q", res.Fields["GENRE"])
	}
	if res.Vars["discnumber"] != "3" || res.Vars["disctotal"] != "12" {
		t.Fatalf("vars = %v", res.Vars)
	}
	if res.Cover == nil || !bytes.Equal(res.Cover.Data, image) {
		t.Fatalf("cover = %+v", res.Cover)
	}
	if res.Cover.Filename != "cover.jpg" {
		t.Fatalf("cover filename = %q", res.Cover.Filename)
	}
}

func TestExtractWebMNoTags(t *testing.T) {
	header := ebmlElem([]byte{0x1A, 0x45, 0xDF, 0xA3}, ebmlElem([]byte{0x42, 0x86}, []byte{1}))
	segment := ebmlElem([]byte{0x18, 0x53, 0x80, 0x67},
		ebmlElem([]byte{0x15, 0x49, 0xA9, 0x66},
			ebmlElem([]byte{0x4D, 0x80}, ebmlText("muxer")),
		),
	)
	path := filepath.Join(t.TempDir(), "empty.webm")
	if err := os.WriteFile(path, append(header, segment...), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Extract(path); err == nil {
		t.Fatal("expected ErrNoTags")
	}
}
