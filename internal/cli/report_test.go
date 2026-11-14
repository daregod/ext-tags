package cli

import (
	"testing"
)

func TestReportDiffAndQuiet(t *testing.T) {
	diff := &diffResult{
		Fields:  []fieldDiff{{Field: "Title", From: "A", To: "B", Change: "changed"}, {Field: "X", To: "Y", Change: "added"}, {Field: "Z", From: "Q", Change: "removed"}},
		Artwork: []coverDiff{{Slot: "front", From: &artworkInfo{MIME: "image/jpeg", Bytes: 1}, To: &artworkInfo{MIME: "image/png", Bytes: 2}}},
	}
	cases := []fileResult{
		{File: "a", Status: "ok", TagFile: "t", Items: 1, Bytes: 2, Backup: "b"},
		{File: "a", Status: "dry-run", TagFile: "t"},
		{File: "a", Status: "diff", Diff: diff},
		{File: "a", Status: "rename-dry", Target: "b"},
		{File: "a", Status: "renamed", Target: "b"},
		{File: "a", Status: "imported", TagFile: "t", Backup: "b", Warning: "w"},
		{File: "a", Status: "import-dry", TagFile: "t", Warning: "w"},
		{File: "a", Status: "skipped", Reason: "r"},
		{File: "a", Status: "no-tags"},
		{File: "a", Status: "unsupported", Reason: "r"},
		{File: "a", Status: "error", Error: "e"},
	}
	for _, res := range cases {
		a := &app{opts: &options{}}
		captureStdout(t, func() { a.report(res) })
	}
	a := &app{opts: &options{json: true}}
	captureStdout(t, func() { a.report(fileResult{File: "a", Status: "ok"}) })
	a = &app{opts: &options{quiet: true}}
	captureStdout(t, func() { a.report(fileResult{File: "a", Status: "error", Error: "e"}) })
}

func TestWriteJSONRecordErrorAndNilCover(t *testing.T) {
	captureStdout(t, func() { writeJSONRecord(make(chan int)) })
	if got := formatCover(nil); got != "none" {
		t.Fatalf("nil cover = %q", got)
	}
}

func TestFieldNameBranches(t *testing.T) {
	if got := fieldName(map[string]string{"TITLE": "x"}, nil, "title"); got != "TITLE" {
		t.Fatalf("old = %q", got)
	}
	if got := fieldName(nil, map[string]string{"TITLE": "x"}, "title"); got != "TITLE" {
		t.Fatalf("new = %q", got)
	}
	if got := fieldName(nil, nil, "title"); got != "title" {
		t.Fatalf("fallback = %q", got)
	}
}
