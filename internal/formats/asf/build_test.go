package asf

import "testing"

func TestBuildExtendedContentMerge(t *testing.T) {
	existing := []Descriptor{
		{Name: "WM/AlbumTitle", Type: 0, Value: utf16Z("old")},
		{Name: "WM/Picture", Type: 1, Value: []byte{3}},
		{Name: "WM/Genre", Type: 0, Value: utf16Z("keep")},
	}
	covers := map[string]*Cover{"front": {Filename: "cover.jpg", MIME: "image/jpeg", Data: []byte{1}}}
	data := buildExtendedContent(map[string]string{"Album": "new"}, covers, existing, false)
	if len(data) == 0 {
		t.Fatal("empty extended content")
	}
	data = buildExtendedContent(nil, nil, existing, true)
	if len(data) != 0 {
		t.Fatal("replace with no fields must be empty")
	}
}
