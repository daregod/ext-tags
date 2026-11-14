package yamlutil

import (
	"bytes"
	"testing"
)

func TestTrimBOM(t *testing.T) {
	if got := TrimBOM([]byte("hello")); !bytes.Equal(got, []byte("hello")) {
		t.Fatalf("plain = %q", got)
	}
	bom := []byte{0xEF, 0xBB, 0xBF}
	if got := TrimBOM(append(bom, []byte("hello")...)); !bytes.Equal(got, []byte("hello")) {
		t.Fatalf("bom = %q", got)
	}
	if got := TrimBOM(nil); got != nil {
		t.Fatalf("nil = %q", got)
	}
}
