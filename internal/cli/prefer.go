package cli

import (
	"os"

	"github.com/daregod/ext-tags/internal/config"
)

// preferExisting reports whether external .tag fields should override
// container fields for this file.
func preferExisting(cfg *config.Config, path string) bool {
	switch cfg.Files.Prefer {
	case "tag":
		return true
	case "newest":
		audio, err := os.Stat(path)
		if err != nil {
			return false
		}
		tag, err := os.Stat(path + cfg.TagFile.Suffix)
		if err != nil {
			return false
		}
		return tag.ModTime().After(audio.ModTime())
	default:
		return false
	}
}
