package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
)

var installationsRoot string

// userAccountNote is what a personal-account installation is labeled with. It
// is not filtered out: the installation is real and carries the app's full
// permissions against that account's repositories, which is worth seeing even
// though gomgr will never use it.
const userAccountNote = "user account — gomgr manages organizations only"

// installationsCmd reports which organizations the GitHub App can reach, and
// where that disagrees with the configuration directories on disk.
//
// gomgr manages one organization per configuration directory and has no view
// above that, deliberately: reaching the enterprise account would mean holding
// admin:enterprise, since the enterprise API does not separate read from write
// where it would need to. This is the honest answer available without it. The
// app already knows every organization it is installed on, that list needs no
// enterprise scope and no second credential, and it is bounded by exactly what
// gomgr could actually change.
var installationsCmd = &cobra.Command{
	Use:   "installations",
	Short: "List the organizations this GitHub App is installed on",
	Long: `List the organizations this GitHub App is installed on.

With --config-root, compare that list against the configuration directories
found underneath it and report both directions of drift: an organization the
app reaches that nothing configures, and a configuration directory for an
organization the app cannot reach.

Requires GitHub App credentials. A personal access token authenticates a
person, and a person has no app installations.`,
	Example: "  gomgr installations -c ./config\n  gomgr installations -c ./config --config-root ./orgs",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// The config directory is only a source of credentials here, so it is
		// optional: --app-id and --private-key, or the environment, are enough
		// to answer a question that is not about any one organization.
		var app config.AppConfig
		if cfgDir != "" {
			cfg, err := config.Load(cfgDir)
			if err != nil {
				return err
			}
			app = cfg.App
		}
		if appIDFlag != 0 {
			app.AppID = appIDFlag
		}
		if privateKeyFlag != "" {
			app.PrivateKey = privateKeyFlag
		}

		appClient, err := gh.AppClient(app)
		if err != nil {
			if errors.Is(err, gh.ErrNoAppCredentials) {
				return fmt.Errorf("installations needs GitHub App credentials: pass --app-id and --private-key, " +
					"set them in app.yaml, or set GITHUB_APP_ID and GITHUB_APP_PRIVATE_KEY. " +
					"GITHUB_TOKEN cannot answer this — a personal access token authenticates a person, " +
					"and a person has no app installations")
			}
			return err
		}

		installs, err := gh.ListInstallations(ctx, appClient)
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		p := func(format string, a ...any) { _, _ = fmt.Fprintf(out, format, a...) }

		if len(installs) == 0 {
			p("This app is not installed on any organization.\n")
			return nil
		}

		if installationsRoot == "" {
			p("Installed on %d account(s):\n", len(installs))
			for _, in := range installs {
				if !in.IsOrg {
					p("  %-30s installation=%-12d %s\n", in.Org, in.ID, userAccountNote)
					continue
				}
				p("  %-30s installation=%-12d repositories=%s\n", in.Org, in.ID, in.RepositorySelection)
			}
			p("\nPass --config-root <dir> to compare this against the config directories on disk.\n")
			return nil
		}

		dirs, err := config.DiscoverOrgDirs(installationsRoot)
		if err != nil {
			return err
		}
		return reportInstallationDrift(out, installs, dirs)
	},
}

// reportInstallationDrift prints the two-way comparison and is the whole reason
// the command takes a config root.
func reportInstallationDrift(out io.Writer, installs []gh.Installation, dirs []config.OrgDir) error {
	p := func(format string, a ...any) { _, _ = fmt.Fprintf(out, format, a...) }

	configured := map[string][]string{}
	var namelessDirs []string
	for _, d := range dirs {
		if d.Org == "" {
			namelessDirs = append(namelessDirs, d.Dir)
			continue
		}
		configured[d.Org] = append(configured[d.Org], d.Dir)
	}

	installed := make(map[string]gh.Installation, len(installs))
	for _, in := range installs {
		installed[in.Org] = in
	}

	p("Installed on %d account(s), %d config directory/ies found:\n\n", len(installs), len(dirs))

	var unmanaged, unreachable, userAccounts []string
	for _, in := range installs {
		// A user account is reported, never counted as drift: gomgr manages
		// organizations, so no config directory could ever name one and
		// listing it as missing would be an instruction to do the impossible.
		if !in.IsOrg {
			userAccounts = append(userAccounts, in.Org)
			p("  -  %-28s %s\n", in.Org, userAccountNote)
			continue
		}
		dirsFor := configured[in.Org]
		switch {
		case len(dirsFor) == 0:
			unmanaged = append(unmanaged, in.Org)
			p("  ?  %-28s installed, no config directory\n", in.Org)
		default:
			p("  ok %-28s %s\n", in.Org, strings.Join(dirsFor, ", "))
		}
	}
	for _, org := range sortedOrgs(configured) {
		if _, ok := installed[org]; !ok {
			unreachable = append(unreachable, org)
			p("  !  %-28s configured in %s, app NOT installed\n", org, strings.Join(configured[org], ", "))
		}
	}
	for _, dir := range namelessDirs {
		p("  !  %-28s app.yaml names no org\n", dir)
	}

	p("\n")
	if len(unmanaged) > 0 {
		p("%d organization(s) the app reaches that nothing configures: %s\n",
			len(unmanaged), strings.Join(unmanaged, ", "))
		p("  Nothing is applied to these. gomgr only acts on an organization a config directory names.\n")
	}
	if len(unreachable) > 0 {
		p("%d configured organization(s) the app cannot reach: %s\n",
			len(unreachable), strings.Join(unreachable, ", "))
		p("  A sync against these fails at authentication. Install the app, or remove the directory.\n")
	}
	if len(namelessDirs) > 0 {
		p("%d config directory/ies whose app.yaml names no org.\n", len(namelessDirs))
	}
	if len(userAccounts) > 0 {
		p("%d user account(s) the app is installed on: %s\n", len(userAccounts), strings.Join(userAccounts, ", "))
		p("  gomgr cannot act on these, but the installation is real and carries the same\n")
		p("  permissions against that account's repositories. Remove it if nothing needs it.\n")
	}
	if len(unmanaged) == 0 && len(unreachable) == 0 && len(namelessDirs) == 0 {
		p("No drift: every installation has a config directory and every config directory is reachable.\n")
	}
	return nil
}

func sortedOrgs(m map[string][]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func init() {
	installationsCmd.Flags().StringVar(&installationsRoot, "config-root", "",
		"Directory to search for config directories, to compare against the installations")
	rootCmd.AddCommand(installationsCmd)
}
