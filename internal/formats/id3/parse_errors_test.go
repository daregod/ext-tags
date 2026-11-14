package id3

import "testing"

func TestParseTagErrors(t *testing.T) {
	cases := [][]byte{
		[]byte("ID3\x01\x00\x00\x00\x00\x00\x00"),
		[]byte("ID3\x03\x00\x00\x00\x00\x10\x00"),
		[]byte("ID3\x03\x00\x40\x00\x00\x00\x00"),
	}
	for _, data := range cases {
		if _, _, _, _, err := parseTag(data); err != nil {
			continue
		}
		t.Fatalf("expected error for %q", data)
	}
	if got := stripV1([]byte("audio")); string(got) != "audio" {
		t.Fatalf("stripV1 plain = %q", got)
	}
}

func TestStripV1AndExtendedHeader(t *testing.T) {
	data := make([]byte, 128)
	copy(data, "TAG")
	if got := stripV1(data); len(got) != 0 {
		t.Fatalf("stripV1 = %d", len(got))
	}
	cases := [][]byte{
		[]byte("ID3\x04\x00\x40\x00\x00\x00\x01"),
		[]byte("ID3\x04\x00\x40\x00\x00\x00\x00\x00"),
	}
	for _, data := range cases {
		if _, _, _, _, err := parseTag(data); err == nil {
			t.Fatalf("expected extended header error for %q", data)
		}
	}
}

func TestParseEmptyVersions(t *testing.T) {
	for _, tag := range [][]byte{
		[]byte("ID3\x02\x00\x00\x00\x00\x00\x00"),
		[]byte("ID3\x04\x00\x00\x00\x00\x00\x00"),
	} {
		if _, _, _, ok, err := parseTag(tag); err != nil || !ok {
			t.Fatalf("tag %q: %v", tag, err)
		}
	}
}
