// Package rules turns raw container metadata into tag fields according to
// the configuration: aliases, ignored fields, templates and copy/rename
// rules.
package rules

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/daregod/ext-tags/internal/config"
	"github.com/daregod/ext-tags/internal/metadata"
)

// Meta is per-file context available to templates.
type Meta struct {
	Path     string
	Filename string
	Dir      string
	Ext      string
	Now      time.Time
}

// Engine is a compiled, reusable rule pipeline.
type Engine struct {
	aliases map[string]string
	ignore  map[string]bool
	include []string
	exclude []string
	rules   []compiled
}

type compiled struct {
	rule config.Rule
	set  *template.Template
	when *template.Template
}

var funcs = template.FuncMap{
	"lower": strings.ToLower,
	"upper": strings.ToUpper,
	"title": title,
	"trim":  strings.TrimSpace,
	"default": func(def, val string) string {
		if strings.TrimSpace(val) == "" {
			return def
		}
		return val
	},
	"replace": func(old, new, s string) string {
		return strings.ReplaceAll(s, old, new)
	},
	"regexReplace": func(expr, repl, s string) (string, error) {
		re, err := regexp.Compile(expr)
		if err != nil {
			return "", fmt.Errorf("regexReplace: %w", err)
		}
		return re.ReplaceAllString(s, repl), nil
	},
	"year":        yearOf,
	"formatTrack": formatTrack,
	"atoi": func(s string) int {
		n, _ := strconv.Atoi(strings.TrimSpace(s))
		return n
	},
	"sanitize": sanitize,
}

// New compiles the configured rules.
func New(cfg *config.Config) (*Engine, error) {
	e := &Engine{
		aliases: make(map[string]string, len(cfg.Fields.Aliases)),
		ignore:  make(map[string]bool, len(cfg.Fields.Ignore)),
		include: append([]string(nil), cfg.Fields.Include...),
		exclude: append([]string(nil), cfg.Fields.Exclude...),
	}
	for key, value := range cfg.Fields.Aliases {
		e.aliases[strings.ToLower(key)] = value
	}
	for _, key := range cfg.Fields.Ignore {
		e.ignore[strings.ToLower(key)] = true
	}

	for i, rule := range cfg.Fields.Rules {
		c := compiled{rule: rule}
		if rule.Set != "" {
			tmpl, err := template.New(fmt.Sprintf("rule-%d-set", i)).Funcs(funcs).Option("missingkey=zero").Parse(rule.Template)
			if err != nil {
				return nil, fmt.Errorf("fields.rules[%d]: template: %w", i, err)
			}
			c.set = tmpl
		}
		if rule.When != "" {
			tmpl, err := template.New(fmt.Sprintf("rule-%d-when", i)).Funcs(funcs).Option("missingkey=zero").Parse(rule.When)
			if err != nil {
				return nil, fmt.Errorf("fields.rules[%d]: when: %w", i, err)
			}
			c.when = tmpl
		}
		e.rules = append(e.rules, c)
	}
	return e, nil
}

// Build applies aliases and rules to raw metadata and returns final fields.
func (e *Engine) Build(res *metadata.Document, meta Meta) (map[string]string, error) {
	return e.BuildWithExisting(res, nil, meta)
}

// BuildWithExisting is Build with fields from an existing .tag merged before
// rules. Existing values never override fields read from the container.
func (e *Engine) BuildWithExisting(res *metadata.Document, existing map[string]string, meta Meta) (map[string]string, error) {
	return e.BuildWithOverrides(res, existing, nil, meta)
}

