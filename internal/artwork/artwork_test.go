package artwork

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name string, data []byte) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestFindPriority(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "folder.jpg", []byte("folder"))
	writeFile(t, dir, "cover.png", []byte("cover"))

	img, err := Find(dir, []string{"cover", "folder"})
	if err != nil {
		t.Fatal(err)
	}
	if img.Filename != "cover.png" || img.MIME != "image/png" || string(img.Data) != "cover" {
		t.Fatalf("image = %+v", img)
	}

	img, err = Find(dir, []string{"folder"})
	if err != nil {
		t.Fatal(err)
	}
	if img.Filename != "folder.jpg" || img.MIME != "image/jpeg" {
		t.Fatalf("image = %+v", img)
	}
}

func TestFindCaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "Cover.JPG", []byte("x"))

	img, err := Find(dir, []string{"cover"})
	if err != nil {
		t.Fatal(err)
	}
	if img.Filename != "Cover.JPG" {
		t.Fatalf("filename = %q", img.Filename)
	}
}

func TestFindMissing(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "track.mp3", []byte("x"))
	if _, err := Find(dir, nil); err != ErrNotFound {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestFindEmptyDir(t *testing.T) {
	if _, err := Find(t.TempDir(), []string{"cover"}); err != ErrNotFound {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
