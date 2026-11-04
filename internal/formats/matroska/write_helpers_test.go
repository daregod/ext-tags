package matroska

import (
	"bytes"
	"testing"
)

func TestWriteHeaderAndChildRanges(t *testing.T) {
	if _, err := readHeader(nil, 0); err == nil {
		t.Fatal("nil header must fail")
	}
	if _, err := readHeader([]byte{0x00}, 0); err == nil {
		t.Fatal("bad id must fail")
	}
	if _, err := readHeader([]byte{0x80}, 0); err == nil {
		t.Fatal("missing size must fail")
	}
	if _, err := childRanges([]byte{0x00}, 0, 1); err == nil {
		t.Fatal("bad child must fail")
	}
	p := &parser{r: bytes.NewReader(nil), size: 0, meta: &Metadata{}}
	if err := p.readSimpleTag(0, 0, 0, 0); err != nil {
		t.Fatal(err)
	}
}
func TestReadSimpleTagValue(t *testing.T) {
	data := append(mkElem(idTagName, []byte("TITLE")), mkElem(idTagString, []byte("Song"))...)
	p := &parser{r: bytes.NewReader(data), size: int64(len(data)), meta: &Metadata{}}
	if err := p.readSimpleTag(0, int64(len(data)), 0, 0); err != nil {
		t.Fatal(err)
	}
	if len(p.meta.Tags) != 1 || p.meta.Tags[0].Name != "TITLE" || p.meta.Tags[0].Value != "Song" {
		t.Fatalf("tags = %+v", p.meta.Tags)
	}
}
func TestReadAttachmentValue(t *testing.T) {
	data := append(mkElem(idFileName, []byte("cover.jpg")), mkElem(idFileMime, []byte("image/jpeg"))...)
	data = append(data, mkElem(idFileData, []byte{1, 2})...)
	p := &parser{r: bytes.NewReader(data), size: int64(len(data)), meta: &Metadata{}}
	if err := p.readAttachment(0, int64(len(data))); err != nil {
		t.Fatal(err)
	}
	if len(p.meta.Attachments) != 1 || p.meta.Attachments[0].Filename != "cover.jpg" {
		t.Fatalf("attachments = %+v", p.meta.Attachments)
	}
}
func TestReadTagWithTargets(t *testing.T) {
	targets := mkElem(idTargetType, []byte{50})
	simple := append(mkElem(idTagName, []byte("TITLE")), mkElem(idTagString, []byte("Song"))...)
	tag := append(mkElem(idTargets, targets), mkElem(idSimpleTag, simple)...)
	data := tag
	p := &parser{r: bytes.NewReader(data), size: int64(len(data)), meta: &Metadata{}}
	if err := p.readTag(0, int64(len(data))); err != nil {
		t.Fatal(err)
	}
	if len(p.meta.Tags) != 1 || p.meta.Tags[0].TargetTypeValue != 50 {
		t.Fatalf("tags = %+v", p.meta.Tags)
	}
}
