package artwork

import "testing"

func TestMimeFor(t *testing.T) {
	cases := map[string]string{
		".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png",
		".webp": "image/webp", ".gif": "image/gif", ".bmp": "application/octet-stream",
	}
	for ext, want := range cases {
		if got := mimeFor(ext); got != want {
			t.Errorf("%s: %q, want %q", ext, got, want)
		}
	}
}
