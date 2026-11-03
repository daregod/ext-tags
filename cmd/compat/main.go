// Command compat compares tag files produced by ext-tags with reference
// ".tag" files created by foo_external_tags (foobar2000).
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"external-tags/internal/ape"
	"external-tags/internal/config"
	"external-tags/internal/extractor"
	"external-tags/internal/rules"
	"external-tags/internal/tagfile"
)

func main() {
	examples := flag.String("examples", filepath.Join("bin", "examples"), "directory with audio files and reference .tag files")
	verbose := flag.Bool("v", false, "print values of differing fields")
	flag.Parse()

	code := run(*examples, *verbose)
	os.Exit(code)
}

func run(dir string, verbose bool) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "compat: %v\n", err)
		return 1
	}

	cfg := config.Default()
	engine, err := rules.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "compat: %v\n", err)
		return 1
	}

	var (
		compared, failed, skipped int
	)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		audio := filepath.Join(dir, entry.Name())
		if !extractor.IsSupported(audio, cfg.Files.Extensions) {
			continue
		}
		refPath := audio + cfg.TagFile.Suffix
		if _, err := os.Stat(refPath); err != nil {
			continue
		}

		ref, err := ape.ReadFile(refPath)
		if err != nil {
			fmt.Printf("ERROR %s: reference: %v\n", entry.Name(), err)
			failed++
			continue
		}

		md, err := extractor.Extract(audio)
		if errors.Is(err, extractor.ErrNoTags) {
			fmt.Printf("SKIP  %s (no tags in audio, reference has %s)\n", entry.Name(), keys(ref))
			skipped++
			continue
		}
		if err != nil {
			fmt.Printf("ERROR %s: %v\n", entry.Name(), err)
			failed++
			continue
		}

		base := filepath.Base(audio)
		meta := rules.Meta{
			Path:     audio,
			Filename: strings.TrimSuffix(base, filepath.Ext(base)),
			Dir:      filepath.Dir(audio),
			Ext:      strings.ToLower(filepath.Ext(audio)),
			Now:      time.Now(),
		}
		fields, err := engine.Build(md, meta)
		if err != nil {
			fmt.Printf("ERROR %s: rules: %v\n", entry.Name(), err)
			failed++
			continue
		}
		data, _, _, err := tagfile.Build(fields, md.Cover, nil, cfg)
		if err != nil {
			fmt.Printf("ERROR %s: build: %v\n", entry.Name(), err)
			failed++
			continue
		}
		gen, err := ape.Parse(data)
		if err != nil {
			fmt.Printf("ERROR %s: generated tag: %v\n", entry.Name(), err)
			failed++
			continue
		}

		compared++
		changed, cosmetic, refOnly, genOnly := compare(ref, gen)
		status := "OK   "
		if len(changed) > 0 {
			status = "DIFF "
			failed++
		}
		fmt.Printf("%s %-60s ref=%d gen=%d", status, entry.Name(), len(ref.Items), len(gen.Items))
		if len(changed) > 0 {
			fmt.Printf(" changed=%v", changed)
		}
		if len(cosmetic) > 0 {
			fmt.Printf(" cosmetic=%v", cosmetic)
		}
		if len(refOnly) > 0 {
			fmt.Printf(" ref-only=%v", refOnly)
		}
		if len(changed) > 0 && verbose {
			fmt.Printf("\n       ref: %s\n       gen: %s", describe(ref, changed), describe(gen, changed))
		}
		if len(genOnly) > 0 {
			fmt.Printf(" gen-only=%v", genOnly)
		}
		fmt.Println()
	}

	fmt.Printf("\ncompared=%d failed=%d skipped=%d\n", compared, failed, skipped)
	if failed > 0 {
		return 1
	}
	return 0
}

func compare(ref, gen *ape.ParsedTag) (changed, cosmetic, refOnly, genOnly []string) {
	refItems := index(ref)
	genItems := index(gen)

	for key, r := range refItems {
		g, ok := genItems[key]
		if !ok {
			refOnly = append(refOnly, r.Key)
			continue
		}
		switch valueState(r.Key, r.Value, g.Value) {
		case stateCosmetic:
			cosmetic = append(cosmetic, r.Key)
		case stateDifferent:
			changed = append(changed, r.Key)
		}
	}
	for key, g := range genItems {
		if _, ok := refItems[key]; !ok {
			genOnly = append(genOnly, g.Key)
		}
	}
	sort.Strings(changed)
	sort.Strings(cosmetic)
	sort.Strings(refOnly)
	sort.Strings(genOnly)
	return changed, cosmetic, refOnly, genOnly
}

const (
	stateEqual = iota
	stateCosmetic
	stateDifferent
)

// valueState compares two values. Differences that do not change the meaning
// (trimmed whitespace, " dB" unit on replaygain fields, cover file name) are
// reported as cosmetic.
func valueState(key string, a, b []byte) int {
	if bytes.Equal(a, b) {
		return stateEqual
	}
	if strings.EqualFold(key, "cover art (front)") {
		if bytes.Equal(coverImage(a), coverImage(b)) {
			return stateCosmetic
		}
		return stateDifferent
	}
	if normalizeText(key, a) == normalizeText(key, b) {
		return stateCosmetic
	}
	return stateDifferent
}

func normalizeText(key string, value []byte) string {
	text := strings.TrimSpace(string(value))
	if strings.HasPrefix(strings.ToLower(key), "replaygain_") {
		text = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(text), " dB"))
	}
	return text
}

func index(t *ape.ParsedTag) map[string]ape.Item {
	m := make(map[string]ape.Item, len(t.Items))
	for _, item := range t.Items {
		m[strings.ToLower(item.Key)] = item
	}
	return m
}

func coverImage(value []byte) []byte {
	if i := bytes.IndexByte(value, 0); i >= 0 && i+1 <= len(value) {
		return value[i+1:]
	}
	return value
}

func keys(t *ape.ParsedTag) []string {
	names := make([]string, 0, len(t.Items))
	for _, item := range t.Items {
		names = append(names, item.Key)
	}
	sort.Strings(names)
	return names
}

func describe(t *ape.ParsedTag, keys []string) string {
	var parts []string
	for _, key := range keys {
		if item, ok := t.Get(key); ok {
			value := item.Value
			if strings.EqualFold(key, "cover art (front)") {
				value = []byte(fmt.Sprintf("<image %d B>", len(coverImage(value))))
			}
			text := strings.ReplaceAll(string(value), "\n", "|")
			if len(text) > 120 {
				text = text[:120] + "..."
			}
			parts = append(parts, fmt.Sprintf("%s=%q", key, text))
		}
	}
	return strings.Join(parts, " ")
}
