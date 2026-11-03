package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDefaultValidate(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("default config must be valid: %v", err)
	}
	if cfg.Files.OnExisting != OnExistingBackup {
		t.Fatalf("on_existing = %q, want backup", cfg.Files.OnExisting)
	}
	if len(cfg.Fields.Rules) == 0 {
		t.Fatal("default config must define track/disc rules")
	}
}

func TestLoadWithoutFile(t *testing.T) {
	cfg, err := Load("", "")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Version != 1 {
		t.Fatalf("version = %d", cfg.Version)
	}
}

func TestLoadProfileMerge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ext-tags.yaml")
	data := []byte(`
version: 1
files:
  on_existing: skip
fields:
  aliases:
    MYKEY: MyField
profiles:
  flac:
    files:
      extensions: [.flac]
    fields:
      rules:
        - set: Album
          template: "{{ upper .album }}"
`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	base, err := Load(path, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(base.Files.Extensions) != 9 {
		t.Fatalf("base extensions = %v", base.Files.Extensions)
	}

	profiled, err := Load(path, "flac")
	if err != nil {
		t.Fatal(err)
	}
	if len(profiled.Files.Extensions) != 1 || profiled.Files.Extensions[0] != ".flac" {
		t.Fatalf("profile extensions = %v", profiled.Files.Extensions)
	}
	if profiled.Files.OnExisting != OnExistingSkip {
		t.Fatalf("profile lost base on_existing: %q", profiled.Files.OnExisting)
	}
	rules := profiled.Fields.Rules
	if len(rules) != 3 || rules[len(rules)-1].Set != "Album" {
		t.Fatalf("profile rules must append to defaults: %+v", rules)
	}
	if profiled.Fields.Aliases["MYKEY"] != "MyField" {
		t.Fatal("profile must keep base aliases")
	}
}

func TestLoadUnknownProfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ext-tags.yaml")
	if err := os.WriteFile(path, []byte("version: 1\nprofiles:\n  music: {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path, "video"); err == nil {
		t.Fatal("expected error for unknown profile")
	}
}

func TestValidateErrors(t *testing.T) {
	cases := []struct {
		name string
		mod  func(*Config)
	}{
		{"bad version", func(c *Config) { c.Version = 2 }},
		{"bad on_existing", func(c *Config) { c.Files.OnExisting = "merge" }},
		{"bad suffix", func(c *Config) { c.TagFile.Suffix = "tag" }},
		{"bad extension", func(c *Config) { c.Files.Extensions = []string{"mp3"} }},
		{"empty extensions", func(c *Config) { c.Files.Extensions = nil }},
		{"two operations", func(c *Config) {
			c.Fields.Rules = []Rule{{Set: "A", Template: "x", Copy: &CopyRule{From: "X", To: "Y"}}}
		}},
		{"set without template", func(c *Config) {
			c.Fields.Rules = []Rule{{Set: "A"}}
		}},
		{"default without from", func(c *Config) {
			c.Fields.Rules = []Rule{{Default: &DefaultRule{Field: "A"}}}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := Default()
			c.mod(cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestMarshalRoundTrip(t *testing.T) {
	data, err := Default().Marshal()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "ext-tags.yaml")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path, ""); err != nil {
		t.Fatalf("marshalled default config must load: %v", err)
	}
}

func TestMinimalExampleMatchesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ext-tags.yaml")
	if err := os.WriteFile(path, MinimalExample(), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path, "")
	if err != nil {
		t.Fatalf("minimal config must load: %v", err)
	}
	if !reflect.DeepEqual(cfg, Default()) {
		t.Fatalf("minimal config differs from defaults:\n got: %#v\nwant: %#v", cfg, Default())
	}
}

func TestLoadLayersExplicit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom.yaml")
	if err := os.WriteFile(path, []byte("version: 1\nfiles:\n  on_existing: skip\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, layers, err := LoadLayers("", path)
	if err != nil {
		t.Fatal(err)
	}
	if layers.Explicit != path || layers.User != "" || layers.Project != "" {
		t.Fatalf("layers = %+v", layers)
	}
	if cfg.Files.OnExisting != OnExistingSkip {
		t.Fatalf("on_existing = %q", cfg.Files.OnExisting)
	}
}

func TestLoadFilesLayering(t *testing.T) {
	dir := t.TempDir()
	userPath := filepath.Join(dir, "user.yaml")
	projectPath := filepath.Join(dir, "project.yaml")
	user := []byte(`
version: 1
files:
  on_existing: skip
fields:
  aliases:
    MYKEY: MyField
  rules:
    - set: Sort
      template: "{{ .title }}"
`)
	project := []byte(`
version: 1
files:
  extensions: [.flac]
fields:
  aliases:
    OTHER: OtherField
  rules:
    - set: Album
      template: "{{ upper .album }}"
`)
	if err := os.WriteFile(userPath, user, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectPath, project, 0644); err != nil {
		t.Fatal(err)
	}

	cfg := Default()
	if err := loadFiles(cfg, []string{userPath, projectPath}, ""); err != nil {
		t.Fatal(err)
	}

	if len(cfg.Files.Extensions) != 1 || cfg.Files.Extensions[0] != ".flac" {
		t.Fatalf("extensions = %v, want project override", cfg.Files.Extensions)
	}
	if cfg.Files.OnExisting != OnExistingSkip {
		t.Fatalf("on_existing = %q, want value from user layer", cfg.Files.OnExisting)
	}
	if cfg.Fields.Aliases["MYKEY"] != "MyField" || cfg.Fields.Aliases["OTHER"] != "OtherField" {
		t.Fatalf("aliases must be merged: %v", cfg.Fields.Aliases)
	}
	rules := cfg.Fields.Rules
	if len(rules) != 4 || rules[2].Set != "Sort" || rules[3].Set != "Album" {
		t.Fatalf("rules must accumulate across layers: %+v", rules)
	}
}

func TestLoadFilesProfileFromProject(t *testing.T) {
	dir := t.TempDir()
	userPath := filepath.Join(dir, "user.yaml")
	projectPath := filepath.Join(dir, "project.yaml")
	if err := os.WriteFile(userPath, []byte("version: 1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectPath, []byte(`
version: 1
profiles:
  flac:
    files:
      extensions: [.flac]
`), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := Default()
	if err := loadFiles(cfg, []string{userPath, projectPath}, "flac"); err != nil {
		t.Fatal(err)
	}
	if len(cfg.Files.Extensions) != 1 || cfg.Files.Extensions[0] != ".flac" {
		t.Fatalf("profile not applied: %v", cfg.Files.Extensions)
	}
}

func TestRulesModeReplace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ext-tags.yaml")
	data := []byte(`
version: 1
fields:
  rules_mode: replace
  rules:
    - set: Sort
      template: "{{ .title }}"
`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Fields.Rules) != 1 || cfg.Fields.Rules[0].Set != "Sort" {
		t.Fatalf("rules_mode replace must drop defaults: %+v", cfg.Fields.Rules)
	}
}

func TestRulesModeInvalid(t *testing.T) {
	cfg := Default()
	cfg.Fields.RulesMode = "merge"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for rules_mode")
	}
}

func TestWarningsMergeWithSkip(t *testing.T) {
	cfg := Default()
	cfg.Files.OnExisting = OnExistingSkip
	if len(cfg.Warnings()) != 1 {
		t.Fatalf("expected warning for merge+skip, got %v", cfg.Warnings())
	}
	cfg.Files.OnExisting = OnExistingBackup
	if len(cfg.Warnings()) != 0 {
		t.Fatalf("unexpected warnings: %v", cfg.Warnings())
	}
}
