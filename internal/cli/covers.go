package cli

import (
	"path/filepath"
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

func copyCovers(in map[string]*metadata.Cover) map[string]*metadata.Cover {
	out := make(map[string]*metadata.Cover, len(in))
	for kind, c := range in {
		out[kind] = c
	}
	return out
}

// selectArtwork applies cover.prefer and the cover byte budgets. Album artwork
// is an explicit override and stays first in the candidate list.
func selectArtwork(embedded, sidecar, album, existing map[string]*metadata.Cover, cfg *config.Config) map[string]*metadata.Cover {
	if !cfg.Cover.Enabled {
		return map[string]*metadata.Cover{}
	}
	out := make(map[string]*metadata.Cover)
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
		if c, ok := coverpkg.Select(candidates, cfg.Cover.Prefer, cfg.Cover.MaxBytes, cfg.Cover.HardMaxBytes); ok {
			out[kind] = c.Cover
		}
	}
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
