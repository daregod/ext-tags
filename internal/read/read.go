// Package read contains shared reading helpers and the dhowden/tag adapter.
package read

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/dhowden/tag"

	"github.com/daregod/ext-tags/internal/metadata"
)

// ErrNoTags means the file has no metadata block at all.
var ErrNoTags = errors.New("no tags")

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

// New returns an empty document.
func New() *metadata.Document {
	return metadata.New()
}

// Taglib reads metadata through dhowden/tag.
func Taglib(path string) (*metadata.Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		if errors.Is(err, tag.ErrNoTagsFound) {
			return nil, ErrNoTags
		}
		return nil, fmt.Errorf("read tags: %w", err)
	}
	res := New()
	res.Format = string(m.Format())
	ApplyTagMetadata(m, res)
	return res, nil
}

// ApplyTagMetadata copies dhowden metadata into a document.
func ApplyTagMetadata(m tag.Metadata, res *metadata.Document) {
	for k, v := range m.Raw() {
		key, ok := NormalizeKey(k)
		if !ok {
			continue
		}
		val, ok := Stringify(v)
		if !ok || val == "" {
			continue
		}
		upperKey := strings.ToUpper(key)
		if IsArtworkField(upperKey) {
			continue
		}
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

	track, trackTotal := TrackNumbers(m.Raw())
	disc, discTotal := DiscNumbers(m.Raw())
	if t, tt := m.Track(); track == "" && t > 0 {
		track = strconv.Itoa(t)
		if trackTotal == "" && tt > 0 {
			trackTotal = strconv.Itoa(tt)
		}
	}
	if d, dt := m.Disc(); disc == "" && d > 0 {
		disc = strconv.Itoa(d)
		if discTotal == "" && dt > 0 {
			discTotal = strconv.Itoa(dt)
		}
	}
	PutVar(res.Vars, "tracknumber", track)
	PutVar(res.Vars, "totaltracks", trackTotal)
	PutVar(res.Vars, "discnumber", disc)
	PutVar(res.Vars, "disctotal", discTotal)

	if p := m.Picture(); p != nil {
		cover := &metadata.Cover{
			Filename: CoverFilename(p),
			MIME:     p.MIMEType,
			Data:     p.Data,
		}
		res.Cover = cover
		res.Artwork[metadata.Front] = cover
	}
}

// PutVar stores a non-empty template variable.
func PutVar(m map[string]string, key, val string) {
	if val != "" {
		m[key] = val
	}
}

// NormalizeKey converts container keys to UTF-8.
func NormalizeKey(k string) (string, bool) {
	if utf8.ValidString(k) {
		return k, true
	}
	if len(k) > 1 && k[0] == 0xA9 && utf8.ValidString(k[1:]) {
		return "©" + k[1:], true
	}
	return "", false
}

// Stringify converts a tag value to a string.
func Stringify(v interface{}) (string, bool) {
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

// TrackNumbers returns number/total from raw fields.
func TrackNumbers(raw map[string]interface{}) (string, string) {
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

// DiscNumbers returns number/total from raw fields.
func DiscNumbers(raw map[string]interface{}) (string, string) {
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

// CoverFilename builds a suggested cover filename from a dhowden picture.
func CoverFilename(p *tag.Picture) string {
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

// CoverName normalises a suggested cover filename.
func CoverName(filename, mime string) string {
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

// FieldValue returns the first non-empty field from names.
func FieldValue(fields map[string]string, names ...string) string {
	for _, name := range names {
		if value := fields[name]; value != "" {
			return value
		}
	}
	return ""
}

// SplitPair splits "3/12" into "3" and "12".
func SplitPair(value string) (string, string) {
	if value == "" {
		return "", ""
	}
	parts := strings.SplitN(value, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

// ReadMagic returns the first n bytes of a file.
func ReadMagic(path string, n int) ([]byte, error) {
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

// IsArtworkField reports keys that are artwork rather than text fields.
func IsArtworkField(key string) bool {
	switch key {
	case "METADATA_BLOCK_PICTURE", "COVERART", "COVERARTMIME":
		return true
	}
	return false
}
