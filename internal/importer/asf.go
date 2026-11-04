package importer

import (
	"github.com/daregod/ext-tags/internal/asf"
)

// ImportASF writes fields and WM/Picture artwork into WMA/WMV/ASF.
func ImportASF(path string, tf *TagFile, opts Options) error {
	if opts.Backup {
		if err := copyFile(path, path+".bak"); err != nil {
			return err
		}
	}
	covers := make(map[string]*asf.Cover, len(tf.Covers))
	for kind, cover := range tf.Covers {
		if cover == nil {
			continue
		}
		covers[kind] = &asf.Cover{Filename: cover.Filename, MIME: cover.MIME, Data: cover.Data}
	}
	if tf.Cover != nil && covers["front"] == nil {
		covers["front"] = &asf.Cover{Filename: tf.Cover.Filename, MIME: tf.Cover.MIME, Data: tf.Cover.Data}
	}
	return asf.Write(path, tf.Fields, covers, opts.Replace)
}
