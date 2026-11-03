// Package config defines the ext-tags configuration schema, built-in
// defaults, layered loading and validation.
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// On-existing policies for the output .tag file.
const (
	OnExistingOverwrite = "overwrite"
	OnExistingSkip      = "skip"
	OnExistingBackup    = "backup"
	OnExistingError     = "error"
)

// EnvConfig is the environment variable used when --config is not given.
const EnvConfig = "EXT_TAGS_CONFIG"

// Config is the effective configuration of a single run.
type Config struct {
	Version int           `yaml:"version"`
	Files   FilesConfig   `yaml:"files"`
	TagFile TagFileConfig `yaml:"tagfile"`
	Cover   CoverConfig   `yaml:"cover"`
	Fields  FieldsConfig  `yaml:"fields"`
}

// FilesConfig controls input discovery and overwrite behaviour.
type FilesConfig struct {
	Extensions    []string `yaml:"extensions"`
	Recursive     bool     `yaml:"recursive"`
	OnExisting    string   `yaml:"on_existing"`
	MergeExisting bool     `yaml:"merge_existing"`
}

// TagFileConfig controls the name of the produced tag file.
type TagFileConfig struct {
	Suffix string `yaml:"suffix"`
}

// CoverConfig controls embedded cover art.
type CoverConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Field    string `yaml:"field"`
	Filename string `yaml:"filename"`
}

// FieldsConfig controls how container fields become tag fields.
type FieldsConfig struct {
	Aliases map[string]string `yaml:"aliases"`
	Ignore  []string          `yaml:"ignore"`
	Rules   []Rule            `yaml:"rules"`
	Include []string          `yaml:"include"`
	Exclude []string          `yaml:"exclude"`
	// RulesMode is "append" (default) or "replace".
	RulesMode string `yaml:"rules_mode,omitempty"`
}

// Rule is a single field transformation. Exactly one operation must be set.
type Rule struct {
	Set      string       `yaml:"set,omitempty"`
	Template string       `yaml:"template,omitempty"`
	Copy     *CopyRule    `yaml:"copy,omitempty"`
	Rename   *RenameRule  `yaml:"rename,omitempty"`
	Default  *DefaultRule `yaml:"default,omitempty"`
	Remove   []string     `yaml:"remove,omitempty"`
	When     string       `yaml:"when,omitempty"`
}

// CopyRule duplicates an existing field.
type CopyRule struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

// RenameRule moves a field to a new name.
type RenameRule struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

// DefaultRule fills a field from another one when it is empty.
type DefaultRule struct {
	Field string `yaml:"field"`
	From  string `yaml:"from"`
}

// Layers describes which configuration files were used, in application order.
type Layers struct {
	Explicit string
	User     string
	Project  string
}

// Empty reports whether only built-in defaults are used.
func (l Layers) Empty() bool {
	return l.Explicit == "" && l.User == "" && l.Project == ""
}

// Paths returns the applied files in order.
func (l Layers) Paths() []string {
	var paths []string
	if l.Explicit != "" {
		paths = append(paths, l.Explicit)
	}
	if l.User != "" {
		paths = append(paths, l.User)
	}
	if l.Project != "" {
		paths = append(paths, l.Project)
	}
	return paths
}

// Label is a human-readable description of the applied layers.
func (l Layers) Label() string {
	if l.Empty() {
		return "built-in defaults"
	}
	return strings.Join(l.Paths(), " + ")
}

