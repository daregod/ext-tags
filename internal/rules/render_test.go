package rules

import (
	"testing"
	"time"

	"github.com/daregod/ext-tags/internal/config"
)

func TestRenderVariants(t *testing.T) {
	engine, err := New(config.Default())
	if err != nil {
		t.Fatal(err)
	}
	meta := Meta{Path: "track.flac", Filename: "track", Dir: ".", Ext: ".flac", Now: time.Now()}
	fields := map[string]string{"Title": "Song", "Artist": "Artist"}
	vars := map[string]string{"tracknumber": "1"}
	got, err := engine.Render("{{.title}} - {{.artist}} ({{.tracknumber}})", fields, vars, meta)
	if err != nil || got != "Song - Artist (1)" {
		t.Fatalf("render = %q %v", got, err)
	}
	got, err = engine.RenderWithExtra("{{.artwork}}", fields, vars, map[string]string{"artwork": "cover.jpg"}, meta)
	if err != nil || got != "cover.jpg" {
		t.Fatalf("extra = %q %v", got, err)
	}
	got, err = engine.RenderRaw("{{.container.TITLE}}/{{.tag.TITLE}}", map[string]string{"TITLE": "C"}, map[string]string{"TITLE": "T"}, vars, meta)
	if err != nil || got != "C/T" {
		t.Fatalf("raw = %q %v", got, err)
	}
	got, err = engine.RenderRawWithExtra("{{.extra}}", nil, nil, nil, map[string]string{"extra": "x"}, meta)
	if err != nil || got != "x" {
		t.Fatalf("raw extra = %q %v", got, err)
	}
	if _, err := engine.Render("{{", nil, nil, meta); err == nil {
		t.Fatal("expected template error")
	}
	if _, err := engine.RenderWithExtra("{{", nil, nil, nil, meta); err == nil {
		t.Fatal("expected render extra error")
	}
	if _, err := engine.RenderRaw("{{", nil, nil, nil, meta); err == nil {
		t.Fatal("expected raw error")
	}
	if _, err := engine.RenderRawWithExtra("{{", nil, nil, nil, nil, meta); err == nil {
		t.Fatal("expected raw extra error")
	}
}
