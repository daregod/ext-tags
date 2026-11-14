package importer

import (
	"sort"
	"strings"

	"github.com/daregod/ext-tags/internal/formats/flac"
	"github.com/daregod/ext-tags/internal/formats/vorbis"
)

var vorbisFields = map[string]string{
	"Title":        "TITLE",
	"Artist":       "ARTIST",
	"Album":        "ALBUM",
	"Album artist": "ALBUMARTIST",
	"Track":        "TRACKNUMBER",
	"Disc":         "DISCNUMBER",
	"Year":         "DATE",
	"Genre":        "GENRE",
	"Composer":     "COMPOSER",
	"Publisher":    "PUBLISHER",
	"Comment":      "COMMENT",
	"Encoder":      "ENCODER",
}

var flacPictureTypes = map[string]uint32{
	"front":  3,
	"back":   4,
	"artist": 8,
	"disc":   6,
	"icon":   1,
}

// ImportFLAC writes the fields and artwork of tf into a FLAC file.
func ImportFLAC(path string, tf *TagFile, opts Options) error {
	if opts.Backup {
		if err := copyFile(path, path+".bak"); err != nil {
			return err
		}
	}

	var comments []vorbis.Comment
	for field, value := range tf.Fields {
		if value == "" {
			continue
		}
		name, ok := vorbisFields[field]
		if !ok {
			name = strings.ToUpper(strings.ReplaceAll(field, " ", ""))
		}
		if field == "Track" || field == "Disc" {
			num, total := splitPair(value)
			comments = append(comments, vorbis.Comment{Name: name, Value: num})
			if total != "" {
				totalName := "TRACKTOTAL"
				if field == "Disc" {
					totalName = "DISCTOTAL"
				}
				comments = append(comments, vorbis.Comment{Name: totalName, Value: total})
			}
			continue
		}
		comments = append(comments, vorbis.Comment{Name: name, Value: value})
	}
	sort.Slice(comments, func(i, j int) bool {
		if comments[i].Name == comments[j].Name {
			return comments[i].Value < comments[j].Value
		}
		return comments[i].Name < comments[j].Name
	})

	covers := tf.Covers
	if len(covers) == 0 && tf.Cover != nil {
		covers = map[string]*Cover{"front": tf.Cover}
	}
	var pictures []vorbis.Picture
	for kind, cover := range covers {
		pictureType, ok := flacPictureTypes[kind]
		if !ok || cover == nil || len(cover.Data) == 0 {
			continue
		}
		pictures = append(pictures, vorbis.Picture{
			Type: pictureType,
			MIME: cover.MIME,
			Data: cover.Data,
		})
	}
	sort.Slice(pictures, func(i, j int) bool { return pictures[i].Type < pictures[j].Type })

	return flac.WriteWithRemove(path, comments, pictures, opts.Replace, onlyVorbis(opts.Only), onlyPictures(opts.Only))
}
func splitPair(value string) (string, string) {
	parts := strings.SplitN(value, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}
func onlyVorbis(fields []string) map[string]bool {
	if len(fields) == 0 {
		return nil
	}
	out := make(map[string]bool, len(fields)+2)
	for _, field := range fields {
		name, ok := vorbisFields[field]
		if !ok {
			name = strings.ToUpper(strings.ReplaceAll(field, " ", ""))
		}
		out[name] = true
		switch field {
		case "Track":
			out["TRACKTOTAL"] = true
		case "Disc":
			out["DISCTOTAL"] = true
		}
	}
	return out
}

var onlyPictureKinds = map[string]uint32{
	"cover art (front)": 3,
	"cover art (back)":  4,
	"artist picture":    8,
	"cover art (disc)":  6,
	"cover art (icon)":  1,
}

func onlyPictures(fields []string) map[uint32]bool {
	out := make(map[uint32]bool)
	for _, field := range fields {
		if typ, ok := onlyPictureKinds[strings.ToLower(strings.TrimSpace(field))]; ok {
			out[typ] = true
			if typ == 3 {
				out[0] = true
			}
		}
	}
	return out
}
