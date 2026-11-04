package importer

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/daregod/ext-tags/internal/metadata"
	"github.com/daregod/ext-tags/internal/tagformat"
)

// fileOps is the narrow filesystem surface used by the atomic import path.
type fileOps interface {
	Stat(name string) (fs.FileInfo, error)
	CreateTemp(dir, pattern string) (string, io.Closer, error)
	Copy(src, dst string) error
	Remove(name string) error
	Rename(oldpath, newpath string) error
	Chmod(name string, mode fs.FileMode) error
}

type osOps struct{}

func (osOps) Stat(name string) (fs.FileInfo, error) { return os.Stat(name) }
func (osOps) CreateTemp(dir, pattern string) (string, io.Closer, error) {
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", nil, err
	}
	return f.Name(), f, nil
}
func (osOps) Copy(src, dst string) error { return copyFile(src, dst) }
func (osOps) Remove(name string) error   { return os.Remove(name) }
func (osOps) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}
func (osOps) Chmod(name string, mode fs.FileMode) error { return os.Chmod(name, mode) }

// WriteAtomic writes through a temporary file, verifies the result with a
// reader, and only then replaces the original.
func WriteAtomic(path string, writer tagformat.Writer, reader tagformat.Reader, doc *metadata.Document, opts tagformat.WriteOptions) error {
	return writeAtomic(path, writer, reader, doc, opts, osOps{})
}

func writeAtomic(path string, writer tagformat.Writer, reader tagformat.Reader, doc *metadata.Document, opts tagformat.WriteOptions, ops fileOps) error {
	backup := opts.Backup
	opts.Backup = false
	if backup {
		if err := ops.Copy(path, path+".bak"); err != nil {
			return fmt.Errorf("backup: %w", err)
		}
	}
	info, err := ops.Stat(path)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	tmpPath, closer, err := ops.CreateTemp(dir, "."+base+".tmp-*"+filepath.Ext(path))
	if err != nil {
		return err
	}
	if err := closer.Close(); err != nil {
		_ = ops.Remove(tmpPath)
		return err
	}
	defer func() { _ = ops.Remove(tmpPath) }()

	if err := ops.Copy(path, tmpPath); err != nil {
		return fmt.Errorf("prepare temp: %w", err)
	}
	if err := writer.Write(tmpPath, doc, opts); err != nil {
		return err
	}
	got, err := reader.Read(tmpPath, tagformat.ReadOptions{})
	if err != nil {
		return fmt.Errorf("verify: %w", err)
	}
	if err := verifyFields(doc, got); err != nil {
		return fmt.Errorf("verify: %w", err)
	}
	if err := ops.Chmod(tmpPath, info.Mode().Perm()); err != nil {
		return err
	}
	if err := ops.Rename(tmpPath, path); err != nil {
		if removeErr := ops.Remove(path); removeErr != nil {
			return fmt.Errorf("replace: %w", err)
		}
		if err := ops.Rename(tmpPath, path); err != nil {
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
