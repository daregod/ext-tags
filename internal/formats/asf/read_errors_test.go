package asf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadFileErrors(t *testing.T) {
	dir := t.TempDir()
	if _, err := ReadFile(filepath.Join(dir, "missing.wma")); err == nil {
		t.Fatal("missing must fail")
	}
	for name, data := range map[string][]byte{
		"short.wma": []byte("short"),
		"bad.wma":   make([]byte, 30),
	} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadFile(path); err == nil {
			t.Fatalf("%s must fail", name)
		}
	}
	badSize := make([]byte, 30)
	copy(badSize, guidHeader[:])
	if _, err := ReadFile(filepath.Join(dir, "bad-size.wma")); err == nil {
		t.Fatal("invalid size must fail")
	}
}

func TestParseHeaderExtensionInvalid(t *testing.T) {
	md := &Metadata{}
	parseHeaderExtension(nil, md, 0)
	parseHeaderExtension(make([]byte, 22), md, 0)
	payload := make([]byte, 22)
	payload[18] = 1
	parseHeaderExtension(payload, md, 0)
}
