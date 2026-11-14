package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/formats/asf"
	"github.com/daregod/ext-tags/internal/formats/matroska"
	"github.com/daregod/ext-tags/internal/formats/mp4"
	"github.com/daregod/ext-tags/internal/formats/ogg"
	"github.com/daregod/ext-tags/internal/formats/riff"
	"github.com/daregod/ext-tags/internal/formats/vorbis"
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/registry"
)

func copyFormatFixture(t *testing.T, name, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "formats", name))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestContainerPicturesFormats(t *testing.T) {
	dir := t.TempDir()
	files := []string{"audio.mp3", "audio.flac", "audio.aiff", "audio.w64", "audio-cover.png.mkv"}
	for _, name := range files {
		t.Run(name, func(t *testing.T) {
			path := copyFormatFixture(t, name, dir)
			if len(containerPictures(path, metadata.New())) == 0 {
				t.Fatalf("no pictures for %s", path)
			}
		})
	}

	t.Run("ogg", func(t *testing.T) {
		path := copyFormatFixture(t, "audio.ogg", dir)
		if err := ogg.WriteComments(path, nil, []vorbis.Picture{{Type: 3, MIME: "image/jpeg", Data: []byte{1}}}, true); err != nil {
			t.Fatal(err)
		}
		if len(containerPictures(path, metadata.New())) == 0 {
			t.Fatal("no ogg picture")
		}
	})
	t.Run("asf", func(t *testing.T) {
		path := copyFormatFixture(t, "audio.wma", dir)
		if err := asf.Write(path, nil, map[string]*asf.Cover{"front": {Filename: "front.jpg", MIME: "image/jpeg", Data: []byte{1}}}, false); err != nil {
			t.Fatal(err)
		}
		if len(containerPictures(path, metadata.New())) == 0 {
			t.Fatal("no asf picture")
		}
	})
	t.Run("mp4", func(t *testing.T) {
		path := copyFormatFixture(t, "audio-aac.m4a", dir)
		if err := mp4.Write(path, nil, &mp4.Cover{Filename: "front.jpg", MIME: "image/jpeg", Data: []byte{1}}, false); err != nil {
			t.Fatal(err)
		}
		if len(containerPictures(path, metadata.New())) == 0 {
			t.Fatal("no mp4 picture")
		}
	})
	for _, name := range []string{"audio.wav", "audio.rf64"} {
		t.Run(name, func(t *testing.T) {
			path := copyFormatFixture(t, name, dir)
			doc := metadata.New()
			doc.Artwork[metadata.Front] = &metadata.Cover{Filename: "front.jpg", MIME: "image/jpeg", Data: []byte{1}}
			if err := riff.Write(path, doc); err != nil {
				t.Fatal(err)
			}
			if len(containerPictures(path, metadata.New())) == 0 {
				t.Fatalf("no riff picture for %s", path)
			}
		})
	}
}

func TestIndexedHelpers(t *testing.T) {
	cases := []struct {
		in   byte
		want string
	}{{0, "front"}, {3, "front"}, {4, "back"}, {8, "artist"}, {6, "disc"}, {1, "icon"}, {2, ""}}
	for _, tc := range cases {
		if got := pictureSlot(tc.in); got != tc.want {
			t.Fatalf("pictureSlot(%d) = %q", tc.in, got)
		}
	}
	for _, tc := range []struct {
		in   uint32
		want string
	}{{3, "front"}, {4, "back"}, {8, "artist"}, {6, "disc"}, {1, "icon"}, {0, ""}} {
		if got := vorbisSlot(tc.in); got != tc.want {
			t.Fatalf("vorbisSlot(%d) = %q", tc.in, got)
		}
	}
	for key, want := range map[string]string{
		"Cover Art (Front)": "front", "Cover Art (Back)": "back", "Artist Picture": "artist",
		"Cover Art (Disc)": "disc", "Cover Art (Icon)": "icon", "other": "",
	} {
		if got := apeSlot(key); got != want {
			t.Fatalf("apeSlot(%q) = %q", key, got)
		}
	}
	for typ, want := range map[byte]string{3: "front", 4: "back", 8: "artist", 6: "disc", 1: "icon", 0: ""} {
		if got := asfSlot(typ); got != want {
			t.Fatalf("asfSlot(%d) = %q", typ, got)
		}
	}
	if got := matroskaSlot([]matroska.Attachment{{Filename: "x.jpg", MIME: "image/jpeg", Data: []byte{1}}}, 0); got != "front" {
		t.Fatalf("matroska slot = %q", got)
	}
	if got := matroskaSlot([]matroska.Attachment{{Filename: "other.jpg", MIME: "image/jpeg", Data: []byte{1}}}, 0); got != "front" {
		t.Fatalf("matroska unknown = %q", got)
	}
	if filename, data := splitPictureValue([]byte("x.jpg\x00abc")); filename != "x.jpg" || string(data) != "abc" {
		t.Fatalf("split = %q %q", filename, data)
	}
	if filename, data := splitPictureValue([]byte("raw")); filename != "" || data != nil {
		t.Fatalf("split raw = %q %q", filename, data)
	}
	for ext, want := range map[string]string{".png": "image/png", ".gif": "image/gif", ".webp": "image/webp", ".jpg": "image/jpeg", ".unknown": "image/jpeg"} {
		if got := coverMIME("x" + ext); got != want {
			t.Fatalf("mime(%s) = %q", ext, got)
		}
	}
}

