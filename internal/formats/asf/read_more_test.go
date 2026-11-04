package asf

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestReadMissing(t *testing.T) {
	if _, err := Read(filepath.Join(t.TempDir(), "missing.wma")); err == nil {
		t.Fatal("missing read must fail")
	}
}

func TestBuildExtendedContentOversize(t *testing.T) {
	fields := map[string]string{"Title": strings.Repeat("x", 70000)}
	if got := buildExtendedContent(fields, nil, nil, true); got != nil {
		t.Fatalf("oversize field must be skipped, got %d bytes", len(got))
	}
}
