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
func extractMatroska(path string) (*Result, error) {
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
		if prev, ok := res.Fields[tag.Name]; ok {
			if prev != tag.Value {
				res.Fields[tag.Name] = prev + "; " + tag.Value
			}
			continue
		}
		res.Fields[tag.Name] = tag.Value
	}

	// Info/Title is a container-level title, not a SimpleTag.
	if md.Title != "" {
		if _, ok := res.Fields["TITLE"]; !ok {
			res.Fields["TITLE"] = md.Title
		}
	}

	putVar(res.Vars, "tracknumber", fieldValue(res.Fields, "PART_NUMBER", "TRACKNUMBER"))
	putVar(res.Vars, "totaltracks", fieldValue(res.Fields, "TOTAL_PARTS", "TRACKTOTAL", "TOTALTRACKS"))

	if att := pickCover(md.Attachments); att != nil {
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

// pickCover chooses the most cover-like image attachment.
func pickCover(attachments []matroska.Attachment) *matroska.Attachment {
	var fallback *matroska.Attachment
	for i := range attachments {
		att := &attachments[i]
		if len(att.Data) == 0 || !strings.HasPrefix(strings.ToLower(att.MIME), "image/") {
			continue
		}
		if strings.Contains(strings.ToLower(att.Filename), "cover") {
			return att
		}
		if fallback == nil {
			fallback = att
		}
	}
	return fallback
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
