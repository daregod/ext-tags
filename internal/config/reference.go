package config

import (
	"sort"
	"strings"
)

// TagReference returns a commented reference for config init --full.
// It is generated from the built-in alias table so aliases cannot drift.
func TagReference() string {
	canonical := make(map[string]bool)
	aliases := make(map[string][]string)
	for raw, name := range defaultAliases {
		canonical[name] = true
		aliases[name] = append(aliases[name], raw)
	}
	for _, name := range []string{"Track", "Disc"} {
		canonical[name] = true
	}
	names := make([]string, 0, len(canonical))
	for name := range canonical {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return strings.ToLower(names[i]) < strings.ToLower(names[j])
	})

	var b strings.Builder
	b.WriteString("# Canonical field names (case-insensitive; output keeps this spelling).\n")
	b.WriteString("# Unknown container fields pass through unchanged unless fields.ignore matches.\n#\n")
	for _, name := range names {
		b.WriteString("#   " + name + "\n")
	}
	b.WriteString("#\n# Raw container tag -> canonical field:\n")
	for _, name := range names {
		raw := aliases[name]
		if len(raw) == 0 {
			continue
		}
		sort.Strings(raw)
		b.WriteString("#   " + strings.Join(raw, ", ") + " -> " + name + "\n")
	}
	b.WriteString("#\n# Artwork filter names for fields.include/exclude:\n")
	b.WriteString("#   Cover Art (Front)  (or the configured cover.field)\n")
	b.WriteString("#   Cover Art (Back)\n")
	b.WriteString("#   Artist Picture\n")
	b.WriteString("#   Cover Art (Disc)\n")
	b.WriteString("#   Cover Art (Icon)\n")
	b.WriteString("#\n# Ignored container fields by default:\n#   " + strings.Join(defaultIgnore, ", ") + "\n")
	return b.String()
}
