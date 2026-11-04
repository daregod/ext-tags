package mp4

import (
	"testing"
)

func TestMakeHdlr(t *testing.T) {
	boxes, err := parseBoxes(makeHdlr())
	if err != nil {
		t.Fatal(err)
	}
	if len(boxes) != 1 || boxes[0].typ != "hdlr" {
		t.Fatalf("boxes = %+v", boxes)
	}
}

func TestShiftOffsets(t *testing.T) {
	stco := make([]byte, 16)
	put32(stco[4:8], 1)
	put32(stco[8:12], 200)
	box := makeBox("stco", stco)
	if err := shiftOffsets(box, 100, 10); err != nil {
		t.Fatal(err)
	}
	parsed, _ := parseBoxes(box)
	if got := read32(parsed[0].data[8:12]); got != 210 {
		t.Fatalf("stco = %d", got)
	}

	co64 := make([]byte, 20)
	put32(co64[4:8], 1)
	put64(co64[8:16], 300)
	box = makeBox("co64", co64)
	if err := shiftOffsets(box, 100, 10); err != nil {
		t.Fatal(err)
	}
	parsed, _ = parseBoxes(box)
	if got := read64(parsed[0].data[8:16]); got != 310 {
		t.Fatalf("co64 = %d", got)
	}
}

func TestRewriteMoovCreatesMetadata(t *testing.T) {
	fields := map[string]string{"Title": "T", "Track": "3/12"}
	cover := &Cover{Filename: "cover.jpg", MIME: "image/jpeg", Data: []byte{1}}
	payload, err := rewriteMoov(nil, fields, cover, true)
	if err != nil {
		t.Fatal(err)
	}
	boxes, err := parseBoxes(payload)
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	var walk func([]byte)
	walk = func(data []byte) {
		children, err := parseBoxes(data)
		if err != nil {
			return
		}
		for _, child := range children {
			found[child.typ] = true
			if child.typ == "udta" || child.typ == "meta" || child.typ == "ilst" {
				payload := child.data
				if child.typ == "meta" && len(payload) >= 4 {
					payload = payload[4:]
				}
				walk(payload)
			}
		}
	}
	for _, b := range boxes {
		found[b.typ] = true
		if b.typ == "udta" {
			walk(b.data)
		}
	}
	if !found["udta"] || !found["meta"] || !found["ilst"] || !found["hdlr"] {
		t.Fatalf("found = %v", found)
	}
}

func put32(dst []byte, v uint32) {
	dst[0], dst[1], dst[2], dst[3] = byte(v>>24), byte(v>>16), byte(v>>8), byte(v)
}
func put64(dst []byte, v uint64) { put32(dst[0:4], uint32(v>>32)); put32(dst[4:8], uint32(v)) }
func read32(b []byte) uint32 {
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}
func read64(b []byte) uint64 { return uint64(read32(b[0:4]))<<32 | uint64(read32(b[4:8])) }
func TestParseBoxesErrors(t *testing.T) {
	for _, data := range [][]byte{
		{0, 0, 0},
		{0, 0, 0, 8, 'f', 't', 'y', 'p', 1},
		{0, 0, 0, 1, 'f', 't', 'y', 'p'},
	} {
		if _, err := parseBoxes(data); err == nil {
			t.Fatalf("data % X must fail", data)
		}
	}
}

func TestShiftOffsetsErrors(t *testing.T) {
	stco := make([]byte, 8)
	put32(stco[4:8], 2)
	if err := shiftOffsets(makeBox("stco", stco), 0, 1); err == nil {
		t.Fatal("truncated stco must fail")
	}
	co64 := make([]byte, 8)
	put32(co64[4:8], 1)
	if err := shiftOffsets(makeBox("co64", co64), 0, 1); err == nil {
		t.Fatal("truncated co64 must fail")
	}
}

func TestSplitPair(t *testing.T) {
	if a, b := splitPair("3/12"); a != "3" || b != "12" {
		t.Fatalf("split = %q/%q", a, b)
	}
	if a, b := splitPair("3"); a != "3" || b != "" {
		t.Fatalf("split = %q/%q", a, b)
	}
}
