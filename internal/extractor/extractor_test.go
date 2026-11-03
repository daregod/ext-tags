package extractor

import "testing"

func TestTrackNumbers(t *testing.T) {
	num, total := trackNumbers(map[string]interface{}{
		"TRACKNUMBER": "02",
		"TRACKTOTAL":  "16",
	})
	if num != "02" || total != "16" {
		t.Fatalf("trackNumbers = %q/%q, want 02/16", num, total)
	}
}

func TestTrackNumbersSlash(t *testing.T) {
	num, total := trackNumbers(map[string]interface{}{"TRCK": "5/12"})
	if num != "5" || total != "12" {
		t.Fatalf("trackNumbers = %q/%q, want 5/12", num, total)
	}
}

func TestDiscNumbers(t *testing.T) {
	num, total := discNumbers(map[string]interface{}{"TPOS": "1/2"})
	if num != "1" || total != "2" {
		t.Fatalf("discNumbers = %q/%q, want 1/2", num, total)
	}
}

func TestNormalizeKeyMP4(t *testing.T) {
	key, ok := normalizeKey("\xa9nam")
	if !ok || key != "©nam" {
		t.Fatalf("normalizeKey = %q/%v, want ©nam/true", key, ok)
	}
}

func TestStringify(t *testing.T) {
	val, ok := stringify([]string{" a ", "b", "a", ""})
	if !ok || val != "a; b" {
		t.Fatalf("stringify = %q/%v, want a; b/true", val, ok)
	}
}

func TestIsSupported(t *testing.T) {
	exts := []string{".mp3", ".flac"}
	if !IsSupported("x/y/track.MP3", exts) {
		t.Fatal("MP3 must be supported")
	}
	if IsSupported("x/y/track.wmv", exts) {
		t.Fatal("wmv must not be supported")
	}
}
