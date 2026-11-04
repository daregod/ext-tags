// Package importer writes tags from an external .tag file back into audio
// containers. MP3 (ID3v1/ID3v2.2-2.4) is supported first.
package importer

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/daregod/ext-tags/internal/ape"
	"github.com/daregod/ext-tags/internal/id3"
)

// Cover is an embedded picture read from a .tag file.
type Cover struct {
	Filename string
	MIME     string
	Data     []byte
}

// TagFile is the decoded content of a .tag file.
type TagFile struct {
	Fields map[string]string
	Cover  *Cover            // front cover alias
	Covers map[string]*Cover // front, back, artist, disc, icon
}

// ReadTagFile parses a .tag file produced by ext-tags or foo_external_tags.
func ReadTagFile(path string) (*TagFile, error) {
	parsed, err := ape.ReadFile(path)
	if err != nil {
		return nil, err
	}
	artKeys := map[string]string{
		"cover art (front)": "front",
		"cover art (back)":  "back",
		"artist picture":    "artist",
		"cover art (disc)":  "disc",
		"cover art (icon)":  "icon",
	}
	tf := &TagFile{Fields: make(map[string]string), Covers: make(map[string]*Cover)}
	for _, item := range parsed.Items {
		if item.Flags&2 != 0 {
			kind, ok := artKeys[strings.ToLower(item.Key)]
			if !ok {
				continue
			}
			filename, data := splitCover(item.Value)
			if len(data) == 0 {
				continue
			}
			cover := &Cover{Filename: filename, MIME: mimeFor(filename), Data: data}
			tf.Covers[kind] = cover
			if kind == "front" {
				tf.Cover = cover
			}
			continue
		}
		tf.Fields[item.Key] = string(item.Value)
	}
	return tf, nil
}

func splitCover(value []byte) (string, []byte) {
	if i := bytes.IndexByte(value, 0); i >= 0 {
		return string(value[:i]), value[i+1:]
	}
	return "", value
}

func mimeFor(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	}
	return "image/jpeg"
}

// Options controls the tag writer.
type Options struct {
	Version int  // ID3v2 major version: 2, 3 or 4 (default 3)
	V1      bool // also write an ID3v1 tag
	Replace bool // drop audio tag frames that are not in the .tag file
	Backup  bool // keep a copy of the audio file as <name>.bak
}

// ImportMP3 writes the fields and cover of tf into an MP3 file.
func ImportMP3(path string, tf *TagFile, opts Options) error {
	if opts.Backup {
		if err := copyFile(path, path+".bak"); err != nil {
			return err
		}
	}
	var cover *id3.Cover
	if tf.Cover != nil {
		cover = &id3.Cover{Filename: tf.Cover.Filename, MIME: tf.Cover.MIME, Data: tf.Cover.Data}
	}
	return id3.Write(path, tf.Fields, cover, id3.Options{
		Version: opts.Version,
		V1:      opts.V1,
		Replace: opts.Replace,
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