// BuildWithOverrides is BuildWithExisting with album-level overrides applied
// after raw/existing fields but before rules.
func (e *Engine) BuildWithOverrides(res *metadata.Document, existing, overrides map[string]string, meta Meta) (map[string]string, error) {
	fields := make(map[string]string, len(res.Fields))
	index := make(map[string]string, len(res.Fields))

	set := func(name, value string) {
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		if name == "" || value == "" {
			return
		}
		key := strings.ToLower(name)
		if existing, ok := index[key]; ok {
			fields[existing] = value
			return
		}
		index[key] = name
		fields[name] = value
	}
	get := func(name string) string {
		return fields[index[strings.ToLower(strings.TrimSpace(name))]]
	}
	del := func(name string) {
		key := strings.ToLower(strings.TrimSpace(name))
		if canonical, ok := index[key]; ok {
			delete(fields, canonical)
			delete(index, key)
		}
	}

	rawKeys := make([]string, 0, len(res.Fields))
	for k := range res.Fields {
		rawKeys = append(rawKeys, k)
	}
	sort.Strings(rawKeys)

	for _, key := range rawKeys {
		lower := strings.ToLower(key)
		if e.ignore[lower] {
			continue
		}
		target, ok := e.aliases[lower]
		if !ok {
			if isID3Frame(key) {
				continue
			}
			target = key
		}
		set(target, res.Fields[key])
	}
	for key, value := range existing {
		if _, ok := index[strings.ToLower(strings.TrimSpace(key))]; !ok {
			set(key, value)
		}
	}
	for key, value := range overrides {
		if value == "" {
			del(key)
			clearNumberVars(res.Vars, key)
			continue
		}
		set(key, value)
	}
	mergeNumberVars(res.Vars, get)

	for i, c := range e.rules {
		data := e.templateData(fields, res.Vars, meta)

		if c.when != nil {
			value, err := execute(c.when, data)
			if err != nil {
				return nil, fmt.Errorf("rule %d when: %w", i, err)
			}
			if !isTrue(value) {
				continue
			}
		}

		switch {
		case c.rule.Set != "":
			value, err := execute(c.set, data)
			if err != nil {
				return nil, fmt.Errorf("rule %d set %q: %w", i, c.rule.Set, err)
			}
			set(c.rule.Set, value)

		case c.rule.Copy != nil:
			set(c.rule.Copy.To, get(c.rule.Copy.From))

		case c.rule.Rename != nil:
			if value := get(c.rule.Rename.From); value != "" {
				del(c.rule.Rename.From)
				set(c.rule.Rename.To, value)
			}

		case c.rule.Default != nil:
			if get(c.rule.Default.Field) == "" {
				set(c.rule.Default.Field, get(c.rule.Default.From))
			}

		case len(c.rule.Remove) > 0:
			for _, name := range c.rule.Remove {
				del(name)
			}
		}
	}

	if len(e.include) > 0 {
		allowed := make(map[string]bool, len(e.include))
		for _, name := range e.include {
			allowed[strings.ToLower(name)] = true
		}
		for name := range fields {
			if !allowed[strings.ToLower(name)] {
				delete(fields, name)
			}
		}
	}
	for _, name := range e.exclude {
		del(name)
	}

	return fields, nil
}

func clearNumberVars(vars map[string]string, field string) {
	switch strings.ToLower(strings.TrimSpace(field)) {
	case "track":
		delete(vars, "tracknumber")
		delete(vars, "totaltracks")
	case "disc":
		delete(vars, "discnumber")
		delete(vars, "disctotal")
	}
}

func mergeNumberVars(vars map[string]string, get func(string) string) {
	if vars == nil {
		return
	}
	for _, spec := range []struct {
		field, number, total string
	}{
		{"Track", "tracknumber", "totaltracks"},
		{"Disc", "discnumber", "disctotal"},
	} {
		value := strings.TrimSpace(get(spec.field))
		if value == "" {
			continue
		}
		num, total, hasTotal := strings.Cut(value, "/")
		if vars[spec.number] == "" && strings.TrimSpace(num) != "" {
			vars[spec.number] = strings.TrimSpace(num)
		}
		if hasTotal && vars[spec.total] == "" && strings.TrimSpace(total) != "" {
			vars[spec.total] = strings.TrimSpace(total)
		}
	}
}

// Render executes a template against final fields and metadata.
func (e *Engine) Render(format string, fields, vars map[string]string, meta Meta) (string, error) {
	return e.RenderWithExtra(format, fields, vars, nil, meta)
}

