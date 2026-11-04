// Package extractor reads raw metadata from audio containers.
//
// It is deliberately policy-free: it returns container fields as they are,
// plus a few normalised values (track/disc numbers) used by field rules.
package extractor

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/dhowden/tag"
)

// ErrNoTags means the file has no metadata block at all.
var ErrNoTags = errors.New("no tags")

// Cover is a picture embedded in the container.
type Cover struct {
	// Filename is a suggested name for the APEv2 cover item ("cover.jpg").
	Filename string
	MIME     string
	Data     []byte
}

// Result is raw metadata, before any field rules are applied.
type Result struct {
	Format string
	Fields map[string]string
	Vars   map[string]string
	Cover  *Cover
}

// IsSupported reports whether path has one of the configured extensions.
func IsSupported(path string, extensions []string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, e := range extensions {
		if ext == e {
			return true
		}
	}
	return false
}

// Options tune container parsing.
type Options struct {
	// CoverNames are preferred attachment base names (cover, folder, box…).
	CoverNames []string
}

// Extract reads metadata from path with default options.
func Extract(path string) (*Result, error) {
	return ExtractWithOptions(path, Options{})
}

// ExtractWithOptions reads metadata from path. It returns ErrNoTags when the
// file has no tag block at all.
func ExtractWithOptions(path string, opts Options) (*Result, error) {
	// AC3 and DTS have no native tag container; foobar2000 appends APEv2.
	// Extension wins over magic: a .dts file may actually be Matroska with DTS.
	switch strings.ToLower(filepath.Ext(path)) {
	case ".ac3":
		return extractAPEv2(path, "AC3")
	case ".dts":
		return extractAPEv2(path, "DTS")
	}
	if magic, err := readMagic(path, 40); err == nil {
		switch {
		case bytes.HasPrefix(magic, ebmlMagic):
			return extractMatroska(path, opts.CoverNames)
		case bytes.HasPrefix(magic, apeMagic):
			return extractAPEv2(path, "APE")
		case bytes.HasPrefix(magic, wvpkMagic):
			return extractAPEv2(path, "WAVPACK")
		case bytes.HasPrefix(magic, []byte("MPCK")):
			return extractAPEv2(path, "MUSEPACK")
		case bytes.HasPrefix(magic, []byte("MP+")):
			return extractAPEv2(path, "MUSEPACK")
		case bytes.HasPrefix(magic, []byte("tBaK")):
			return extractAPEv2(path, "TAK")
		case bytes.HasPrefix(magic, asfMagic):
			return extractASF(path)
		case bytes.HasPrefix(magic, oggMagic):
			return extractOgg(path)
		case len(magic) >= 12 && bytes.Equal(magic[:4], riffMagic) && string(magic[8:12]) == "WAVE":
			return extractWAV(path)
		case len(magic) >= 12 && bytes.Equal(magic[:4], rf64Magic) && string(magic[8:12]) == "WAVE":
			return extractRF64(path)
		case len(magic) >= 40 && string(magic[0:4]) == "riff" && string(magic[24:28]) == "wave":
			return extractWave64(path)
		case len(magic) >= 12 && bytes.Equal(magic[:4], formMagic) && (string(magic[8:12]) == "AIFF" || string(magic[8:12]) == "AIFC"):
			return extractAIFF(path)
		}
	}
	return extractTaglib(path)
}

func extractTaglib(path string) (*Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		if errors.Is(err, tag.ErrNoTagsFound) {
			return empty(), ErrNoTags
		}
		return nil, fmt.Errorf("read tags: %w", err)
	}

	res := empty()
	res.Format = string(m.Format())
	applyTagMetadata(m, res)

	return res, nil
}

