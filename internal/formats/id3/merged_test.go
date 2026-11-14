package id3

import "testing"

func TestBuildMergedPreservesFrames(t *testing.T) {
	existing, err := Build(map[string]string{"Title": "Old", "Comment": "Keep", "Label": "Custom"}, map[string]*Cover{
		"front": {Filename: "front.jpg", MIME: "image/jpeg", Data: []byte{1}},
		"back":  {Filename: "back.jpg", MIME: "image/jpeg", Data: []byte{2}},
	}, 3)
	if err != nil {
		t.Fatal(err)
	}
	tag, err := BuildMerged(existing, map[string]string{"Title": "New"}, nil, 3)
	if err != nil {
		t.Fatal(err)
	}
	frames, _, _, _, err := parseTag(tag)
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, frame := range frames {
		ids[frame.id] = true
	}
	for _, id := range []string{"TIT2", "COMM", "TXXX", "APIC"} {
		if !ids[id] {
			t.Fatalf("missing frame %s: %v", id, ids)
		}
	}
	pictures := ReadPictures(tag)
	if len(pictures) != 2 {
		t.Fatalf("pictures = %d", len(pictures))
	}
}

func TestBuildMergedOnlyRemovesSelected(t *testing.T) {
	existing, err := Build(map[string]string{"Title": "Old", "Comment": "Remove", "Label": "Keep"}, map[string]*Cover{
		"front": {Filename: "front.jpg", MIME: "image/jpeg", Data: []byte{1}},
		"back":  {Filename: "back.jpg", MIME: "image/jpeg", Data: []byte{2}},
	}, 3)
	if err != nil {
		t.Fatal(err)
	}
	tag, err := BuildMergedOnly(existing, map[string]string{"Title": "New"}, nil, 3, []string{"Title", "Comment", "Cover Art (Back)"})
	if err != nil {
		t.Fatal(err)
	}
	frames, _, _, _, err := parseTag(tag)
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, frame := range frames {
		ids = append(ids, frame.id)
	}
	if len(ids) != 3 { // TIT2, TXXX(Label), APIC(front)
		t.Fatalf("frames = %v", ids)
	}
	pictures := ReadPictures(tag)
	if len(pictures) != 1 || pictures[0].Type != 3 || string(pictures[0].Data) != "\x01" {
		t.Fatalf("pictures = %+v", pictures)
	}
}

func TestBuildMergedFallbacks(t *testing.T) {
	tag, err := BuildMerged(nil, map[string]string{"Title": "New"}, nil, 3)
	if err != nil || len(tag) == 0 {
		t.Fatalf("nil existing: %v %v", tag, err)
	}
	tag, err = BuildMerged([]byte("bad"), map[string]string{"Title": "New"}, nil, 3)
	if err != nil || len(tag) == 0 {
		t.Fatalf("bad existing: %v %v", tag, err)
	}
}