func TestContainerPicturesFallback(t *testing.T) {
	doc := metadata.New()
	doc.Artwork[metadata.Front] = &metadata.Cover{Filename: "cover.jpg", MIME: "image/jpeg", Data: []byte{1}}
	got := containerPictures("track.mp4", doc)
	if len(got) != 1 || got[0].Slot != metadata.Front {
		t.Fatalf("fallback = %+v", got)
	}
	if got := containerPictures("track.mp4", metadata.New()); len(got) != 0 {
		t.Fatalf("empty fallback = %+v", got)
	}
	if got := containerPictures("track.unknown", metadata.New()); len(got) != 0 {
		t.Fatalf("unknown fallback = %+v", got)
	}
}

func TestLoadIndexedArtworkErrors(t *testing.T) {
	cfg := config.Default()
	if _, err := (&app{opts: &options{}, registry: registry.NewDefault()}).loadIndexedArtwork("track.xyz", cfg); err == nil {
		t.Fatal("expected unsupported format")
	}
	if _, err := (&app{opts: &options{}, registry: registry.NewDefault()}).loadIndexedArtwork(filepath.Join(t.TempDir(), "missing.flac"), cfg); err == nil {
		t.Fatal("expected read error")
	}
	path := copyFormatFixture(t, "audio-empty.wav", t.TempDir())
	items, err := (&app{opts: &options{}, registry: registry.NewDefault()}).loadIndexedArtwork(path, cfg)
	if err != nil || len(items) != 0 {
		t.Fatalf("empty = %+v %v", items, err)
	}
}

func TestParsePictureIndex(t *testing.T) {
	for _, value := range []string{"container:0", "tag:2"} {
		if _, _, ok := parsePictureIndex(value); !ok {
			t.Fatalf("parse %q", value)
		}
	}
	for _, value := range []string{"", "container", "tag:x", "other:0", "tag:-1"} {
		if _, _, ok := parsePictureIndex(value); ok {
			t.Fatalf("parse invalid %q", value)
		}
	}
}

func TestRiffID3Branches(t *testing.T) {
	for _, name := range []string{"audio.wav", "audio.rf64", "audio.w64", "audio.aiff"} {
		path := copyFormatFixture(t, name, t.TempDir())
		_ = riffID3(path)
	}
	if got := riffID3("track.unknown"); got != nil {
		t.Fatalf("unknown = %v", got)
	}
}

func TestContainerPicturesErrorsAndMatroskaSlot(t *testing.T) {
	dir := t.TempDir()
	badFlac := filepath.Join(dir, "bad.flac")
	if err := os.WriteFile(badFlac, []byte("not flac"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := containerPictures(badFlac, metadata.New()); len(got) != 0 {
		t.Fatalf("bad flac = %+v", got)
	}
	badMP3 := filepath.Join(dir, "bad.mp3")
	if err := os.WriteFile(badMP3, []byte("not id3"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := containerPictures(badMP3, metadata.New()); len(got) != 0 {
		t.Fatalf("bad mp3 = %+v", got)
	}
	attachments := []matroska.Attachment{
		{Filename: "cover.jpg", MIME: "image/jpeg", Data: []byte{1}},
		{Filename: "other.jpg", MIME: "image/jpeg", Data: []byte{2}},
	}
	if got := matroskaSlot(attachments, 1); got != "" {
		t.Fatalf("slot = %q", got)
	}
}