// applyTagMetadata copies raw container metadata into res.
func applyTagMetadata(m tag.Metadata, res *Result) {
	for k, v := range m.Raw() {
		key, ok := normalizeKey(k)
		if !ok {
			continue
		}
		val, ok := stringify(v)
		if !ok || val == "" {
			continue
		}
		upperKey := strings.ToUpper(key)
		if (isTrackKey(upperKey) || isTrackTotalKey(upperKey) || isDiscKey(upperKey) || isDiscTotalKey(upperKey)) && val == "0" {
			continue
		}
		if prev, exists := res.Fields[key]; exists {
			if prev != val {
				res.Fields[key] = prev + "; " + val
			}
			continue
		}
		res.Fields[key] = val
	}

	// Fallbacks for formats where Raw() does not expose common fields.
	for key, val := range map[string]string{
		"TITLE":       m.Title(),
		"ARTIST":      m.Artist(),
		"ALBUM":       m.Album(),
		"ALBUMARTIST": m.AlbumArtist(),
		"GENRE":       m.Genre(),
		"COMMENT":     m.Comment(),
	} {
		if val == "" {
			continue
		}
		if _, exists := res.Fields[key]; !exists {
			res.Fields[key] = val
		}
	}
	if year := m.Year(); year != 0 {
		if _, exists := res.Fields["YEAR"]; !exists {
			res.Fields["YEAR"] = strconv.Itoa(year)
		}
	}

	track, trackTotal := trackNumbers(m.Raw())
	disc, discTotal := discNumbers(m.Raw())
	if track == "" {
		if n, total := m.Track(); n != 0 {
			track = strconv.Itoa(n)
			if total > 0 {
				trackTotal = strconv.Itoa(total)
			}
		}
	}
	if disc == "" {
		if n, total := m.Disc(); n != 0 {
			disc = strconv.Itoa(n)
			if total > 0 {
				discTotal = strconv.Itoa(total)
			}
		}
	}
	putVar(res.Vars, "tracknumber", track)
	putVar(res.Vars, "totaltracks", trackTotal)
	putVar(res.Vars, "discnumber", disc)
	putVar(res.Vars, "disctotal", discTotal)

	if p := m.Picture(); p != nil {
		res.Cover = &Cover{
			Filename: coverFilename(p),
			MIME:     p.MIMEType,
			Data:     p.Data,
		}
	}

}

func empty() *Result {
	return &Result{Fields: map[string]string{}, Vars: map[string]string{}}
}

func putVar(m map[string]string, key, val string) {
	if val != "" {
		m[key] = val
	}
}

// normalizeKey converts container keys to UTF-8. MP4 atoms carry a LATIN-1
// copyright sign (0xA9) which is rewritten to the Unicode "©".
func normalizeKey(k string) (string, bool) {
	if utf8.ValidString(k) {
		return k, true
	}
	if len(k) > 1 && k[0] == 0xA9 && utf8.ValidString(k[1:]) {
		return "©" + k[1:], true
	}
	return "", false
}

func stringify(v interface{}) (string, bool) {
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val), true
	case []string:
		seen := make(map[string]bool, len(val))
		unique := make([]string, 0, len(val))
		for _, s := range val {
			s = strings.TrimSpace(s)
			if s != "" && !seen[s] {
				seen[s] = true
				unique = append(unique, s)
			}
		}
		return strings.Join(unique, "; "), true
	case int:
		return strconv.Itoa(val), true
	case int64:
		return strconv.FormatInt(val, 10), true
	}
	return "", false
}

func isTrackKey(k string) bool {
	switch k {
	case "TRACKNUMBER", "TRACK", "TRCK":
		return true
	}
	return false
}

func isTrackTotalKey(k string) bool {
	switch k {
	case "TRACKTOTAL", "TOTALTRACKS", "TRACKS", "TOTALTRACK":
		return true
	}
	return false
}

func isDiscKey(k string) bool {
	switch k {
	case "DISCNUMBER", "DISC", "TPOS":
		return true
	}
	return false
}

func isDiscTotalKey(k string) bool {
	switch k {
	case "DISCTOTAL", "TOTALDISCS", "DISCS", "TOTALDISC":
		return true
	}
	return false
}

func trackNumbers(raw map[string]interface{}) (string, string) {
	var num, total string
	for k, v := range raw {
		upper := strings.ToUpper(k)
		if isTrackKey(upper) {
			str := getStr(v)
			parts := strings.Split(str, "/")
			num = parts[0]
			if len(parts) > 1 {
				total = parts[1]
			}
		}
		if isTrackTotalKey(upper) {
			total = getStr(v)
		}
	}
	if num == "0" {
		num = ""
	}
	if total == "0" {
		total = ""
	}
	return num, total
}

func discNumbers(raw map[string]interface{}) (string, string) {
	var num, total string
	for k, v := range raw {
		upper := strings.ToUpper(k)
		if isDiscKey(upper) {
			str := getStr(v)
			parts := strings.Split(str, "/")
			num = parts[0]
			if len(parts) > 1 {
				total = parts[1]
			}
		}
		if isDiscTotalKey(upper) {
			total = getStr(v)
		}
	}
	if num == "0" {
		num = ""
	}
	if total == "0" {
		total = ""
	}
	return num, total
}

func getStr(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case []string:
		if len(val) > 0 {
			return val[0]
		}
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	}
	return ""
}

func coverFilename(p *tag.Picture) string {
	switch strings.ToLower(p.Ext) {
	case "jpg", "jpeg":
		return "cover.jpg"
	case "png":
		return "cover.png"
	}
	switch p.MIMEType {
	case "image/jpeg":
		return "cover.jpg"
	case "image/png":
		return "cover.png"
	}
	if p.Ext != "" {
		return "cover." + strings.ToLower(p.Ext)
	}
	return "cover"
}