// defaultAliases maps raw container keys to foobar2000 field names.
// Keys are matched case-insensitively; MP4 atoms are normalised by the
// extractor to start with "©".
var defaultAliases = map[string]string{
	"TIT2": "Title", "TPE1": "Artist", "TALB": "Album", "TPE2": "Album artist",
	"TCON": "Genre", "TDRC": "Year", "TYER": "Year",
	"TCOM": "Composer", "COMM": "Comment", "TPUB": "Publisher",
	"TCOP": "Copyright", "TSSE": "Encoder", "TLAN": "Language", "TENC": "Encoded by",
	"TBPM": "BPM", "TKEY": "Key", "TEXT": "Lyricist", "TMED": "Media",

	"TITLE": "Title", "ARTIST": "Artist", "ALBUM": "Album",
	"ALBUMARTIST": "Album artist", "ALBUM ARTIST": "Album artist",
	"DATE": "Year", "YEAR": "Year", "GENRE": "Genre",
	"COMPOSER": "Composer", "PERFORMER": "Performer",
	"COMMENT": "Comment", "RATING": "Rating",

	"CATALOGNUMBER":          "Catalognumber",
	"CATALOG NUMBER":         "Catalognumber",
	"COMMENT CATALOG NUMBER": "Comment catalog number",
	"COPYRIGHT":              "Copyright",
	"ISRC":                   "Isrc",
	"LABEL":                  "Label",
	"ORGANIZATION":           "Publisher",
	"PUBLISHER":              "Publisher",
	"MUSICBRAINZ_ALBUMID":    "Musicbrainz_albumid",
	"MUSICBRAINZ_TRACKID":    "Musicbrainz_trackid",
	"RELEASETIME":            "Releasetime",
	"UPC":                    "Upc",
	"WWW":                    "Www",
	"REPLAYGAIN_TRACK_GAIN":  "replaygain_track_gain",
	"REPLAYGAIN_TRACK_PEAK":  "replaygain_track_peak",

	// MP4/iTunes atoms (the leading U+00A9 is added by the extractor).
	"©nam": "Title", "©ART": "Artist", "©alb": "Album", "©gen": "Genre",
	"©wrt": "Composer", "©cmt": "Comment", "©day": "Year",
	"aART": "Album artist", "cpil": "Itunescompilation", "tmpo": "Tempo",
}

// defaultIgnore lists container fields that never become tag fields.
var defaultIgnore = []string{
	"vendor", "iTunNORM", "iTunSMPB", "Encoding Params", "covr", "©too",
	"TRACKNUMBER", "TRACKTOTAL", "TOTALTRACKS", "TRACKS", "TOTALTRACK", "TRCK",
	"DISCNUMBER", "DISCTOTAL", "TOTALDISCS", "DISCS", "TOTALDISC", "TPOS",
}

// Default returns the built-in configuration used when no file is present.
func Default() *Config {
	aliases := make(map[string]string, len(defaultAliases))
	for k, v := range defaultAliases {
		aliases[k] = v
	}

	return &Config{
		Version: 1,
		Files: FilesConfig{
			Extensions:    []string{".mp3", ".flac", ".opus", ".ogg", ".m4a", ".mp4", ".wav", ".aif", ".aiff"},
			Recursive:     false,
			OnExisting:    OnExistingBackup,
			MergeExisting: true,
		},
		TagFile: TagFileConfig{Suffix: ".tag"},
		Cover: CoverConfig{
			Enabled:  true,
			Field:    "Cover Art (Front)",
			Filename: "cover",
		},
		Fields: FieldsConfig{
			Aliases: aliases,
			Ignore:  append([]string(nil), defaultIgnore...),
			Rules: []Rule{
				{Set: "Track", Template: `{{ .tracknumber }}{{ if .totaltracks }}/{{ .totaltracks }}{{ end }}`},
				{Set: "Disc", Template: `{{ .discnumber }}{{ if .disctotal }}/{{ .disctotal }}{{ end }}`},
			},
		},
	}
}

// MinimalExample returns a small, commented configuration with the most
// commonly changed keys. Omitted keys keep their built-in defaults.
func MinimalExample() []byte {
	return []byte(minimalYAML)
}

// LoadLayers resolves and applies the configuration layers:
// built-in defaults, then the user config, then the project config.
// An explicit file (flag or env) replaces both discovered layers.
func LoadLayers(profile, explicit string) (*Config, Layers, error) {
	cfg := Default()
	layers := ResolveLayers(explicit)

	if err := loadFiles(cfg, layers.Paths(), profile); err != nil {
		return nil, layers, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, layers, fmt.Errorf("%s: %w", layers.Label(), err)
	}
	return cfg, layers, nil
}

// ResolveLayers finds the configuration files without parsing them.
func ResolveLayers(explicit string) Layers {
	var layers Layers
	if explicit != "" {
		layers.Explicit = explicit
		return layers
	}
	if path, ok := UserPath(); ok {
		layers.User = path
	}
	if path, ok := ProjectPath(); ok {
		layers.Project = path
	}
	return layers
}

