package registry

import (
	"testing"

	"github.com/daregod/ext-tags/internal/config"
)

func TestDefaultRegistry(t *testing.T) {
	reg := NewDefault()
	for _, ext := range config.Default().Files.Extensions {
		if _, ok := reg.ReaderFor("track" + ext); !ok {
			t.Errorf("no reader for %s", ext)
		}
	}
	wantReaders := map[string]string{
		".mp3": "mp3", ".flac": "flac", ".ogg": "ogg", ".opus": "ogg",
		".ape": "apev2", ".wv": "apev2", ".m4a": "mp4", ".alac": "mp4",
		".mkv": "matroska", ".wma": "asf", ".wav": "riff", ".aac": "aac",
	}
	for ext, name := range wantReaders {
		reader, ok := reg.ReaderFor("track" + ext)
		if !ok || reader.Name() != name {
			t.Errorf("%s: reader=%v ok=%v, want %s", ext, reader, ok, name)
		}
	}
	wantWriters := map[string]string{
		".mp3": "mp3", ".mp2": "mp3", ".mp1": "mp3",
		".flac": "flac", ".ogg": "ogg", ".oga": "ogg", ".opus": "ogg", ".spx": "ogg",
		".ape": "apev2", ".wv": "apev2", ".mpc": "apev2", ".tak": "apev2",
		".ac3": "apev2", ".dts": "apev2",
		".m4a": "mp4", ".m4b": "mp4", ".mp4": "mp4", ".alac": "mp4",
		".mkv": "matroska", ".webm": "matroska",
		".wma": "asf", ".wmv": "asf", ".asf": "asf",
	}
	for ext, name := range wantWriters {
		writer, ok := reg.WriterFor("track" + ext)
		if !ok || writer.Name() != name {
			t.Errorf("%s: writer=%v ok=%v, want %s", ext, writer, ok, name)
		}
	}
	if _, ok := reg.WriterFor("track.txt"); ok {
		t.Fatal("unexpected writer for .txt")
	}
}
