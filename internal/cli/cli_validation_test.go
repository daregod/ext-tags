package cli

import (
	"testing"

	"github.com/daregod/ext-tags/internal/config"
)

func TestExtractOutputValidation(t *testing.T) {
	cases := []options{
		{output: "x", stdout: true},
		{stdout: true, json: true},
		{stdout: true, dryRun: true},
		{stdout: true, summary: true},
	}
	for _, opts := range cases {
		a := &app{opts: &opts}
		if err := a.validateExtractOutput(); err == nil {
			t.Fatalf("expected error for %+v", opts)
		}
	}
	a := &app{opts: &options{output: "-"}}
	if err := a.validateExtractOutput(); err != nil || !a.opts.stdout || a.opts.output != "" {
		t.Fatalf("stdout redirect: %+v %v", a.opts, err)
	}
}

func TestApplyOverridesConflict(t *testing.T) {
	a := &app{opts: &options{}}
	cfg := config.Default()
	changed := func(name string) bool { return name == "merge" || name == "no-merge" }
	if err := a.applyOverrides(cfg, changed); err == nil {
		t.Fatal("expected conflict")
	}
}
