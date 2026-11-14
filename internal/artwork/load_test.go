package artwork

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cover.png")
	if err := os.WriteFile(path, []byte{1, 2, 3}, 0644); err != nil {
		t.Fatal(err)
	}
	img, err := Load(path)
	if err != nil || img.MIME != "image/png" || img.Filename != "cover.png" || len(img.Data) != 3 {
		t.Fatalf("load = %+v %v", img, err)
	}
	if _, err := Load(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("expected missing error")
	}
}
