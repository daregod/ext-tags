package extractor

import "testing"

func TestMapKey(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
		keep bool
	}{
		{"mp4 title", "\xa9nam", "Title", true},
		{"mp4 artist", "\xa9ART", "Artist", true},
		{"mp4 album", "\xa9alb", "Album", true},
		{"mp4 composer", "\xa9wrt", "Composer", true},
		{"mp4 album artist", "aART", "Album artist", true},
		{"mp4 compilation", "cpil", "Itunescompilation", true},
		{"mp4 tempo", "tmpo", "Tempo", true},
		{"id3 title", "TIT2", "Title", true},
		{"vorbis title", "TITLE", "Title", true},
		{"custom field", "MYFIELD", "MYFIELD", true},
		{"track number", "TRACKNUMBER", "", false},
		{"total tracks", "TOTALTRACKS", "", false},
		{"disc number", "DISCNUMBER", "", false},
		{"vorbis vendor", "vendor", "", false},
		{"itunnorm", "iTunNORM", "", false},
		{"itunsmpb", "iTunSMPB", "", false},
		{"encoding params", "Encoding Params", "", false},
		{"mp4 tool", "\xa9too", "", false},
		{"unknown atom", "\xa9xyz", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, keep := mapKey(c.in)
			if keep != c.keep {
				t.Fatalf("mapKey(%q) keep = %v, want %v", c.in, keep, c.keep)
			}
			if got != c.want {
				t.Fatalf("mapKey(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestGetTrack(t *testing.T) {
	raw := map[string]interface{}{
		"TRACKNUMBER": "02",
		"TRACKTOTAL":  "16",
	}
	num, total := getTrack(raw)
	if num != "02" || total != "16" {
		t.Fatalf("getTrack = %q/%q, want 02/16", num, total)
	}
}

func TestGetDisc(t *testing.T) {
	raw := map[string]interface{}{
		"DISCNUMBER": "1/2",
	}
	num, total := getDisc(raw)
	if num != "1" || total != "2" {
		t.Fatalf("getDisc = %q/%q, want 1/2", num, total)
	}
}
