package asf

import (
	"encoding/binary"
	"testing"
)

func makePictureValue(mime, description string, data []byte) []byte {
	out := []byte{3}
	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(data)))
	out = append(out, size...)
	out = append(out, utf16le(mime)...)
	out = append(out, 0, 0)
	out = append(out, utf16le(description)...)
	out = append(out, 0, 0)
	return append(out, data...)
}

func TestParsePicture(t *testing.T) {
	pic, ok := ParsePicture(makePictureValue("image/jpeg", "front", []byte{1, 2}))
	if !ok || pic.MIME != "image/jpeg" || len(pic.Data) != 2 {
		t.Fatalf("picture = %+v, %v", pic, ok)
	}
	for _, value := range [][]byte{
		{},
		{3, 0, 0, 0, 0},
		{3, 1, 0, 0, 0, 'x'},
		append(makePictureValue("image/jpeg", "", nil), 1),
	} {
		if _, ok := ParsePicture(value); ok {
			t.Fatalf("value % X must fail", value)
		}
	}
}
