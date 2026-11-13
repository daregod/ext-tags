package cli

import "fmt"

type runStats struct {
	processed   int
	ok          int
	dryRun      int
	diff        int
	renamed     int
	imported    int
	skipped     int
	noTags      int
	unsupported int
	errors      int
	warnings    int
	items       int
	bytes       int
}

func (a *app) addResult(r fileResult) {
	s := &a.stats
	s.processed++
	switch r.Status {
	case "ok":
		s.ok++
		s.items += r.Items
		s.bytes += r.Bytes
	case "dry-run", "rename-dry", "import-dry":
		s.dryRun++
	case "diff":
		s.diff++
	case "renamed":
		s.renamed++
	case "imported":
		s.imported++
		s.items += r.Items
		s.bytes += r.Bytes
	case "skipped":
		s.skipped++
	case "no-tags":
		s.noTags++
	case "unsupported":
		s.unsupported++
	case "error":
		s.errors++
	}
	if r.Warning != "" {
		s.warnings++
	}
}

type summaryRecord struct {
	Type        string `json:"type"`
	Processed   int    `json:"processed"`
	OK          int    `json:"ok"`
	DryRun      int    `json:"dry_run"`
	Diff        int    `json:"diff"`
	Renamed     int    `json:"renamed"`
	Imported    int    `json:"imported"`
	Skipped     int    `json:"skipped"`
	NoTags      int    `json:"no_tags"`
	Unsupported int    `json:"unsupported"`
	Errors      int    `json:"errors"`
	Warnings    int    `json:"warnings"`
	Items       int    `json:"items"`
	Bytes       int    `json:"bytes"`
}

func (a *app) printSummary() {
	if !a.opts.summary {
		return
	}
	s := a.stats
	if a.opts.json {
		writeJSONRecord(summaryRecord{
			Type: "summary", Processed: s.processed, OK: s.ok, DryRun: s.dryRun,
			Diff: s.diff, Renamed: s.renamed, Imported: s.imported, Skipped: s.skipped,
			NoTags: s.noTags, Unsupported: s.unsupported, Errors: s.errors,
			Warnings: s.warnings, Items: s.items, Bytes: s.bytes,
		})
		return
	}
	fmt.Printf("SUMMARY processed=%d ok=%d dry_run=%d diff=%d renamed=%d imported=%d skipped=%d no_tags=%d unsupported=%d errors=%d warnings=%d items=%d bytes=%d\n",
		s.processed, s.ok, s.dryRun, s.diff, s.renamed, s.imported, s.skipped,
		s.noTags, s.unsupported, s.errors, s.warnings, s.items, s.bytes)
}