// RenderWithExtra is Render with additional template values, such as artwork.
func (e *Engine) RenderWithExtra(format string, fields, vars, extra map[string]string, meta Meta) (string, error) {
	tmpl, err := template.New("render").Funcs(funcs).Option("missingkey=zero").Parse(format)
	if err != nil {
		return "", err
	}
	data := e.renderData(fields, vars, meta)
	for key, value := range extra {
		if id := ident(key); id != "" {
			data[id] = value
		}
	}
	return executeData(tmpl, data)
}

// RenderRaw renders raw container and .tag fields, without albums or rules.
func (e *Engine) RenderRaw(format string, container, tag map[string]string, vars map[string]string, meta Meta) (string, error) {
	tmpl, err := template.New("render-raw").Funcs(funcs).Option("missingkey=zero").Parse(format)
	if err != nil {
		return "", err
	}
	raw := make(map[string]string, len(container)+len(tag))
	for key, value := range container {
		raw[key] = value
	}
	for key, value := range tag {
		raw[key] = value
	}
	data := e.renderData(raw, vars, meta)
	data["raw"] = raw
	data["container"] = container
	data["tag"] = tag
	return executeData(tmpl, data)
}

func sanitize(value string) string {
	value = strings.TrimSpace(value)
	var b strings.Builder
	for _, r := range value {
		switch {
		case r == '/' || r == '\\' || r == ':' || r == '*' || r == '?':
			b.WriteByte('_')
		case r == '"' || r == '<' || r == '>' || r == '|':
			b.WriteByte('_')
		case r < 0x20 || r == 0x7f:
			b.WriteByte('_')
		default:
			b.WriteRune(r)
		}
	}
	out := strings.TrimRight(b.String(), " .")
	if out == "" || out == "." || out == ".." {
		out = "_"
	}
	base := out
	if i := strings.IndexByte(base, '.'); i >= 0 {
		base = base[:i]
	}
	switch strings.ToUpper(base) {
	case "CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		out = "_" + out
	}
	return out
}

func (e *Engine) renderData(fields, vars map[string]string, meta Meta) map[string]any {
	data := make(map[string]any)
	for key, value := range e.templateData(fields, vars, meta) {
		data[key] = value
	}
	data["fields"] = fields
	data["vars"] = vars
	return data
}

func (e *Engine) templateData(fields, vars map[string]string, meta Meta) map[string]string {
	data := make(map[string]string, len(fields)+len(vars)+5)

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if id := ident(k); id != "" {
			data[id] = fields[k]
		}
	}
	for k, v := range vars {
		if id := ident(k); id != "" {
			data[id] = v
		}
	}

	data["filename"] = meta.Filename
	data["dirname"] = meta.Dir
	data["ext"] = meta.Ext
	data["path"] = meta.Path
	data["now"] = meta.Now.Format(time.RFC3339)
	return data
}

func execute(t *template.Template, data map[string]string) (string, error) {
	return executeData(t, data)
}

func executeData(t *template.Template, data any) (string, error) {
	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

func isTrue(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "false", "0", "no":
		return false
	}
	return true
}

// ident converts a field name into a template identifier:
// "Album artist" -> "albumartist", "©nam" -> "nam".
func ident(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isID3Frame(k string) bool {
	if len(k) != 4 {
		return false
	}
	for _, c := range k {
		if c < 'A' || c > 'Z' {
			return false
		}
	}
	return true
}

func title(s string) string {
	fields := strings.Fields(s)
	for i, f := range fields {
		runes := []rune(strings.ToLower(f))
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
		}
		fields[i] = string(runes)
	}
	return strings.Join(fields, " ")
}

var yearPattern = regexp.MustCompile(`(?:19|20)\d{2}`)

func yearOf(value string) string {
	return yearPattern.FindString(value)
}

func formatTrack(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	numPart, totalPart, withTotal := strings.Cut(value, "/")
	num, err := strconv.Atoi(strings.TrimSpace(numPart))
	if err != nil {
		return value
	}
	if !withTotal {
		return strconv.Itoa(num)
	}
	total, err := strconv.Atoi(strings.TrimSpace(totalPart))
	if err != nil || total <= 0 {
		return value
	}
	width := 1
	if total >= 10 {
		width = len(strconv.Itoa(total))
	}
	return fmt.Sprintf("%0*d/%0*d", width, num, width, total)
}