// UserPath returns the per-user configuration file, if it exists.
func UserPath() (string, bool) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", false
	}
	path := filepath.Join(dir, "ext-tags", "config.yaml")
	if _, err := os.Stat(path); err == nil {
		return path, true
	}
	return "", false
}

// ProjectPath returns the configuration file in the current directory.
func ProjectPath() (string, bool) {
	for _, name := range []string{"ext-tags.yaml", "ext-tags.yml", ".ext-tags.yaml"} {
		if _, err := os.Stat(name); err != nil {
			continue
		}
		if abs, err := filepath.Abs(name); err == nil {
			return abs, true
		}
		return name, true
	}
	return "", false
}

// Load applies a single configuration file over the defaults.
func Load(path, profile string) (*Config, error) {
	cfg := Default()
	if path == "" {
		if profile != "" {
			return nil, fmt.Errorf("profile %q requested, but no config file was given", profile)
		}
		return cfg, nil
	}
	if err := loadFiles(cfg, []string{path}, profile); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return cfg, nil
}

func loadFiles(cfg *Config, files []string, profile string) error {
	for _, file := range files {
		if err := loadFile(cfg, file); err != nil {
			return err
		}
	}
	if profile == "" {
		return nil
	}

	// The profile is looked up from the highest-priority file downwards.
	for i := len(files) - 1; i >= 0; i-- {
		applied, err := applyProfile(cfg, files[i], profile)
		if err != nil {
			return err
		}
		if applied {
			return nil
		}
	}

	var names []string
	for _, file := range files {
		found, err := profileNames(file)
		if err != nil {
			return err
		}
		names = append(names, found...)
	}
	if len(names) == 0 {
		return fmt.Errorf("profile %q requested, but no config file defines profiles", profile)
	}
	return fmt.Errorf("unknown profile %q (available: %s)", profile, strings.Join(names, ", "))
}

func loadFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	return mergeInto(cfg, data, path)
}

// mergeInto applies one YAML layer on top of cfg. Unlike plain yaml.Unmarshal
// it keeps rules additive: a layer appends its rules to the lower layers
// unless it sets fields.rules_mode: replace.
func mergeInto(cfg *Config, data []byte, label string) error {
	prevRules := append([]Rule(nil), cfg.Fields.Rules...)

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parse %s: %w", label, err)
	}

	var probe struct {
		Fields struct {
			Rules     *[]Rule `yaml:"rules"`
			RulesMode string  `yaml:"rules_mode"`
		} `yaml:"fields"`
	}
	if err := yaml.Unmarshal(data, &probe); err != nil {
		return fmt.Errorf("parse %s: %w", label, err)
	}

	if probe.Fields.Rules == nil {
		cfg.Fields.Rules = prevRules
		return nil
	}
	if cfg.Fields.RulesMode == "replace" {
		return nil
	}
	cfg.Fields.Rules = append(prevRules, cfg.Fields.Rules...)
	return nil
}

func applyProfile(cfg *Config, path, name string) (bool, error) {
	node, ok, err := profileNode(path, name)
	if err != nil || !ok {
		return false, err
	}
	// Unmarshal the profile over the already populated config: keys the
	// profile does not mention keep their current values.
	override, err := yaml.Marshal(node)
	if err != nil {
		return false, fmt.Errorf("profile %q: %w", name, err)
	}
	if err := mergeInto(cfg, override, fmt.Sprintf("profile %q in %s", name, path)); err != nil {
		return false, err
	}
	return true, nil
}

