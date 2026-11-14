package importer

import (
	"github.com/daregod/ext-tags/internal/formats/mp4"
)

// ImportMP4 writes fields and the front cover into MP4/M4A.
func ImportMP4(path string, tf *TagFile, opts Options) error {
	if opts.Backup {
		if err := copyFile(path, path+".bak"); err != nil {
			return err
		}
	}
	var cover *mp4.Cover
	if tf.Cover != nil {
		cover = &mp4.Cover{Filename: tf.Cover.Filename, MIME: tf.Cover.MIME, Data: tf.Cover.Data}
	} else if tf.Covers != nil {
		if c := tf.Covers["front"]; c != nil {
			cover = &mp4.Cover{Filename: c.Filename, MIME: c.MIME, Data: c.Data}
		}
	}
	if len(opts.Only) > 0 {
		return mp4.WriteOnly(path, tf.Fields, cover, opts.Replace, opts.Only)
	}
	return mp4.Write(path, tf.Fields, cover, opts.Replace)
}
