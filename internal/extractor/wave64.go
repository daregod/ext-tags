package extractor

import (
	"fmt"

	"external-tags/internal/riff"
)

// extractWave64 reads Wave64 metadata: an ID3v2 chunk (and LIST/INFO when
// present).
func extractWave64(path string) (*Result, error) {
	md, err := riff.ReadWave64(path)
	if err != nil {
		return nil, fmt.Errorf("read wave64: %w", err)
	}
	return buildRIFF(md, "WAVE64")
}