func profileNode(path, name string) (yaml.Node, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return yaml.Node{}, false, fmt.Errorf("read config: %w", err)
	}
	var raw struct {
		Profiles map[string]yaml.Node `yaml:"profiles"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return yaml.Node{}, false, fmt.Errorf("parse %s: %w", path, err)
	}
	node, ok := raw.Profiles[name]
	return node, ok, nil
}

func profileNames(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var raw struct {
		Profiles map[string]yaml.Node `yaml:"profiles"`
	}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	names := make([]string, 0, len(raw.Profiles))
	for name := range raw.Profiles {
		names = append(names, name)
	}
	return names, nil
}

// Warnings returns non-fatal configuration notes.
func (c *Config) Warnings() []string {
	var warnings []string
	if c.Files.MergeExisting && (c.Files.OnExisting == OnExistingSkip || c.Files.OnExisting == OnExistingError) {
		warnings = append(warnings, fmt.Sprintf(
			"files.merge_existing has no effect with files.on_existing: %s (nothing is written)",
			c.Files.OnExisting))
	}
	return warnings
}

// Validate checks the configuration for structural errors.
func (c *Config) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("unsupported config version %d (expected 1)", c.Version)
	}
	if len(c.Files.Extensions) == 0 {
		return fmt.Errorf("files.extensions is empty")
	}
	for _, ext := range c.Files.Extensions {
		if !strings.HasPrefix(ext, ".") || ext != strings.ToLower(ext) {
			return fmt.Errorf("files.extensions: %q must be lowercase and start with a dot", ext)
		}
	}
	switch c.Files.OnExisting {
	case OnExistingOverwrite, OnExistingSkip, OnExistingBackup, OnExistingError:
	default:
		return fmt.Errorf("files.on_existing: unknown value %q (want overwrite, skip, backup or error)", c.Files.OnExisting)
	}
	if !strings.HasPrefix(c.TagFile.Suffix, ".") {
		return fmt.Errorf("tagfile.suffix %q must start with a dot", c.TagFile.Suffix)
	}
	if c.Cover.Enabled && c.Cover.Field == "" {
		return fmt.Errorf("cover.field is empty")
	}
	switch c.Fields.RulesMode {
	case "", "append", "replace":
	default:
		return fmt.Errorf("fields.rules_mode: unknown value %q (want append or replace)", c.Fields.RulesMode)
	}
	for i, r := range c.Fields.Rules {
		if err := r.validate(); err != nil {
			return fmt.Errorf("fields.rules[%d]: %w", i, err)
		}
	}
	return nil
}

func (r Rule) validate() error {
	ops := 0
	if r.Set != "" {
		ops++
		if r.Template == "" {
			return fmt.Errorf("set %q requires template", r.Set)
		}
	}
	if r.Copy != nil {
		ops++
		if r.Copy.From == "" || r.Copy.To == "" {
			return fmt.Errorf("copy requires from and to")
		}
	}
	if r.Rename != nil {
		ops++
		if r.Rename.From == "" || r.Rename.To == "" {
			return fmt.Errorf("rename requires from and to")
		}
	}
	if r.Default != nil {
		ops++
		if r.Default.Field == "" || r.Default.From == "" {
			return fmt.Errorf("default requires field and from")
		}
	}
	if len(r.Remove) > 0 {
		ops++
	}
	if ops != 1 {
		return fmt.Errorf("exactly one of set, copy, rename, default or remove is required")
	}
	return nil
}

// Marshal renders the configuration as YAML (2-space indent).
func (c *Config) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(c); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

const minimalYAML = `# ext-tags configuration (minimal)
#
# Only the keys you actually change need to be here: everything else keeps
# the built-in defaults. Use ` + "`ext-tags config show`" + ` to see the effective
# configuration and ` + "`ext-tags config init --full`" + ` for a complete dump.

version: 1

files:
  # overwrite | skip | backup | error
  on_existing: backup
  # keep fields from an existing .tag that are absent in the audio file
  # (for example Rating, which foobar2000 keeps in its media library)
  merge_existing: true
  # recurse into subdirectories when a directory is passed
  recursive: false
  # extensions: [.mp3, .flac, .m4a]

cover:
  # embed the front cover into the .tag file
  enabled: true

# fields:
#   aliases:                    # container key -> foobar2000 field
#     ORGANIZATION: Publisher
#   ignore: [vendor, iTunNORM]
#   rules_mode: append          # append (default) or replace
#   rules:                      # applied after the built-in rules
#     - set: Track
#       template: '{{ .tracknumber }}{{ if .totaltracks }}/{{ .totaltracks }}{{ end }}'
#     - copy: {from: Label, to: Publisher}
#     - default: {field: "Album artist", from: Artist}
#   include: []
#   exclude: []
#
# profiles:
#   music:
#     files:
#       extensions: [.flac, .mp3]
`
