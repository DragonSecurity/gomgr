package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/DragonSecurity/gomgr/internal/config"
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

		client, err := newClient(ctx, cfg)
		if err != nil {
			return err
		}

		result, err := insync.ImportRulesets(ctx, client, cfg, insync.ImportOptions{Only: importOnly})
		if err != nil {
			return err
		}

		if !importWrite {
			printImportPreview(result)
			return nil
		}
		return writeImport(cfgDir, result)
	},
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
		fmt.Printf("Skipped %s already declared in your configuration.\n", plural(result.AlreadyDeclared, "ruleset", "rulesets"))
	}
	if len(result.Skipped) > 0 {
		fmt.Printf("\nLeft %s untouched on GitHub, having no way to express them:\n",
			plural(len(result.Skipped), "ruleset", "rulesets"))
		for _, s := range result.Skipped {
			where := s.Name
			if s.Repo != "" {
				where = s.Repo + "/" + s.Name
			}
			fmt.Printf("  - %s: %s\n", where, s.Reason)
		}
	}
	if len(result.Unmanaged) > 0 {
		fmt.Printf("\nFound rulesets on %s that appear in no team file, leaving\n"+
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
		path, container, err := config.FindRepoDefinitionFile(dir, repo)
		if err != nil {
			return err
		}
		if path == "" {
			// ImportRulesets only reports repos the config knows about, so a
			// miss here means the config changed under us mid-run.
			return fmt.Errorf("repository %q is no longer declared in repos.yaml or any team file", repo)
		}
		if err := config.InsertRepoRulesets(path, container, repo, specs(result.Repos[repo])); err != nil {
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

var importTeamsCmd = &cobra.Command{
	Use:   "teams",
	Short: "Adopt teams that exist on GitHub but are not in your YAML",
	Long: `Scan the organization for teams the configuration does not declare and render
them as team files: name, privacy, description, maintainers, members, and the
repositories each team reaches with the permission it holds.

This is the command for bringing an organization under management that was
never under management before. Teams your configuration already declares are
left alone.

Without --write this only prints what it found. With --write each team is
written to teams/<slug>.yaml; an existing file of that name is never
overwritten.`,
	Example: `  gomgr import teams -c ./config
  gomgr import teams -c ./config --write`,
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
		client, err := newClient(ctx, cfg)
		if err != nil {
			return err
		}

		result, err := insync.ImportTeams(ctx, client, cfg)
		if err != nil {
			return err
		}
		insync.LogTeamImportWarnings(result)

		if !importWrite {
			printTeamPreview(result)
			return nil
		}
		return writeTeamImport(cfgDir, result)
	},
}

func printTeamPreview(result *insync.TeamImportResult) {
	if result.Total() == 0 {
		fmt.Println("Nothing to adopt: every team on GitHub is already declared.")
		reportTeamSkips(result)
		return
	}

	fmt.Printf("%s available to adopt.\n\n", plural(result.Total(), "team", "teams"))
	for _, t := range result.Teams {
		fmt.Printf("  - %-30s %s, %s, %s\n",
			t.Config.ResolvedSlug(),
			plural(len(t.Config.Maintainers), "maintainer", "maintainers"),
			plural(len(t.Config.Members), "member", "members"),
			plural(len(t.Config.Repositories), "repository", "repositories"))
	}

	reportTeamSkips(result)
	fmt.Println("\nRe-run with --write to create the team files.")
}

func reportTeamSkips(result *insync.TeamImportResult) {
	if result.AlreadyDeclared > 0 {
		fmt.Printf("\nSkipped %s already declared in your configuration.\n",
			plural(result.AlreadyDeclared, "team", "teams"))
	}
	if len(result.Skipped) > 0 {
		fmt.Printf("\nLeft %s untouched on GitHub, having no way to express them:\n",
			plural(len(result.Skipped), "team", "teams"))
		for _, s := range result.Skipped {
			fmt.Printf("  - %s: %s\n", s.Slug, s.Reason)
		}
	}
	if len(result.Ungranted) == 0 {
		return
	}

	fmt.Printf("\nNo team reaches %s, so nothing in your configuration\n"+
		"covers them:\n", plural(len(result.Ungranted), "repository", "repositories"))
	for _, repo := range result.Ungranted {
		fmt.Printf("  - %s\n", repo)
	}
	fmt.Println("Grant them to a team before your next sync.")

	// A repository no team reaches is not merely undocumented: under
	// delete_unmanaged_repos it is what the next sync deletes. Named
	// separately from the list above, because the two are not the same set —
	// a repository repos.yaml declares, or one that is already archived, is
	// reached by no team and still left alone.
	if result.DeletionRisk {
		fmt.Printf("\n⚠  delete_unmanaged_repos is set, so THE NEXT SYNC WOULD DELETE %s:\n",
			plural(len(result.WouldDelete), "repository", "repositories"))
		for _, repo := range result.WouldDelete {
			fmt.Printf("  - %s\n", repo)
		}
	}
}

func writeTeamImport(dir string, result *insync.TeamImportResult) error {
	if result.Total() == 0 {
		fmt.Println("Nothing to adopt: every team on GitHub is already declared.")
		reportTeamSkips(result)
		return nil
	}

	for _, t := range result.Teams {
		path, err := config.WriteTeamFile(dir, t.Config)
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			rel = path
		}
		fmt.Printf("adopted team %-30s -> %s\n", t.Config.ResolvedSlug(), rel)
	}

	// Loading the directory proves the new files parse and validate together,
	// before the user finds out on their next sync.
	if _, err := config.Load(dir); err != nil {
		return fmt.Errorf("configuration no longer loads after writing (review with `git status`): %w", err)
	}

	fmt.Printf("\n%s adopted.\n", plural(result.Total(), "team", "teams"))
	reportTeamSkips(result)
	fmt.Println("\nReview with `git status` and `git diff`, then commit and open a pull request.")
	return nil
}

func init() {
	importRulesetsCmd.Flags().BoolVar(&importWrite, "write", false,
		"Splice the adopted rulesets into the configuration files instead of only printing them")
	importRulesetsCmd.Flags().StringSliceVar(&importOnly, "only", nil,
		"Restrict the repository scan to names matching these globs (repeatable)")
	importTeamsCmd.Flags().BoolVar(&importWrite, "write", false,
		"Create the team files instead of only printing what was found")
	importCmd.AddCommand(importRulesetsCmd)
	importCmd.AddCommand(importTeamsCmd)
	rootCmd.AddCommand(importCmd)
}
