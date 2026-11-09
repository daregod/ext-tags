package album

import "testing"

func TestMergePreservesExistingValues(t *testing.T) {
	existing := &Config{
		Album:   map[string]string{"Album": "Manual", "Genre": ""},
		Tracks:  map[string]map[string]string{"1": {"Comment": ""}},
		Artwork: map[string]string{"front": "manual.jpg"},
	}
	generated := &Config{
		Album:   map[string]string{"Album": "Generated", "Title": "Generated Title"},
		Tracks:  map[string]map[string]string{"1": {"Comment": "x", "Title": "Track"}},
		Artwork: map[string]string{"front": "generated.jpg", "back": "back.jpg"},
	}
	got := Merge(existing, generated)
	if got.Album["Album"] != "Manual" || got.Album["Title"] != "Generated Title" {
		t.Fatalf("album = %#v", got.Album)
	}
	if _, ok := got.Album["Genre"]; !ok || got.Album["Genre"] != "" {
		t.Fatal("explicit empty album field was lost")
	}
	if got.Tracks["1"]["Comment"] != "" || got.Tracks["1"]["Title"] != "Track" {
		t.Fatalf("track = %#v", got.Tracks["1"])
	}
	if got.Artwork["front"] != "manual.jpg" || got.Artwork["back"] != "back.jpg" {
		t.Fatalf("artwork = %#v", got.Artwork)
	}
}

func TestMergeMatchesNumericTrackKeys(t *testing.T) {
	existing := &Config{Tracks: map[string]map[string]string{"01": {"Title": "Old"}}}
	generated := &Config{Tracks: map[string]map[string]string{"1": {"Title": "New", "Track": "01/12"}}}
	got := Merge(existing, generated)
	if len(got.Tracks) != 1 || got.Tracks["01"]["Title"] != "Old" || got.Tracks["01"]["Track"] != "01/12" {
		t.Fatalf("tracks = %#v", got.Tracks)
	}
}
