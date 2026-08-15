package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	insync "github.com/DragonSecurity/gomgr/internal/sync"
	"github.com/DragonSecurity/gomgr/internal/util"
)

var (
	importWrite bool
	importOnly  []string
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Adopt existing GitHub state into your configuration",
	Long: `Read what already exists on GitHub and render it back as gomgr YAML.

Useful when an organization was managed by hand before gomgr, or when somebody
configured something in the web UI: gomgr can see those resources but has no way
to adopt them into configuration short of transcribing them by hand.`,
}

var importRulesetsCmd = &cobra.Command{
	Use:   "rulesets",
	Short: "Adopt rulesets that exist on GitHub but are not in your YAML",
	Long: `Scan the organization and every repository for rulesets the configuration
does not declare, and render them as YAML.

Rulesets your configuration already declares are left alone: those are gomgr's
to define, and re-importing them would overwrite your YAML with whatever the
live state happens to be.

Without --write this only prints what it found. With --write the entries are
spliced into org.yaml and the teams/*.yaml file that already declares each
repository, leaving the rest of those files — comments included — untouched, so
the result is a reviewable diff in your config repository.`,
	Example: `  gomgr import rulesets -c ./config
  gomgr import rulesets -c ./config --write
  gomgr import rulesets -c ./config --only 'svc-*' --write`,
	RunE: func(_ *cobra.Command, _ []string) error {
		if cfgDir == "" {
			return fmt.Errorf("--config/-c flag is required")
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if debug {
			util.EnableDebug()
		}

		cfg, err := config.Load(cfgDir)
		if err != nil {
			return err
		}

		client, appInfo, err := gh.NewClientFromEnv(ctx, cfg.App)
		if err != nil {
			return err
		}
		if appInfo != "" {
			util.Infof("auth: %s", appInfo)
		}

		result, err := insync.ImportRulesets(ctx, client, cfg, insync.ImportOptions{Only: importOnly})
		if err != nil {
			return err
		}

		if err := validateImport(result); err != nil {
			return err
		}

		if !importWrite {
			printImportPreview(result)
			return nil
		}
		return writeImport(cfgDir, result)
	},
}

// validateImport runs the adopted rulesets through the same checks a config
// file gets, so a ruleset GitHub allows but gomgr's schema does not is caught
// here rather than on the next sync — after it has been written to disk.
func validateImport(result *insync.ImportResult) error {
	if err := config.ValidateRulesets(config.ScopeOrg, "imported org rulesets", specs(result.Org)); err != nil {
		return fmt.Errorf("imported rulesets do not round-trip: %w", err)
	}
	for _, repo := range result.RepoNames() {
		where := fmt.Sprintf("imported rulesets for %s", repo)
		if err := config.ValidateRulesets(config.ScopeRepo, where, specs(result.Repos[repo])); err != nil {
			return fmt.Errorf("imported rulesets do not round-trip: %w", err)
		}
	}
	return nil
}

func specs(imported []insync.ImportedRuleset) []config.RulesetConfig {
	out := make([]config.RulesetConfig, 0, len(imported))
	for _, i := range imported {
		out = append(out, i.Spec)
	}
	return out
}

func printImportPreview(result *insync.ImportResult) {
	if result.Total() == 0 {
		fmt.Printf("Scanned %s: nothing to adopt.\n", plural(result.Scanned, "repository", "repositories"))
		reportSkipped(result)
		return
	}

	fmt.Printf("Scanned %s; %s available to adopt.\n\n", plural(result.Scanned, "repository", "repositories"), plural(result.Total(), "ruleset", "rulesets"))

	if len(result.Org) > 0 {
		fmt.Println("# org.yaml")
		printSpecs(result.Org)
	}
	for _, repo := range result.RepoNames() {
		fmt.Printf("# repository %s\n", repo)
		printSpecs(result.Repos[repo])
	}

	reportSkipped(result)
	fmt.Println("\nRe-run with --write to splice these into your configuration.")
}

func printSpecs(imported []insync.ImportedRuleset) {
	for _, i := range imported {
		summary := i.Spec.Preset
		if summary == "" {
			summary = i.Spec.Target + " ruleset"
		}
		fmt.Printf("  - %-40s %s, %s\n", i.Spec.Name, summary, i.Spec.Enforcement)
	}
	fmt.Println()
}

func reportSkipped(result *insync.ImportResult) {
	if result.AlreadyDeclared > 0 {
		fmt.Printf("%s already declared in your configuration were left alone.\n", plural(result.AlreadyDeclared, "ruleset", "rulesets"))
	}
	if len(result.Unmanaged) > 0 {
		fmt.Printf("\n%s hold rulesets but appear in no team file, so there is\n"+
			"nowhere to write them. Add the repository to a team first:\n", plural(len(result.Unmanaged), "repository", "repositories"))
		for _, repo := range result.Unmanaged {
			fmt.Printf("  - %s\n", repo)
		}
	}
}

func writeImport(dir string, result *insync.ImportResult) error {
	if result.Total() == 0 {
		fmt.Printf("Scanned %s: nothing to adopt.\n", plural(result.Scanned, "repository", "repositories"))
		reportSkipped(result)
		return nil
	}

	touched := map[string]int{}

	if len(result.Org) > 0 {
		path := filepath.Join(dir, "org.yaml")
		if err := config.InsertOrgRulesets(path, specs(result.Org)); err != nil {
			return err
		}
		touched[path] += len(result.Org)
	}

	for _, repo := range result.RepoNames() {
		path, err := config.FindTeamFileForRepo(dir, repo)
		if err != nil {
			return err
		}
		if path == "" {
			// ImportRulesets only reports repos the config knows about, so a
			// miss here means the config changed under us mid-run.
			return fmt.Errorf("repository %q is no longer declared in any team file", repo)
		}
		if err := config.InsertRepoRulesets(path, repo, specs(result.Repos[repo])); err != nil {
			return err
		}
		touched[path] += len(result.Repos[repo])
	}

	paths := make([]string, 0, len(touched))
	for path := range touched {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			rel = path
		}
		fmt.Printf("adopted %s -> %s\n", plural(touched[path], "ruleset", "rulesets"), rel)
	}

	// Loading the whole directory proves the edits produced a configuration
	// gomgr can still read, before the user finds out on their next sync.
	if _, err := config.Load(dir); err != nil {
		return fmt.Errorf("configuration no longer loads after writing (review with `git diff`): %w", err)
	}

	fmt.Printf("\n%s adopted across %s.\n", plural(result.Total(), "ruleset", "rulesets"), plural(len(paths), "file", "files"))
	reportSkipped(result)
	fmt.Println("\nReview with `git diff`, then commit and open a pull request.")
	return nil
}

// plural counts a noun. Both forms are spelled out because the nouns here
// include "repository", which no suffix rule gets right.
func plural(n int, singular, many string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}
	return fmt.Sprintf("%d %s", n, many)
}

func init() {
	importRulesetsCmd.Flags().BoolVar(&importWrite, "write", false,
		"Splice the adopted rulesets into the configuration files instead of only printing them")
	importRulesetsCmd.Flags().StringSliceVar(&importOnly, "only", nil,
		"Restrict the repository scan to names matching these globs (repeatable)")
	importCmd.AddCommand(importRulesetsCmd)
	rootCmd.AddCommand(importCmd)
}
