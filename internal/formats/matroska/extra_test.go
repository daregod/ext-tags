package matroska

import (
	"bytes"
	"testing"
)

func TestIsAllOnes(t *testing.T) {
	if !isAllOnes([]byte{0xFF}) || !isAllOnes([]byte{0x7F, 0xFF}) {
		t.Fatal("all ones expected")
	}
	if isAllOnes([]byte{0x7F, 0xFE}) {
		t.Fatal("not all ones")
	}
}

func TestVintLen(t *testing.T) {
	cases := []struct {
		first byte
		id    bool
		want  int
		ok    bool
	}{
		{0x80, false, 1, true},
		{0x40, false, 2, true},
		{0x20, true, 3, true},
		{0x10, true, 4, true},
		{0x08, true, 0, false},
		{0x00, false, 0, false},
	}
	for _, tc := range cases {
		got, ok := vintLen(tc.first, tc.id)
		if got != tc.want || ok != tc.ok {
			t.Errorf("vintLen(%#x,%v) = %d,%v", tc.first, tc.id, got, ok)
		}
	}
}

func TestEncodeSize(t *testing.T) {
	for _, n := range []int{0, 1, 126, 127, 128, 16000} {
		data := encodeSize(n)
		if len(data) == 0 || data[0] == 0 {
			t.Fatalf("n=%d data=% X", n, data)
		}
		if !bytes.Equal(encodeSize(n), data) {
			t.Fatalf("unstable for %d", n)
		}
	}
}
