package importer

import (
	"strings"

	"github.com/daregod/ext-tags/internal/formats/matroska"
	"github.com/daregod/ext-tags/internal/read"
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
	if len(opts.Only) == 0 {
		return matroska.Write(path, tf.Fields, covers)
	}

	artOnly := onlyArtworkSlots(opts.Only)
	existing, err := matroska.Read(path, nil)
	if err != nil && err != read.ErrNoTags {
		return err
	}
	fields := make(map[string]string)
	if existing != nil {
		for key, value := range existing.Fields {
			fields[key] = value
		}
		for kind, cover := range existing.Artwork {
			if covers[kind] == nil {
				covers[kind] = &matroska.Cover{Filename: cover.Filename, MIME: cover.MIME, Data: cover.Data}
			}
		}
	}
	for _, field := range opts.Only {
		raw := matroskaName(field)
		for key := range fields {
			if strings.EqualFold(key, raw) || (field == "Track" && (strings.EqualFold(key, "TRACKNUMBER") || strings.EqualFold(key, "PART_NUMBER"))) || (field == "Disc" && (strings.EqualFold(key, "DISCNUMBER") || strings.EqualFold(key, "DISC"))) {
				delete(fields, key)
			}
		}
		value, ok := tf.Fields[field]
		if ok && value != "" {
			fields[raw] = value
		}
	}
	if len(artOnly) > 0 {
		md, err := matroska.ReadFile(path)
		if err != nil {
			return err
		}
		attachments := append([]matroska.Attachment(nil), md.Attachments...)
		for kind := range artOnly {
			idx := matroska.AttachmentIndex(attachments, kind)
			cover := tf.Covers[kind]
			if kind == "front" && cover == nil {
				cover = tf.Cover
			}
			if cover != nil {
				attachment := matroska.Attachment{Filename: cover.Filename, MIME: cover.MIME, Data: cover.Data}
				if idx >= 0 {
					attachments[idx] = attachment
				} else {
					attachments = append(attachments, attachment)
				}
			} else if idx >= 0 {
				attachments = append(attachments[:idx], attachments[idx+1:]...)
			}
		}
		return matroska.WriteWithAttachments(path, fields, attachments)
	}
	return matroska.WriteKeepAttachments(path, fields)
}

func matroskaName(field string) string {
	switch field {
	case "Track":
		return "PART_NUMBER"
	case "Disc":
		return "DISC"
	case "Album artist":
		return "ALBUMARTIST"
	case "Year":
		return "DATE"
	}
	return strings.ToUpper(strings.ReplaceAll(field, " ", "_"))
}
