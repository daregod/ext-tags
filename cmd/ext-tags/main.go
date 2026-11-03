package main

import (
	"fmt"
	"os"
	"path/filepath"

	"external-tags/internal/extractor"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <audio-file> [...]\n", os.Args[0])
		os.Exit(2)
	}

	code := 0
	for _, path := range os.Args[1:] {
		ext := filepath.Ext(path)
		if !extractor.SupportedExtensions[ext] {
			fmt.Fprintf(os.Stderr, "SKIP  %s (unsupported format %s)\n", path, ext)
			continue
		}

		if err := processFile(path); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR %s: %v\n", path, err)
			code = 1
		}
	}
	os.Exit(code)
}

func processFile(path string) error {
	t, err := extractor.Extract(path)
	if err != nil {
		return err
	}
	if len(t.Items) == 0 {
		fmt.Printf("SKIP  %s (no tags)\n", path)
		return nil
	}

	data, err := t.Serialize()
	if err != nil {
		return fmt.Errorf("serialize: %w", err)
	}

	out := path + ".tag"
	if err := os.WriteFile(out, data, 0644); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	fmt.Printf("OK    %s → %s (%d items, %d bytes)\n", path, out, len(t.Items), len(data))
	return nil
}
