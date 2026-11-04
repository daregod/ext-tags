package importer

import (
	"github.com/daregod/ext-tags/internal/matroska"
)

// ImportMatroska writes fields and artwork attachments into MKV/WebM.
func ImportMatroska(path string, tf *TagFile, opts Options) error {
	if opts.Backup {
		if err := copyFile(path, path+".bak"); err != nil {
			return err
		}
	}
	covers := make(map[string]*matroska.Cover, len(tf.Covers))
	for kind, cover := range tf.Covers {
		if cover == nil {
			continue
		}
		covers[kind] = &matroska.Cover{Filename: cover.Filename, MIME: cover.MIME, Data: cover.Data}
	}
	if tf.Cover != nil && covers["front"] == nil {
		covers["front"] = &matroska.Cover{Filename: tf.Cover.Filename, MIME: tf.Cover.MIME, Data: tf.Cover.Data}
	}
	return matroska.Write(path, tf.Fields, covers)
}
