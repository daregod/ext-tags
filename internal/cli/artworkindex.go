package cli

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/formats/flac"
	"github.com/daregod/ext-tags/internal/formats/id3"
	"github.com/daregod/ext-tags/internal/formats/matroska"
	"github.com/daregod/ext-tags/internal/formats/riff"
	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/read"
	"github.com/daregod/ext-tags/internal/tagfile"
	"github.com/daregod/ext-tags/internal/tagformat"
)

type indexedArtwork struct {
	Source   string
	Index    int
	Slot     string
	Filename string
	MIME     string
	Data     []byte
}

func (a *app) loadIndexedArtwork(path string, cfg *config.Config) ([]indexedArtwork, error) {
	reader, ok := a.getRegistry().ReaderFor(path)
	if !ok {
		return nil, errors.New("unsupported format " + strings.ToLower(filepath.Ext(path)))
	}
	doc, err := reader.Read(path, tagformat.ReadOptions{CoverNames: cfg.Cover.Names})
	if errors.Is(err, read.ErrNoTags) {
		doc = metadata.New()
	} else if err != nil {
		return nil, err
	}
	pictures := containerPictures(path, doc)
	out := make([]indexedArtwork, 0, len(pictures))
	for i, pic := range pictures {
		out = append(out, indexedArtwork{Source: "container", Index: i, Slot: pic.Slot, Filename: pic.Filename, MIME: pic.MIME, Data: pic.Data})
	}
	if cfg.Files.MergeExisting {
		tag, err := tagfile.ReadExisting(path, cfg.TagFile.Suffix)
		if err != nil {
			return nil, err
		}
		for i, pic := range tagfile.ExistingPictures(tag) {
			out = append(out, indexedArtwork{Source: "tag", Index: i, Slot: pic.Slot, Filename: pic.Filename, MIME: pic.MIME, Data: pic.Data})
		}
	}
	return out, nil
}

func containerPictures(path string, doc *metadata.Document) []metadata.Picture {
	if len(doc.Pictures) > 0 {
		return doc.Pictures
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp3", ".mp2", ".mp1":
		var out []metadata.Picture
		for _, cover := range id3.ReadFilePictures(path) {
			out = append(out, metadata.Picture{Slot: pictureSlot(cover.Type), Filename: cover.Filename, MIME: cover.MIME, Data: cover.Data})
		}
		if len(out) > 0 {
			return out
		}
	case ".flac":
		if _, pictures, err := flac.Read(path); err == nil && len(pictures) > 0 {
			var out []metadata.Picture
			for _, pic := range pictures {
				out = append(out, metadata.Picture{Slot: vorbisSlot(pic.Type), MIME: pic.MIME, Data: pic.Data})
			}
			return out
		}
	case ".mkv", ".webm":
		if md, err := matroska.ReadFile(path); err == nil && len(md.Attachments) > 0 {
			var out []metadata.Picture
			for i, att := range md.Attachments {
				out = append(out, metadata.Picture{Slot: matroskaSlot(md.Attachments, i), Filename: att.Filename, MIME: att.MIME, Data: att.Data})
			}
			return out
		}
	case ".wav", ".rf64", ".w64", ".aif", ".aiff":
		if existing := riffID3(path); len(existing) > 0 {
			var out []metadata.Picture
			for _, cover := range id3.ReadPictures(existing) {
				out = append(out, metadata.Picture{Slot: pictureSlot(cover.Type), Filename: cover.Filename, MIME: cover.MIME, Data: cover.Data})
			}
			if len(out) > 0 {
				return out
			}
		}
	}
	// Formats without a full picture list: fall back to the selected slots.
	var out []metadata.Picture
	for _, kind := range artworkKinds {
		if cover := doc.Artwork[kind]; cover != nil {
			out = append(out, metadata.Picture{Slot: kind, Filename: cover.Filename, MIME: cover.MIME, Data: cover.Data})
		}
	}
	return out
}

func pictureSlot(pictureType byte) string {
	switch pictureType {
	case 0, 3:
		return "front"
	case 4:
		return "back"
	case 8:
		return "artist"
	case 6:
		return "disc"
	case 1:
		return "icon"
	}
	return ""
}

func vorbisSlot(pictureType uint32) string {
	switch pictureType {
	case 3:
		return "front"
	case 4:
		return "back"
	case 8:
		return "artist"
	case 6:
		return "disc"
	case 1:
		return "icon"
	}
	return ""
}

func matroskaSlot(attachments []matroska.Attachment, index int) string {
	for _, kind := range artworkKinds {
		if matroska.AttachmentIndex(attachments, kind) == index {
			return kind
		}
	}
	return ""
}

func riffID3(path string) []byte {
	lower := strings.ToLower(filepath.Ext(path))
	switch lower {
	case ".wav":
		if md, err := riff.ReadFile(path); err == nil {
			return md.ID3
		}
	case ".rf64":
		if md, err := riff.ReadRF64(path); err == nil {
			return md.ID3
		}
	case ".w64":
		if md, err := riff.ReadWave64(path); err == nil {
			return md.ID3
		}
	case ".aif", ".aiff":
		if md, err := riff.ReadAIFF(path); err == nil {
			return md.ID3
		}
	}
	return nil
}
