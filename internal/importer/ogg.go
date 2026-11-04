package importer

import (
	"sort"
	"strings"

	"github.com/daregod/ext-tags/internal/ogg"
	"github.com/daregod/ext-tags/internal/vorbis"
)

// ImportOgg writes fields and artwork into Ogg Vorbis, Opus or Speex.
func ImportOgg(path string, tf *TagFile, opts Options) error {
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

	return ogg.WriteComments(path, comments, pictures, opts.Replace)
}
