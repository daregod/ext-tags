// Command genfixtures generates tiny test fixtures for every audio format
// ext-tags knows about. It replaces scripts/gen-fixtures.ps1 and works on
// Windows, Linux and macOS as long as ffmpeg (and optionally flac/mp2enc)
// is installed.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type generator struct {
	out   string
	force bool
}

func main() {
	out := flag.String("out", "testdata/formats", "output directory")
	force := flag.Bool("force", false, "regenerate existing files")
	flag.Parse()

	g := &generator{out: *out, force: *force}
	if err := os.MkdirAll(g.out, 0o755); err != nil {
		fail(err)
	}
	g.cleanup()
	if err := g.run(); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "genfixtures:", err)
	os.Exit(1)
}

func (g *generator) path(name string) string { return filepath.Join(g.out, name) }

func (g *generator) cleanup() {
	for _, pattern := range []string{"*.tag", "*.tag.bak"} {
		matches, _ := filepath.Glob(filepath.Join(g.out, pattern))
		for _, p := range matches {
			_ = os.Remove(p)
		}
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (g *generator) run() error {
	cover := g.path("cover.jpg")
	folder := g.path("folder.jpg")
	poster := g.path("poster.jpg")
	box := g.path("box.png")

	for _, img := range []struct {
		path  string
		color string
	}{
		{cover, "red"}, {folder, "green"}, {box, "yellow"}, {poster, "blue"},
	} {
		if err := g.ffmpegRaw("-f", "lavfi", "-i",
			fmt.Sprintf("color=c=%s:s=32x32:d=1", img.color),
			"-frames:v", "1", img.path); err != nil {
			return err
		}
	}

	meta := []string{
		"-metadata", "title=Fixture Title",
		"-metadata", "artist=Fixture Artist",
		"-metadata", "album=Fixture Album",
		"-metadata", "genre=Test",
		"-metadata", "date=2026",
		"-metadata", "track=3",
	}
	audio := []string{"-f", "lavfi", "-i",
		"sine=frequency=440:duration=0.4:sample_rate=22050", "-ac", "1"}
	audioCover := []string{"-i", cover, "-map", "0:a", "-map", "1:v",
		"-c:v", "copy", "-disposition:v", "attached_pic"}

	steps := []struct {
		name string
		args []string
	}{
		{"audio.mp3", join(audio, audioCover, []string{"-c:a", "libmp3lame", "-b:a", "48k", "-id3v2_version", "3"}, meta, []string{g.path("audio.mp3")})},
		{"audio-id3v2.4.mp3", join(audio, audioCover, []string{"-c:a", "libmp3lame", "-b:a", "48k", "-id3v2_version", "4"}, meta, []string{g.path("audio-id3v2.4.mp3")})},
		{"audio-id3v1.mp3", nil}, // generated below
		{"audio-raw.mp2", join(audio, []string{"-c:a", "mp2", "-b:a", "64k", "-f", "mp2"}, []string{g.path("audio-raw.mp2")})},
		{"audio-raw.aac", join(audio, []string{"-c:a", "aac", "-b:a", "32k", "-f", "adts"}, []string{g.path("audio-raw.aac")})},
		{"audio.flac", join(audio, audioCover, []string{"-c:a", "flac"}, meta, []string{g.path("audio.flac")})},
		{"audio.ogg", join(audio, []string{"-c:a", "libvorbis", "-q:a", "0"}, meta, []string{g.path("audio.ogg")})},
		{"audio.opus", join(audio, []string{"-c:a", "libopus", "-b:a", "24k"}, meta, []string{g.path("audio.opus")})},
		{"audio-aac.m4a", join(audio, audioCover, []string{"-c:a", "aac", "-b:a", "32k"}, meta, []string{g.path("audio-aac.m4a")})},
		{"audio-alac.m4a", join(audio, audioCover, []string{"-c:a", "alac"}, meta, []string{g.path("audio-alac.m4a")})},
		{"audio.wav", join(audio, []string{"-c:a", "pcm_s16le"}, meta, []string{g.path("audio.wav")})},
		{"audio-empty.wav", join(audio, []string{"-fflags", "+bitexact", "-flags:a", "+bitexact", "-c:a", "pcm_s16le"}, []string{g.path("audio-empty.wav")})},
		{"audio.aiff", join(audio, audioCover, []string{"-c:a", "pcm_s16be", "-write_id3v2", "1", "-id3v2_version", "3"}, meta, []string{g.path("audio.aiff")})},
		{"audio-unicode.flac", join(audio, audioCover, []string{"-c:a", "flac"}, []string{"-metadata", "title=Тестовый заголовок", "-metadata", "artist=Исполнитель", "-metadata", "album=Альбом", "-metadata", "genre=Тест", "-metadata", "date=2026", "-metadata", "track=3"}, []string{g.path("audio-unicode.flac")})},
		{"audio.wv", join(audio, []string{"-c:a", "wavpack"}, meta, []string{g.path("audio.wv")})},
		{"audio.wma", join(audio, []string{"-c:a", "wmav2", "-b:a", "32k", "-f", "asf"}, meta, []string{g.path("audio.wma")})},
		{"audio.mkv", join(audio, []string{"-c:a", "flac"}, meta, []string{"-attach", cover, "-metadata:s:t", "mimetype=image/jpeg", "-metadata:s:t", "filename=cover.jpg", g.path("audio.mkv")})},
		{"audio-track.mkv", join(audio, []string{"-c:a", "flac"}, []string{"-metadata", "title=Album Title", "-metadata", "album=Some Album", "-metadata", "genre=Test", "-metadata", "date=2026", "-metadata", "track=3", "-metadata:s:a:0", "ARTIST=Track Artist"}, []string{g.path("audio-track.mkv")})},
		{"audio-both.mkv", join(audio, []string{"-c:a", "flac"}, []string{"-metadata", "title=Album Title", "-metadata", "ALBUM=Some Album", "-metadata", "ARTIST=Album Artist", "-metadata", "genre=Test", "-metadata", "date=2026", "-metadata", "track=3", "-metadata:s:a:0", "ARTIST=Track Artist"}, []string{g.path("audio-both.mkv")})},
		{"audio-attachments.mkv", join(audio, []string{"-c:a", "flac"}, meta, []string{"-attach", poster, "-metadata:s:t:0", "mimetype=image/jpeg", "-metadata:s:t:0", "filename=poster.jpg", "-attach", folder, "-metadata:s:t:1", "mimetype=image/jpeg", "-metadata:s:t:1", "filename=folder.jpg", g.path("audio-attachments.mkv")})},
		{"audio-cover.png.mkv", join(audio, []string{"-c:a", "flac"}, meta, []string{"-attach", box, "-metadata:s:t", "mimetype=image/png", "-metadata:s:t", "filename=box.png", g.path("audio-cover.png.mkv")})},
		{"audio.webm", join(audio, []string{"-c:a", "libopus", "-b:a", "24k"}, meta, []string{g.path("audio.webm")})},
		{"audio-flac.oga", join(audio, []string{"-c:a", "flac", "-f", "ogg"}, meta, []string{g.path("audio-flac.oga")})},
		{"audio.spx", join(audio, []string{"-c:a", "libspeex", "-f", "ogg"}, meta, []string{g.path("audio.spx")})},
		{"audio.ac3", join(audio, []string{"-c:a", "ac3", "-b:a", "64k", "-f", "ac3"}, []string{g.path("audio.ac3")})},
		{"audio.dts", join(audio, []string{"-c:a", "dca", "-strict", "-2", "-f", "dts"}, []string{g.path("audio.dts")})},
		{"audio.au", join(audio, []string{"-c:a", "pcm_s16be", "-f", "au"}, []string{g.path("audio.au")})},
		{"audio.rf64", join(audio, []string{"-c:a", "pcm_s16le", "-f", "wav", "-rf64", "always"}, meta, []string{g.path("audio.rf64")})},
		{"audio.ts", join(audio, []string{"-c:a", "mp2", "-b:a", "64k", "-f", "mpegts"}, meta, []string{g.path("audio.ts")})},
	}

	for _, step := range steps {
		if step.args == nil {
			continue
		}
		if err := g.ffmpeg(step.name, step.args); err != nil {
			return err
		}
	}

	if err := g.id3v1(); err != nil {
		return err
	}
	if err := g.prependID3("audio-raw.mp2", "audio.mp2"); err != nil {
		return err
	}
	if err := g.prependID3("audio-raw.aac", "audio.aac"); err != nil {
		return err
	}
	if err := g.mp1(); err != nil {
		return err
	}
	if err := g.wave64(); err != nil {
		return err
	}

	video := []string{"-f", "lavfi", "-i", "color=c=black:s=32x32:d=0.4",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=0.4:sample_rate=22050"}
	return g.ffmpeg("video.wmv", join(video,
		[]string{"-c:v", "wmv2", "-c:a", "wmav2", "-b:a", "32k", "-f", "asf"},
		meta, []string{g.path("video.wmv")}))
}

func join(parts ...[]string) []string {
	var out []string
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

func (g *generator) ffmpeg(name string, args []string) error {
	path := g.path(name)
	if !g.force && exists(path) {
		fmt.Printf("%-18s    skip\n", name)
		return nil
	}
	if err := g.ffmpegRaw(args...); err != nil {
		return fmt.Errorf("ffmpeg failed for %s: %w", name, err)
	}
	if !exists(path) {
		return fmt.Errorf("ffmpeg produced no file %s", name)
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	fmt.Printf("%-18s %7d B\n", name, info.Size())
	return nil
}

func (g *generator) ffmpegRaw(args ...string) error {
	cmd := exec.Command("ffmpeg", append([]string{"-y", "-hide_banner", "-loglevel", "error"}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (g *generator) id3v1() error {
	src := g.path("audio-id3v1-src.mp3")
	args := []string{"-f", "lavfi", "-i", "sine=frequency=440:duration=0.4:sample_rate=22050", "-ac", "1",
		"-c:a", "libmp3lame", "-b:a", "48k", "-write_id3v1", "1",
		"-metadata", "title=Fixture Title", "-metadata", "artist=Fixture Artist",
		"-metadata", "album=Fixture Album", "-metadata", "genre=Test",
		"-metadata", "date=2026", "-metadata", "track=3", src}
	if err := g.ffmpegRaw(args...); err != nil {
		return err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	tagLen := id3Size(data)
	if 10+tagLen > len(data) {
		return fmt.Errorf("audio-id3v1: bad ID3 size")
	}
	if err := os.WriteFile(g.path("audio-id3v1.mp3"), data[10+tagLen:], 0o644); err != nil {
		return err
	}
	_ = os.Remove(src)
	info, _ := os.Stat(g.path("audio-id3v1.mp3"))
	fmt.Printf("%-18s %7d B\n", "audio-id3v1.mp3", info.Size())
	return nil
}

func id3Size(data []byte) int {
	if len(data) < 10 || string(data[:3]) != "ID3" {
		return 0
	}
	return int(data[6]&0x7f)<<21 | int(data[7]&0x7f)<<14 | int(data[8]&0x7f)<<7 | int(data[9]&0x7f)
}

func (g *generator) prependID3(rawName, outName string) error {
	mp3, err := os.ReadFile(g.path("audio.mp3"))
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(g.path(rawName))
	if err != nil {
		return err
	}
	tagLen := id3Size(mp3)
	if 10+tagLen > len(mp3) {
		return fmt.Errorf("%s: bad ID3 size", "audio.mp3")
	}
	out := append(append([]byte{}, mp3[:10+tagLen]...), raw...)
	if err := os.WriteFile(g.path(outName), out, 0o644); err != nil {
		return err
	}
	_ = os.Remove(g.path(rawName))
	fmt.Printf("%-18s %7d B\n", outName, len(out))
	return nil
}

func (g *generator) mp1() error {
	path := g.path("audio.mp1")
	if !g.force && exists(path) {
		return nil
	}
	mp2enc, ok := findMP2Enc()
	if !ok {
		fmt.Println("audio.mp1          skip (mp2enc not found)")
		return nil
	}
	wav := g.path("audio-mp1-src.wav")
	raw := g.path("audio-mp1-raw.mp1")
	if err := g.ffmpegRaw("-f", "lavfi", "-i",
		"sine=frequency=440:duration=0.4:sample_rate=44100", "-ac", "2",
		"-c:a", "pcm_s16le", wav); err != nil {
		return err
	}
	in, err := os.Open(wav)
	if err != nil {
		return err
	}
	cmd := exec.Command(mp2enc, "-l", "1", "-b", "384", "-r", "44100", "-o", raw)
	cmd.Stdin = in
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	in.Close()
	if err != nil || !exists(raw) {
		_ = os.Remove(wav)
		_ = os.Remove(raw)
		fmt.Println("audio.mp1          skip (mp2enc failed)")
		return nil
	}
	if err := g.prependID3File(raw, path); err != nil {
		return err
	}
	_ = os.Remove(wav)
	_ = os.Remove(raw)
	info, _ := os.Stat(path)
	fmt.Printf("%-18s %7d B\n", "audio.mp1", info.Size())
	return nil
}

func (g *generator) prependID3File(rawPath, outPath string) error {
	mp3, err := os.ReadFile(g.path("audio.mp3"))
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(rawPath)
	if err != nil {
		return err
	}
	tagLen := id3Size(mp3)
	out := append(append([]byte{}, mp3[:10+tagLen]...), raw...)
	return os.WriteFile(outPath, out, 0o644)
}

func findMP2Enc() (string, bool) {
	for _, name := range []string{
		filepath.Join("bin", "mjpegtools-2.1.0-mingw-bin64", "bin", "mp2enc.exe"),
		filepath.Join("bin", "mjpegtools-2.1.0-mingw-bin64", "bin", "mp2enc"),
	} {
		if exists(name) {
			return name, true
		}
	}
	if p, err := exec.LookPath("mp2enc"); err == nil {
		return p, true
	}
	return "", false
}

func (g *generator) wave64() error {
	path := g.path("audio.w64")
	if !g.force && exists(path) {
		return nil
	}
	flac, err := exec.LookPath("flac")
	if err != nil {
		fmt.Println("audio.w64          skip (flac not found)")
		return nil
	}
	if !exists(g.path("audio.wav")) {
		if err := g.ffmpeg("audio.wav", join([]string{"-f", "lavfi", "-i",
			"sine=frequency=440:duration=0.4:sample_rate=22050", "-ac", "1"},
			[]string{"-c:a", "pcm_s16le"}, []string{g.path("audio.wav")})); err != nil {
			return err
		}
	}
	tmpFlac := g.path("audio-w64.tmp.flac")
	raw := g.path("audio-w64.tmp.w64")
	if err := exec.Command(flac, "-s", "-f", "-o", tmpFlac, g.path("audio.wav")).Run(); err != nil {
		_ = os.Remove(tmpFlac)
		fmt.Println("audio.w64          skip (flac failed)")
		return nil
	}
	if err := exec.Command(flac, "-s", "-d", "-f", "--force-wave64-format", "-o", raw, tmpFlac).Run(); err != nil {
		_ = os.Remove(tmpFlac)
		_ = os.Remove(raw)
		fmt.Println("audio.w64          skip (flac failed)")
		return nil
	}
	base, err := os.ReadFile(raw)
	if err != nil {
		return err
	}
	mp3, err := os.ReadFile(g.path("audio.mp3"))
	if err != nil {
		return err
	}
	id3 := mp3[:10+id3Size(mp3)]
	if len(base) < 64 {
		return fmt.Errorf("audio.w64: short file")
	}
	fmtSize := binary.LittleEndian.Uint64(base[56:64])
	insertAt := 40 + int(fmtSize)
	guid := append([]byte("ID3 "), 0xF3, 0xAC, 0xD3, 0x11, 0x8C, 0xD1, 0x00, 0xC0, 0x4F, 0x8E, 0xDB, 0x8A)
	chunkSize := 24 + len(id3)
	chunk := make([]byte, (chunkSize+7)&^7)
	copy(chunk, guid)
	binary.LittleEndian.PutUint64(chunk[16:24], uint64(chunkSize))
	copy(chunk[24:], id3)
	out := make([]byte, 0, len(base)+len(chunk))
	out = append(out, base[:insertAt]...)
	out = append(out, chunk...)
	out = append(out, base[insertAt:]...)
	binary.LittleEndian.PutUint64(out[16:24], uint64(len(out)))
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return err
	}
	_ = os.Remove(tmpFlac)
	_ = os.Remove(raw)
	info, _ := os.Stat(path)
	fmt.Printf("%-18s %7d B\n", "audio.w64", info.Size())
	return nil
}
