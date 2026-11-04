// Package flac rewrites FLAC metadata blocks (Vorbis comments and pictures).
package flac

import (
	"fmt"
	"os"
	"strings"

	"github.com/daregod/ext-tags/internal/vorbis"
)

const (
	blockStreamInfo = 0
	blockPadding    = 1
	blockComment    = 4
	blockPicture    = 6
)

type block struct {
	typ  byte
	data []byte
}

// Read returns comments and pictures of a FLAC file.
func Read(path string) ([]vorbis.Comment, []vorbis.Picture, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	blocks, _, err := parse(data)
	if err != nil {
		return nil, nil, err
	}
	var comments []vorbis.Comment
	var pictures []vorbis.Picture
	for _, b := range blocks {
		switch b.typ {
		case blockComment:
			_, cs, err := vorbis.ParseComments(b.data, false)
			if err != nil {
				return nil, nil, err
			}
			comments = cs
		case blockPicture:
			p, err := vorbis.ParsePicture(b.data)
			if err != nil {
				return nil, nil, err
			}
			pictures = append(pictures, p)
		}
	}
	return comments, pictures, nil
}

// Write replaces/merges comments and pictures, keeping the audio frames intact.
func Write(path string, comments []vorbis.Comment, pictures []vorbis.Picture, replace bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	blocks, audio, err := parse(data)
	if err != nil {
		return err
	}

	var existing []vorbis.Comment
	var existingPictures []vorbis.Picture
	vendor := "ext-tags"
	for _, b := range blocks {
		switch b.typ {
		case blockComment:
			if v, cs, err := vorbis.ParseComments(b.data, false); err == nil {
				if v != "" {
					vendor = v
				}
				existing = cs
			}
		case blockPicture:
			if p, err := vorbis.ParsePicture(b.data); err == nil {
				existingPictures = append(existingPictures, p)
			}
		}
	}

	merged := comments
	if !replace && len(existing) > 0 {
		replaced := make(map[string]bool, len(comments))
		for _, c := range comments {
			replaced[strings.ToUpper(c.Name)] = true
		}
		for _, c := range existing {
			if !replaced[strings.ToUpper(c.Name)] {
				merged = append(merged, c)
			}
		}
	}

	mergedPictures := pictures
	if !replace && len(existingPictures) > 0 {
		replaced := make(map[uint32]bool, len(pictures))
		for _, p := range pictures {
			replaced[p.Type] = true
		}
		for _, p := range existingPictures {
			if !replaced[p.Type] {
				mergedPictures = append(mergedPictures, p)
			}
		}
	}

	out := make([]block, 0, len(blocks)+2)
	for _, b := range blocks {
		if b.typ == blockComment || b.typ == blockPicture {
			continue
		}
		out = append(out, b)
	}
	if len(merged) > 0 {
		out = append(out, block{typ: blockComment, data: vorbis.MarshalComments(vendor, merged, false)})
	}
	for _, p := range mergedPictures {
		out = append(out, block{typ: blockPicture, data: vorbis.MarshalPicture(p)})
	}
	if len(out) == 0 {
		return fmt.Errorf("refusing to write FLAC without metadata blocks")
	}

	buf := make([]byte, 4, 4+len(audio))
	copy(buf, "fLaC")
	for i, b := range out {
		header := []byte{b.typ, 0, 0, 0}
		if i == len(out)-1 {
			header[0] |= 0x80
		}
		length := len(b.data)
		header[1] = byte(length >> 16)
		header[2] = byte(length >> 8)
		header[3] = byte(length)
		buf = append(buf, header...)
		buf = append(buf, b.data...)
	}
	buf = append(buf, audio...)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, buf, info.Mode().Perm())
}

func parse(data []byte) ([]block, []byte, error) {
	if len(data) < 4 || string(data[:4]) != "fLaC" {
		return nil, nil, fmt.Errorf("not a FLAC file")
	}
	offset := 4
	var blocks []block
	for {
		if offset+4 > len(data) {
			return nil, nil, fmt.Errorf("truncated FLAC metadata")
		}
		header := data[offset]
		last := header&0x80 != 0
		typ := header & 0x7F
		length := int(data[offset+1])<<16 | int(data[offset+2])<<8 | int(data[offset+3])
		offset += 4
		if offset+length > len(data) {
			return nil, nil, fmt.Errorf("truncated FLAC metadata block")
		}
		blocks = append(blocks, block{typ: typ, data: append([]byte(nil), data[offset:offset+length]...)})
		offset += length
		if last {
			break
		}
	}
	return blocks, data[offset:], nil
}
