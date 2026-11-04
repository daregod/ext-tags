package matroska

import (
	"bytes"
	"testing"
)

func vint(n int) []byte {
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

func elem(id []byte, parts ...[]byte) []byte {
	var body bytes.Buffer
	for _, part := range parts {
		body.Write(part)
	}
	out := append([]byte{}, id...)
	out = append(out, vint(body.Len())...)
	out = append(out, body.Bytes()...)
	return out
}

func text(n string) []byte { return []byte(n) }

func simpleTag(name, value string) []byte {
	return elem([]byte{0x67, 0xC8},
		elem([]byte{0x45, 0xA3}, text(name)),
		elem([]byte{0x44, 0x87}, text(value)),
	)
}

func TestParseTagsTitleAndAttachment(t *testing.T) {
	info := elem([]byte{0x15, 0x49, 0xA9, 0x66},
		elem([]byte{0x7B, 0xA9}, text("Container Title")),
		elem([]byte{0x4D, 0x80}, text("muxer")),
	)
	tags := elem([]byte{0x12, 0x54, 0xC3, 0x67},
		elem([]byte{0x73, 0x73},
			elem([]byte{0x63, 0xC0}),
			simpleTag("ARTIST", "Noisia"),
			simpleTag("TITLE", "Stigma"),
			simpleTag("GENRE", "Drum & Bass"),
			simpleTag("PART_NUMBER", "3"),
			simpleTag("TOTAL_PARTS", "12"),
		),
	)
	image := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	attachments := elem([]byte{0x19, 0x41, 0xA4, 0x69},
		elem([]byte{0x61, 0xA7},
			elem([]byte{0x46, 0x6E}, text("poster.jpg")),
			elem([]byte{0x46, 0x60}, text("image/jpeg")),
			elem([]byte{0x46, 0x5C}, image),
		),
	)
	segment := elem([]byte{0x18, 0x53, 0x80, 0x67}, info, tags, attachments)

	md, err := Parse(segment)
	if err != nil {
		t.Fatal(err)
	}
	if md.Title != "Container Title" {
		t.Fatalf("title = %q", md.Title)
	}
	if len(md.Tags) != 5 {
		t.Fatalf("tags = %+v", md.Tags)
	}
	if md.Tags[0].Name != "ARTIST" || md.Tags[0].Value != "Noisia" {
		t.Fatalf("first tag = %+v", md.Tags[0])
	}
	if len(md.Attachments) != 1 {
		t.Fatalf("attachments = %+v", md.Attachments)
	}
	if md.Attachments[0].Filename != "poster.jpg" || md.Attachments[0].MIME != "image/jpeg" {
		t.Fatalf("attachment = %+v", md.Attachments[0])
	}
	if !bytes.Equal(md.Attachments[0].Data, image) {
		t.Fatal("attachment data mismatch")
	}
}

func TestParseSkipsCluster(t *testing.T) {
	cluster := elem([]byte{0x1F, 0x43, 0xB6, 0x75}, bytes.Repeat([]byte{0xAA}, 1024))
	tags := elem([]byte{0x12, 0x54, 0xC3, 0x67},
		elem([]byte{0x73, 0x73}, simpleTag("TITLE", "After Cluster")),
	)
	segment := elem([]byte{0x18, 0x53, 0x80, 0x67}, cluster, tags)

	md, err := Parse(segment)
	if err != nil {
		t.Fatal(err)
	}
	if len(md.Tags) != 1 || md.Tags[0].Value != "After Cluster" {
		t.Fatalf("tags = %+v", md.Tags)
	}
}

func TestParseEmpty(t *testing.T) {
	if _, err := Parse(nil); err != nil {
		t.Fatalf("empty input must not fail: %v", err)
	}
}
