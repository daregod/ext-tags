package id3

import "testing"

func TestBuild(t *testing.T) {
	covers := map[string]*Cover{"front": {Filename: "cover.jpg", MIME: "image/jpeg", Data: []byte{1}}}
	for _, version := range []int{2, 3, 4} {
		data, err := Build(map[string]string{
			"Title": "Title", "Comment": "Comment", "Catalog#": "CAT-1",
		}, covers, version)
		if err != nil {
			t.Fatal(err)
		}
		frames, gotVersion, _, has, err := parseTag(data)
		if err != nil || !has || gotVersion != version {
			t.Fatalf("v%d: version=%d has=%v err=%v", version, gotVersion, has, err)
		}
		ids := map[string]bool{}
		for _, frame := range frames {
			ids[frame.id] = true
		}
		if !ids[commentID(version)] || !ids[txxxID(version)] || !ids[apicID(version)] {
			t.Fatalf("v%d ids = %v", version, ids)
		}
	}
}
