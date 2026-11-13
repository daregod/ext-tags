package cli

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/daregod/ext-tags/internal/artwork"
	"github.com/daregod/ext-tags/internal/config"
	coverpkg "github.com/daregod/ext-tags/internal/cover"
	"github.com/daregod/ext-tags/internal/metadata"
)

var artworkKinds = []string{"front", "back", "artist", "disc", "icon"}

func (a *app) sidecarArtwork(path string, cfg *config.Config) map[string]*metadata.Cover {
	out := make(map[string]*metadata.Cover)
	if !cfg.Cover.Enabled || !cfg.Cover.Directory {
		return out
	}
	specs := []struct {
		kind  string
		names []string
	}{
		{"front", cfg.Cover.Names},
		{"back", cfg.Cover.Back},
		{"artist", cfg.Cover.Artist},
		{"disc", cfg.Cover.Disc},
		{"icon", cfg.Cover.Icon},
	}
	for _, spec := range specs {
		if len(spec.names) == 0 {
			continue
		}
		img, err := artwork.Find(filepath.Dir(path), spec.names)
		if err != nil {
			continue
		}
		name := spec.kind + strings.ToLower(filepath.Ext(img.Filename))
		if spec.kind == "front" {
			name = "cover" + strings.ToLower(filepath.Ext(img.Filename))
		}
		out[spec.kind] = &metadata.Cover{Filename: name, MIME: img.MIME, Data: img.Data}
	}
	return out
}

func documentArtwork(doc *metadata.Document) map[string]*metadata.Cover {
	out := copyCovers(doc.Artwork)
	if doc.Cover != nil {
		if _, ok := out[metadata.Front]; !ok {
			out[metadata.Front] = doc.Cover
		}
	}
	return out
}

func copyCovers(in map[string]*metadata.Cover) map[string]*metadata.Cover {
	out := make(map[string]*metadata.Cover, len(in))
	for kind, c := range in {
		out[kind] = c
	}
	return out
}

type artworkSkip struct {
	Slot  string
	Cover *metadata.Cover
	Hard  int64
}

// selectArtwork applies cover.prefer and the cover byte budgets. Album artwork
// is an explicit override and stays first in the candidate list.
func selectArtwork(embedded, sidecar, album, existing map[string]*metadata.Cover, cfg *config.Config) map[string]*metadata.Cover {
	out, _ := selectArtworkWithOptions(embedded, sidecar, album, existing, cfg, false)
	return out
}

func selectArtworkWithOptions(embedded, sidecar, album, existing map[string]*metadata.Cover, cfg *config.Config, ignoreHard bool) (map[string]*metadata.Cover, []artworkSkip) {
	if !cfg.Cover.Enabled {
		return map[string]*metadata.Cover{}, nil
	}
	out := make(map[string]*metadata.Cover)
	var skipped []artworkSkip
	for _, kind := range artworkKinds {
		var candidates []coverpkg.Candidate
		if c := album[kind]; c != nil {
			candidates = append(candidates, coverpkg.Candidate{Source: "album", Cover: c})
		}
		if c := embedded[kind]; c != nil {
			candidates = append(candidates, coverpkg.Candidate{Source: "embedded", Cover: c})
		}
		if c := sidecar[kind]; c != nil {
			candidates = append(candidates, coverpkg.Candidate{Source: "sidecar", Cover: c})
		}
		if c := existing[kind]; c != nil {
			candidates = append(candidates, coverpkg.Candidate{Source: "existing", Cover: c})
		}
		hard := cfg.Cover.HardMaxBytes
		if ignoreHard {
			hard = 0
		}
		if c, ok := coverpkg.Select(candidates, cfg.Cover.Prefer, cfg.Cover.MaxBytes, hard); ok {
			out[kind] = c.Cover
			continue
		}
		if !ignoreHard && hard > 0 {
			for _, c := range candidates {
				if c.Cover != nil && int64(len(c.Cover.Data)) > hard {
					skipped = append(skipped, artworkSkip{Slot: kind, Cover: c.Cover, Hard: hard})
					break
				}
			}
		}
	}
	return out, skipped
}

type artworkInfo struct {
	Filename string `json:"filename,omitempty"`
	MIME     string `json:"mime,omitempty"`
	Bytes    int    `json:"bytes,omitempty"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
}

func makeArtworkInfo(cover *metadata.Cover) *artworkInfo {
	if cover == nil || len(cover.Data) == 0 {
		return nil
	}
	info := &artworkInfo{Filename: cover.Filename, MIME: cover.MIME, Bytes: len(cover.Data)}
	if info.MIME == "" {
		info.MIME = http.DetectContentType(cover.Data)
	}
	if cfg, _, err := image.DecodeConfig(bytes.NewReader(cover.Data)); err == nil {
		info.Width, info.Height = cfg.Width, cfg.Height
	}
	return info
}

func artworkTemplateData(covers map[string]*metadata.Cover) map[string]string {
	return artworkTemplateDataFor("", covers)
}

func artworkTemplateDataFor(prefix string, covers map[string]*metadata.Cover) map[string]string {
	out := make(map[string]string)
	var found []string
	for _, kind := range artworkKinds {
		cover := covers[kind]
		if cover == nil {
			continue
		}
		found = append(found, fmt.Sprintf("%s:%s(%d)", kind, cover.Filename, len(cover.Data)))
		out[prefix+"artwork"+kind] = cover.Filename
		out[prefix+"artwork"+kind+"bytes"] = strconv.Itoa(len(cover.Data))
	}
	out[prefix+"artwork"] = strings.Join(found, "; ")
	out[prefix+"artworkcount"] = strconv.Itoa(len(found))
	return out
}

func applyCoverBudget(covers map[string]*metadata.Cover, cfg *config.Config) map[string]*metadata.Cover {
	out := make(map[string]*metadata.Cover)
	if !cfg.Cover.Enabled {
		return out
	}
	for _, kind := range artworkKinds {
		c := covers[kind]
		if c == nil {
			continue
		}
		if chosen, ok := coverpkg.Select([]coverpkg.Candidate{{Source: "existing", Cover: c}}, cfg.Cover.Prefer, cfg.Cover.MaxBytes, cfg.Cover.HardMaxBytes); ok {
			out[kind] = chosen.Cover
		}
	}
	return out
}
