package id3

import "testing"

func TestOnlyHelpers(t *testing.T) {
	ids, txxx := onlyRemovals(3, []string{"Title", "Comment", "Rating", "Cover Art (Front)"})
	if !ids["TIT2"] || !ids["COMM"] || !txxx["rating"] {
		t.Fatalf("ids=%v txxx=%v", ids, txxx)
	}
	for field, kind := range map[string]string{
		"Cover Art (Front)": "front",
		"Cover Art (Back)":  "back",
		"Artist Picture":    "artist",
		"Cover Art (Disc)":  "disc",
		"Cover Art (Icon)":  "icon",
		"other":             "",
	} {
		if got := artworkKind(field); got != kind {
			t.Fatalf("%s = %q", field, got)
		}
	}
}

func TestTXXXDescription(t *testing.T) {
	cases := [][]byte{
		append([]byte{0}, []byte("latin\x00")...),
		append([]byte{3}, []byte("utf8\x00")...),
		{1, 0xFF, 0xFE, 'x', 0, 0, 0},
		{2, 0, 'x', 0, 0},
		{},
	}
	for _, payload := range cases {
		_ = txxxDescription(payload)
	}
	if got := txxxDescription([]byte{3, 'a', 'b', 0}); got != "ab" {
		t.Fatalf("utf8 = %q", got)
	}
}

func TestAPICKind(t *testing.T) {
	front := append([]byte{0}, []byte("image/jpeg\x00")...)
	front = append(front, 3, 0)
	if got := apicKind(3, front); got != "front" {
		t.Fatalf("front = %q", got)
	}
	pic := []byte{0, 'J', 'P', 'G', 4, 0}
	if got := apicKind(2, pic); got != "back" {
		t.Fatalf("pic = %q", got)
	}
	if got := apicKind(3, []byte{0}); got != "" {
		t.Fatalf("short = %q", got)
	}
}
