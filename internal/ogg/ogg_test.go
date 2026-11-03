package ogg

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func oggPage(serial uint32, packets ...[]byte) []byte {
	var lacing, data []byte
	for _, packet := range packets {
		n := len(packet)
		for n >= 255 {
			lacing = append(lacing, 255)
			n -= 255
		}
		lacing = append(lacing, byte(n))
		data = append(data, packet...)
	}

	header := make([]byte, 27)
	copy(header, "OggS")
	header[4] = 0
	header[5] = 2
	binary.LittleEndian.PutUint32(header[14:18], serial)
	binary.LittleEndian.PutUint32(header[18:22], 0)
	header[26] = byte(len(lacing))
	return append(append(header, lacing...), data...)
}

func commentBlock(vendor string, tags ...[2]string) []byte {
	var out []byte
	lenBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBuf, uint32(len(vendor)))
	out = append(out, lenBuf...)
	out = append(out, vendor...)
	binary.LittleEndian.PutUint32(lenBuf, uint32(len(tags)))
	out = append(out, lenBuf...)
	for _, tag := range tags {
		line := tag[0] + "=" + tag[1]
		binary.LittleEndian.PutUint32(lenBuf, uint32(len(line)))
		out = append(out, lenBuf...)
		out = append(out, line...)
	}
	return out
}

func TestSpeexComments(t *testing.T) {
	header := append([]byte("Speex   "), make([]byte, 72)...)
	path := writeTemp(t, oggPage(1, header, commentBlock("libspeex", [2]string{"TITLE", "Song"}, [2]string{"ARTIST", "Artist"})))
	tags, pic, err := ReadTags(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 2 || tags[0].Name != "TITLE" || tags[0].Value != "Song" {
		t.Fatalf("tags = %+v", tags)
	}
	if pic != nil {
		t.Fatal("unexpected picture")
	}
}

func TestOggFLACComments(t *testing.T) {
	first := append([]byte{0x7F}, []byte("FLAC")...)
	first = append(first, make([]byte, 64)...)
	comments := commentBlock("reference libFLAC", [2]string{"TITLE", "Flac Song"})
	block := append([]byte{0x84, 0, 0, 0}, comments...)
	block[1] = byte(len(comments) >> 16)
	block[2] = byte(len(comments) >> 8)
	block[3] = byte(len(comments))

	path := writeTemp(t, oggPage(2, first, block))
	tags, _, err := ReadTags(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 1 || tags[0].Value != "Flac Song" {
		t.Fatalf("tags = %+v", tags)
	}
}
func writeTemp(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sample.ogg")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
