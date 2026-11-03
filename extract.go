package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dhowden/tag"
)

// Маппинг оригинальных ключей (ID3v2 и Vorbis) в чистые ключи APEv2
var keyMap = map[string]string{
	// ID3v2 фреймы (MP3)
	"TIT2": "Title", "TPE1": "Artist", "TALB": "Album", "TPE2": "Album artist",
	"TCON": "Genre", "TDRC": "Year", "TYER": "Year", "TRCK": "Track",
	"TPOS": "Disc", "TCOM": "Composer", "COMM": "Comment", "TPUB": "Publisher",
	"TCOP": "Copyright", "TSSE": "Encoder", "TLAN": "Language", "TENC": "Encoded by",
	"TBPM": "BPM", "TKEY": "Key", "TEXT": "Lyricist", "TMED": "Media",

	// Vorbis Comments (FLAC / Opus) - приводим к ВЕРХНЕМУ регистру для поиска
	"TITLE": "Title", "ARTIST": "Artist", "ALBUM": "Album",
	"ALBUMARTIST": "Album artist", "ALBUM ARTIST": "Album artist",
	"DATE": "Year", "YEAR": "Year", "GENRE": "Genre",
	"COMPOSER": "Composer", "PERFORMER": "Performer",
	"TRACKNUMBER": "Track", "TRACKTOTAL": "Total tracks", "TOTALTRACKS": "Total tracks",
	"DISCNUMBER": "Disc", "DISCTOTAL": "Total discs", "TOTALDISCS": "Total discs",
	"COMMENT": "Comment", "RATING": "Rating",

	// Кастомные поля
	"CATALOGNUMBER":          "Catalognumber",
	"CATALOG NUMBER":         "Catalognumber",
	"COMMENT CATALOG NUMBER": "Comment catalog number",
	"COPYRIGHT":              "Copyright",
	"ISRC":                   "Isrc",
	"LABEL":                  "Label",
	"ORGANIZATION":           "Publisher",
	"PUBLISHER":              "Publisher",
	"MUSICBRAINZ_ALBUMID":    "Musicbrainz_albumid",
	"MUSICBRAINZ_TRACKID":    "Musicbrainz_trackid",
	"RELEASETIME":            "Releasetime",
	"UPC":                    "Upc",
	"WWW":                    "Www",
	"REPLAYGAIN_TRACK_GAIN":  "replaygain_track_gain",
	"REPLAYGAIN_TRACK_PEAK":  "replaygain_track_peak",
}

func isID3Frame(k string) bool {
	if len(k) != 4 {
		return false
	}
	for _, c := range k {
		if c < 'A' || c > 'Z' {
			return false
		}
	}
	return true
}

func ExtractFromAudio(path string) (*APETag, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		return nil, fmt.Errorf("read tags: %w", err)
	}

	t := NewAPETag()
	seen := make(map[string]bool)

	// Функция добавления. Нормализуем ключ (нижний регистр, без пробелов) для надежной проверки
	addText := func(k, v string) {
		if k == "" || v == "" {
			return
		}
		norm := strings.ToLower(strings.ReplaceAll(k, " ", ""))
		if seen[norm] {
			return
		}
		seen[norm] = true
		t.AddText(k, v)
	}

	// 1. Основной проход по сырым данным
	for k, v := range m.Raw() {
		upperK := strings.ToUpper(k)
		mappedKey, ok := keyMap[upperK]
		if !ok {
			if isID3Frame(k) {
				continue // Пропускаем неизвестные ID3v2 фреймы
			}
			mappedKey = k // Оставляем как есть
		}

		switch val := v.(type) {
		case string:
			addText(mappedKey, val)
		case []string:
			if len(val) > 0 {
				// Дедупликация значений внутри массива (например, Date: [2025, 2025] -> "2025")
				unique := []string{}
				vSeen := make(map[string]bool)
				for _, s := range val {
					s = strings.TrimSpace(s)
					if s != "" && !vSeen[s] {
						vSeen[s] = true
						unique = append(unique, s)
					}
				}
				if len(unique) > 0 {
					addText(mappedKey, strings.Join(unique, "; "))
				}
			}
		case int:
			addText(mappedKey, strconv.Itoa(val))
		case int64:
			addText(mappedKey, strconv.FormatInt(val, 10))
		}
	}

	// 2. Резервный сбор (сработает только если поля не были добавлены из Raw)
	if m.Title() != "" {
		addText("Title", m.Title())
	}
	if m.Artist() != "" {
		addText("Artist", m.Artist())
	}
	if m.Album() != "" {
		addText("Album", m.Album())
	}
	if m.AlbumArtist() != "" {
		addText("Album artist", m.AlbumArtist())
	}
	if m.Genre() != "" {
		addText("Genre", m.Genre())
	}
	if m.Composer() != "" {
		addText("Composer", m.Composer())
	}
	if m.Comment() != "" {
		addText("Comment", m.Comment())
	}
	if m.Year() != 0 {
		addText("Year", strconv.Itoa(m.Year()))
	}

	// 3. Обложка
	if p := m.Picture(); p != nil {
		t.AddCoverArt("Cover Art (Front)", coverFilename(p), p.Data)
	}

	return t, nil
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
