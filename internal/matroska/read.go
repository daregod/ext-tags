package matroska

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/read"
)

// Read reads MKV/WebM metadata.
func Read(path string, coverNames []string) (*metadata.Document, error) {
	md, err := ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read matroska: %w", err)
	}
	res := read.New()
	res.Format = "MATROSKA"
	for _, tag := range md.Tags {
		if tag.Name == "" || tag.Value == "" {
			continue
		}
		name := tag.Name
		albumScope := tag.TrackUID == 0 && tag.TargetTypeValue != 30
		if albumScope {
			switch strings.ToUpper(name) {
			case "ARTIST":
				name = "ALBUMARTIST"
			case "PART_NUMBER":
				name = "DISCNUMBER"
			case "TOTAL_PARTS":
				name = "DISCTOTAL"
			}
		} else if strings.EqualFold(name, "ENCODER") {
			continue
		}
		if prev, ok := res.Fields[name]; ok {
			if prev != tag.Value {
				res.Fields[name] = prev + "; " + tag.Value
			}
			continue
		}
		res.Fields[name] = tag.Value
	}
	if md.Title != "" {
		if _, ok := res.Fields["TITLE"]; !ok {
			res.Fields["TITLE"] = md.Title
		}
	}
	read.PutVar(res.Vars, "tracknumber", read.FieldValue(res.Fields, "PART_NUMBER", "TRACKNUMBER"))
	read.PutVar(res.Vars, "totaltracks", read.FieldValue(res.Fields, "TOTAL_PARTS", "TRACKTOTAL", "TOTALTRACKS"))
	read.PutVar(res.Vars, "discnumber", read.FieldValue(res.Fields, "DISCNUMBER", "DISC"))
	read.PutVar(res.Vars, "disctotal", read.FieldValue(res.Fields, "DISCTOTAL", "TOTALDISCS"))
	if att := pickCover(md.Attachments, coverNames); att != nil {
		cover := &metadata.Cover{Filename: read.CoverName(att.Filename, att.MIME), MIME: att.MIME, Data: att.Data}
		res.Cover = cover
		res.Artwork[metadata.Front] = cover
	}
	if len(res.Fields) == 0 && res.Cover == nil {
		return res, read.ErrNoTags
	}
	return res, nil
}

var defaultCoverNames = []string{"cover", "folder", "box", "disc", "media"}

func pickCover(attachments []Attachment, names []string) *Attachment {
	if len(names) == 0 {
		names = defaultCoverNames
	}
	var images []*Attachment
	for i := range attachments {
		att := &attachments[i]
		if len(att.Data) == 0 || !strings.HasPrefix(strings.ToLower(att.MIME), "image/") {
			continue
		}
		images = append(images, att)
	}
	for _, name := range names {
		for _, att := range images {
			if strings.EqualFold(strings.TrimSuffix(att.Filename, filepath.Ext(att.Filename)), name) {
				return att
			}
		}
	}
	if len(images) > 0 {
		return images[0]
	}
	return nil
}
