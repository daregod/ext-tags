package matroska

import (
	"os"
	"path/filepath"
	"testing"
)

func copyMKV(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "formats", name))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestIndexHelpers(t *testing.T) {
	attachments := []Attachment{
		{Filename: "cover.jpg", MIME: "image/jpeg", Data: []byte{1}},
		{Filename: "back.jpg", MIME: "image/jpeg", Data: []byte{2}},
		{Filename: "icon.png", MIME: "image/png", Data: []byte{3}},
	}
	if got := FrontIndex(attachments); got != 0 {
		t.Fatalf("front = %d", got)
	}
	if got := AttachmentIndex(attachments, "back"); got != 1 {
		t.Fatalf("back = %d", got)
	}
	if got := AttachmentIndex(attachments, "icon"); got != 2 {
		t.Fatalf("icon = %d", got)
	}
	if got := AttachmentIndex(attachments, "disc"); got != -1 {
		t.Fatalf("disc = %d", got)
	}
	if got := FrontIndex(nil); got != -1 {
		t.Fatalf("empty front = %d", got)
	}
	if got := FrontIndex([]Attachment{{Filename: "x.jpg", MIME: "text/plain", Data: []byte{1}}}); got != -1 {
		t.Fatalf("non-image front = %d", got)
	}
}

func TestWriteKeepAndWithAttachments(t *testing.T) {
	path := copyMKV(t, "audio-cover.png.mkv")
	before, err := ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(before.Attachments) == 0 {
		t.Fatal("fixture has no attachments")
	}
	if err := WriteKeepAttachments(path, map[string]string{"Title": "Keep"}); err != nil {
		t.Fatal(err)
	}
	after, err := ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Attachments) != len(before.Attachments) {
		t.Fatalf("keep attachments = %d", len(after.Attachments))
	}
	if err := WriteWithAttachments(path, map[string]string{"Title": "Replace"}, []Attachment{{Filename: "new.jpg", MIME: "image/jpeg", Data: []byte{9}}}); err != nil {
		t.Fatal(err)
	}
	after, err = ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Attachments) != 1 || after.Attachments[0].Filename != "new.jpg" {
		t.Fatalf("attachments = %+v", after.Attachments)
	}
	if err := WriteWithAttachments(path, map[string]string{"Title": "Drop"}, []Attachment{}); err != nil {
		t.Fatal(err)
	}
	after, err = ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Attachments) != 0 {
		t.Fatalf("attachments = %+v", after.Attachments)
	}
}

func TestBuildAttachmentListDefaults(t *testing.T) {
	data, ok := buildAttachmentList([]Attachment{{Data: []byte{1}}, {Filename: "x", Data: []byte{2}}})
	if !ok || len(data) == 0 {
		t.Fatal("buildAttachmentList")
	}
	if data, ok := buildAttachmentList(nil); ok || data != nil {
		t.Fatal("empty buildAttachmentList")
	}
}
