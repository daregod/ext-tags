package importer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/daregod/ext-tags/internal/format"
	"github.com/daregod/ext-tags/internal/metadata"
)

// WriteAtomic writes through a temporary file, verifies the result with a
// reader, and only then replaces the original.
func WriteAtomic(path string, writer format.Writer, reader format.Reader, doc *metadata.Document, opts format.WriteOptions) error {
	backup := opts.Backup
	opts.Backup = false
	if backup {
		if err := copyFile(path, path+".bak"); err != nil {
			return fmt.Errorf("backup: %w", err)
		}
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	tmp, err := os.CreateTemp(dir, "."+base+".tmp-*"+filepath.Ext(path))
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	defer os.Remove(tmpPath)

	if err := copyFile(path, tmpPath); err != nil {
		return fmt.Errorf("prepare temp: %w", err)
	}
	if err := writer.Write(tmpPath, doc, opts); err != nil {
		return err
	}
	got, err := reader.Read(tmpPath, format.ReadOptions{})
	if err != nil {
		return fmt.Errorf("verify: %w", err)
	}
	if err := verifyFields(doc, got); err != nil {
		return fmt.Errorf("verify: %w", err)
	}
	if err := os.Chmod(tmpPath, info.Mode().Perm()); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		if removeErr := os.Remove(path); removeErr != nil {
			return fmt.Errorf("replace: %w", err)
		}
		if err := os.Rename(tmpPath, path); err != nil {
			return fmt.Errorf("replace: %w", err)
		}
	}
	return nil
}

func verifyFields(want, got *metadata.Document) error {
	if want == nil || got == nil {
		return nil
	}
	for field, value := range want.Fields {
		if value == "" || field == "Track" || field == "Disc" {
			continue
		}
		switch field {
		case "Title", "Artist", "Album", "Genre", "Year":
		default:
			continue
		}
		if matchesField(got, field, value) {
			continue
		}
		return fmt.Errorf("field %s is not readable after write", field)
	}
	return nil
}

func matchesField(doc *metadata.Document, field, want string) bool {
	names := []string{field}
	switch field {
	case "Artist":
		names = append(names, "AlbumArtist")
	case "Year":
		names = append(names, "Date")
	}
	for _, name := range names {
		for key, value := range doc.Fields {
			if strings.EqualFold(key, name) && strings.EqualFold(value, want) {
				return true
			}
		}
	}
	return false
}
