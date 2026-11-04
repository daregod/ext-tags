package importer

import "testing"

func TestMimeFor(t *testing.T) {
	cases := map[string]string{
		"cover.jpg": "image/jpeg", "cover.jpeg": "image/jpeg",
		"cover.png": "image/png", "cover.webp": "image/webp",
		"cover.gif": "image/gif", "cover.bin": "image/jpeg",
	}
	for name, want := range cases {
		if got := mimeFor(name); got != want {
			t.Errorf("%s: %q, want %q", name, got, want)
		}
	}
}
