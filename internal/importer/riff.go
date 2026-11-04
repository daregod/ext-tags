package importer

import (
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/riff"
)

// ImportRIFF writes ID3v2 metadata into WAV/RF64/Wave64/AIFF.
func ImportRIFF(path string, tf *TagFile, opts Options) error {
	if opts.Backup {
		if err := copyFile(path, path+".bak"); err != nil {
			return err
		}
	}
	doc := metadata.New()
	for key, value := range tf.Fields {
		doc.Fields[key] = value
	}
	for kind, cover := range tf.Covers {
		doc.Artwork[kind] = cover
	}
	doc.Cover = tf.Cover
	return riff.Write(path, doc)
}
