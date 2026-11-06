package cli

import (
	"testing"

	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/metadata"
)

func coverBytes(n int) *metadata.Cover {
	return &metadata.Cover{Filename: "cover.jpg", MIME: "image/jpeg", Data: make([]byte, n)}
}

func TestSelectArtworkBudgetAndFallback(t *testing.T) {
	cfg := config.Default()
	cfg.Cover.Prefer = "sidecar"
	cfg.Cover.MaxBytes = 2 << 20
	cfg.Cover.HardMaxBytes = 8 << 20

	got := selectArtwork(
		map[string]*metadata.Cover{"front": coverBytes(3 << 20)},
		map[string]*metadata.Cover{"front": coverBytes(1 << 20)},
		nil, nil, cfg,
	)
	if len(got["front"].Data) != 1<<20 {
		t.Fatalf("expected sidecar, got %d", len(got["front"].Data))
	}

	got = selectArtwork(
		map[string]*metadata.Cover{"front": coverBytes(1 << 20)},
		map[string]*metadata.Cover{"front": coverBytes(3 << 20)},
		nil, nil, cfg,
	)
	if len(got["front"].Data) != 1<<20 {
		t.Fatalf("expected embedded fallback, got %d", len(got["front"].Data))
	}

	cfg.Cover.HardMaxBytes = 2 << 20
	got = selectArtwork(
		map[string]*metadata.Cover{"front": coverBytes(3 << 20)},
		map[string]*metadata.Cover{"front": coverBytes(4 << 20)},
		nil, nil, cfg,
	)
	if got["front"] != nil {
		t.Fatal("hard limit must remove all candidates")
	}
}

func TestApplyCoverBudgetImport(t *testing.T) {
	cfg := config.Default()
	cfg.Cover.MaxBytes = 1 << 20
	cfg.Cover.HardMaxBytes = 2 << 20
	got := applyCoverBudget(map[string]*metadata.Cover{"front": coverBytes(3 << 20)}, cfg)
	if got["front"] != nil {
		t.Fatal("oversized cover survived import budget")
	}
}
