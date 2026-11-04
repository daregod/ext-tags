package importer

import (
	"sort"
	"strings"

	"github.com/daregod/ext-tags/internal/flac"
	"github.com/daregod/ext-tags/internal/vorbis"
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

	return flac.Write(path, comments, pictures, opts.Replace)
}
func splitPair(value string) (string, string) {
	parts := strings.SplitN(value, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}
