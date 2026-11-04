package config

import "testing"

func TestValidateAllErrors(t *testing.T) {
	cases := []func(*Config){
		func(c *Config) { c.Version = 2 },
		func(c *Config) { c.Files.Extensions = nil },
		func(c *Config) { c.Files.Extensions = []string{"mp3"} },
		func(c *Config) { c.Files.OnExisting = "bad" },
		func(c *Config) { c.TagFile.Suffix = "tag" },
		func(c *Config) { c.Cover.Field = "" },
		func(c *Config) { c.Cover.Prefer = "bad" },
		func(c *Config) { c.Cover.Names = []string{"bad/name"} },
		func(c *Config) { c.Fields.RulesMode = "bad" },
		func(c *Config) { c.Fields.Rules = []Rule{{}} },
		func(c *Config) { c.Fields.Rules = []Rule{{Set: "X"}} },
		func(c *Config) { c.Fields.Rules = []Rule{{Copy: &CopyRule{}}} },
		func(c *Config) { c.Fields.Rules = []Rule{{Rename: &RenameRule{}}} },
		func(c *Config) { c.Fields.Rules = []Rule{{Default: &DefaultRule{}}} },
	}
	for i, mutate := range cases {
		cfg := Default()
		mutate(cfg)
		if err := cfg.Validate(); err == nil {
			t.Errorf("case %d: expected validation error", i)
		}
	}
}
