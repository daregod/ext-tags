package importer

import "testing"

func TestSplitCover(t *testing.T) {
	name, data := splitCover([]byte("cover.jpg\x00\xFF\xD8"))
	if name != "cover.jpg" || len(data) != 2 {
		t.Fatalf("name=%q data=% X", name, data)
	}
	name, data = splitCover([]byte{1, 2})
	if name != "" || len(data) != 2 {
		t.Fatalf("name=%q data=% X", name, data)
	}
}
