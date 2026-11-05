package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/daregod/ext-tags/internal/update"
	"github.com/daregod/ext-tags/internal/version"
)

type selfUpdateOptions struct {
	check  bool
	dryRun bool
	force  bool
}

func (a *app) selfUpdateCmd() *cobra.Command {
	opts := &selfUpdateOptions{}
	cmd := &cobra.Command{
		Use:   "self-update",
		Short: "Update a standalone ext-tags binary from GitHub releases",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runSelfUpdate(opts)
		},
	}
	cmd.Flags().BoolVar(&opts.check, "check", false, "check for a newer release without downloading")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "show what would be downloaded without replacing")
	cmd.Flags().BoolVar(&opts.force, "force", false, "update even if managed by a package manager or already current")
	return cmd
}

func (a *app) runSelfUpdate(opts *selfUpdateOptions) error {
	rawExe, err := os.Executable()
	if err != nil {
		return err
	}
	exe := rawExe
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	if runtime.GOOS == "windows" {
		_ = os.Remove(exe + ".old")
	}

	manager, help := update.Managed(rawExe)
	if manager == "" {
		manager, help = update.Managed(exe)
	}
	if manager != "" {
		if !opts.force && opts.check {
			fmt.Fprintf(os.Stderr, "warning: ext-tags is managed by %s; update with `%s`\n", manager, help)
		} else if !opts.force {
			return fmt.Errorf("ext-tags is managed by %s; update with:\n  %s\n(use --force only if you know what you are doing)", manager, help)
		} else {
			fmt.Fprintf(os.Stderr, "warning: %s owns this file; the package manager may overwrite the update; use `%s`\n", manager, help)
		}
	}

	u := update.New(version.Version)
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		u.Token = token
	} else if token := os.Getenv("GH_TOKEN"); token != "" {
		u.Token = token
	}

	rel, err := u.Check()
	if err != nil {
		return fmt.Errorf("check update: %w", err)
	}
	latest := update.NormalizeVersion(rel.TagName)
	current := update.NormalizeVersion(version.Version)
	if opts.check {
		fmt.Printf("current: %s\nlatest:  %s\n", current, latest)
		return nil
	}
	if !opts.force && !update.IsNewer(latest, current) {
		fmt.Printf("ext-tags %s is up to date\n", current)
		return nil
	}
	asset, err := u.AssetFor(rel)
	if err != nil {
		return err
	}
	if opts.dryRun {
		fmt.Printf("would update %s -> %s\nasset: %s\nurl:   %s\n", current, latest, asset.Name, asset.URL)
		return nil
	}
	data, err := u.FetchBinary(rel, asset)
	if err != nil {
		return err
	}
	info, err := os.Stat(exe)
	if err != nil {
		return err
	}
	if err := update.Apply(exe, data, info.Mode()); err != nil {
		return fmt.Errorf("replace executable: %w", err)
	}
	fmt.Printf("updated %s -> %s\n", current, latest)
	return nil
}
