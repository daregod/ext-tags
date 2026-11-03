package extractor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"external-tags/internal/matroska"
)

var ebmlMagic = []byte{0x1A, 0x45, 0xDF, 0xA3}

// extractMatroska reads WebM/MKV metadata.
func extractMatroska(path string, coverNames []string) (*Result, error) {
	md, err := matroska.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read matroska: %w", err)
	}

	res := empty()
	res.Format = "MATROSKA"

	for _, tag := range md.Tags {
		if tag.Name == "" || tag.Value == "" {
			continue
		}

		name := tag.Name
		albumScope := tag.TrackUID == 0 && tag.TargetTypeValue != 30
		if albumScope {
			// foobar2000 treats album-level Matroska tags this way.
			switch strings.ToUpper(name) {
			case "ARTIST":
				name = "ALBUMARTIST"
			case "PART_NUMBER":
				name = "DISCNUMBER"
			case "TOTAL_PARTS":
				name = "DISCTOTAL"
			}
		} else if strings.EqualFold(name, "ENCODER") {
			// Codec-level encoder; foobar2000 keeps the muxer value only.
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

	// Info/Title is a container-level title, not a SimpleTag.
	if md.Title != "" {
		if _, ok := res.Fields["TITLE"]; !ok {
			res.Fields["TITLE"] = md.Title
		}
	}

	putVar(res.Vars, "tracknumber", fieldValue(res.Fields, "PART_NUMBER", "TRACKNUMBER"))
	putVar(res.Vars, "totaltracks", fieldValue(res.Fields, "TOTAL_PARTS", "TRACKTOTAL", "TOTALTRACKS"))
	putVar(res.Vars, "discnumber", fieldValue(res.Fields, "DISCNUMBER", "DISC"))
	putVar(res.Vars, "disctotal", fieldValue(res.Fields, "DISCTOTAL", "TOTALDISCS"))

	if att := pickCover(md.Attachments, coverNames); att != nil {
		res.Cover = &Cover{
			Filename: coverName(att.Filename, att.MIME),
			MIME:     att.MIME,
			Data:     att.Data,
		}
	}

	if len(res.Fields) == 0 && res.Cover == nil {
		return res, ErrNoTags
	}
	return res, nil
}

// defaultCoverNames is used when the configuration does not override names.
var defaultCoverNames = []string{"cover", "folder", "box", "disc", "media"}

// pickCover chooses the most cover-like image attachment: exact base-name
// matches in the configured priority order, then the first image.
func pickCover(attachments []matroska.Attachment, names []string) *matroska.Attachment {
	if len(names) == 0 {
		names = defaultCoverNames
	}

	var images []*matroska.Attachment
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

func fieldValue(fields map[string]string, names ...string) string {
	for _, name := range names {
		for key, value := range fields {
			if strings.EqualFold(key, name) {
				return value
			}
		}
	}
	return ""
}

func coverName(filename, mime string) string {
	if filename != "" {
		if ext := strings.ToLower(filepath.Ext(filename)); ext != "" {
			return "cover" + ext
		}
	}
	switch strings.ToLower(mime) {
	case "image/jpeg":
		return "cover.jpg"
	case "image/png":
		return "cover.png"
	}
	if i := strings.Index(strings.ToLower(mime), "/"); i >= 0 && i+1 < len(mime) {
		return "cover." + strings.ToLower(mime[i+1:])
	}
	return "cover"
}

// readMagic returns the first n bytes of a file.
func readMagic(path string, n int) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := make([]byte, n)
	read, err := f.Read(buf)
	if err != nil && read == 0 {
		return nil, err
	}
	return buf[:read], nil
}
