package extractor

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"external-tags/internal/ape"
	"github.com/dhowden/tag"
)

var SupportedExtensions = map[string]bool{
	".mp3":  true,
	".flac": true,
	".opus": true,
	".ogg":  true,
	".m4a":  true,
	".mp4":  true,
	".wav":  true,
	".aif":  true,
	".aiff": true,
}

var keyMap = map[string]string{
	"TIT2": "Title", "TPE1": "Artist", "TALB": "Album", "TPE2": "Album artist",
	"TCON": "Genre", "TDRC": "Year", "TYER": "Year",
	"TCOM": "Composer", "COMM": "Comment", "TPUB": "Publisher",
	"TCOP": "Copyright", "TSSE": "Encoder", "TLAN": "Language", "TENC": "Encoded by",
	"TBPM": "BPM", "TKEY": "Key", "TEXT": "Lyricist", "TMED": "Media",

	"TITLE": "Title", "ARTIST": "Artist", "ALBUM": "Album",
	"ALBUMARTIST": "Album artist", "ALBUM ARTIST": "Album artist",
	"DATE": "Year", "YEAR": "Year", "GENRE": "Genre",
	"COMPOSER": "Composer", "PERFORMER": "Performer",
	"COMMENT": "Comment", "RATING": "Rating",

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

func Extract(path string) (*ape.Tag, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m, err := tag.ReadFrom(f)
	if err != nil {
		return nil, fmt.Errorf("read tags: %w", err)
	}

	t := ape.NewTag()
	seen := make(map[string]bool)

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

	for k, v := range m.Raw() {
		upperK := strings.ToUpper(k)

		// Пропускаем треки и диски, так как они обрабатываются строго ниже
		if upperK == "TRACKNUMBER" || upperK == "TRACKTOTAL" || upperK == "TOTALTRACKS" || upperK == "TRACKS" || upperK == "TOTALTRACK" ||
			upperK == "DISCNUMBER" || upperK == "DISCTOTAL" || upperK == "TOTALDISCS" || upperK == "DISCS" || upperK == "TOTALDISC" {
			continue
		}

		mappedKey, ok := keyMap[upperK]
		if !ok {
			if isID3Frame(k) {
				continue
			}
			mappedKey = k
		}

		switch val := v.(type) {
		case string:
			addText(mappedKey, val)
		case []string:
			if len(val) > 0 {
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

	// Обработка треков и дисков вручную
	track, trackTotal := getTrack(m.Raw())
	disc, discTotal := getDisc(m.Raw())

	// Резервный вариант, если в сырых данных ничего не нашли
	if track == "" {
		if n, tot := m.Track(); n != 0 {
			track = fmt.Sprintf("%d", n)
			if tot > 0 {
				trackTotal = fmt.Sprintf("%d", tot)
			}
		}
	}
	if disc == "" {
		if n, tot := m.Disc(); n != 0 {
			disc = fmt.Sprintf("%d", n)
			if tot > 0 {
				discTotal = fmt.Sprintf("%d", tot)
			}
		}
	}

	// Записываем в формате "02/16"
	if track != "" {
		if trackTotal != "" {
			addText("Track", track+"/"+trackTotal)
		} else {
			addText("Track", track)
		}
	}
	if disc != "" {
		if discTotal != "" {
			addText("Disc", disc+"/"+discTotal)
		} else {
			addText("Disc", disc)
		}
	}

	if p := m.Picture(); p != nil {
		t.AddCoverArt("Cover Art (Front)", coverFilename(p), p.Data)
	}

	return t, nil
}

func getTrack(raw map[string]interface{}) (string, string) {
	var num, total string
	for k, v := range raw {
		upperK := strings.ToUpper(k)
		if upperK == "TRACKNUMBER" || upperK == "TRACK" || upperK == "TRCK" {
			str := getStr(v)
			parts := strings.Split(str, "/")
			num = parts[0]
			if len(parts) > 1 {
				total = parts[1]
			}
		}
		if upperK == "TRACKTOTAL" || upperK == "TOTALTRACKS" || upperK == "TRACKS" || upperK == "TOTALTRACK" {
			total = getStr(v)
		}
	}
	return num, total
}

func getDisc(raw map[string]interface{}) (string, string) {
	var num, total string
	for k, v := range raw {
		upperK := strings.ToUpper(k)
		if upperK == "DISCNUMBER" || upperK == "DISC" || upperK == "TPOS" {
			str := getStr(v)
			parts := strings.Split(str, "/")
			num = parts[0]
			if len(parts) > 1 {
				total = parts[1]
			}
		}
		if upperK == "DISCTOTAL" || upperK == "TOTALDISCS" || upperK == "DISCS" || upperK == "TOTALDISC" {
			total = getStr(v)
		}
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
