package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/DragonSecurity/gomgr/internal/migrate"
)

var (
	migrateFrom      string
	migrateTo        string
	migrateForce     bool
	migrateOnMissing string
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Convert a Python github-user-management config directory into gomgr's layout",
	Long: `Convert a configuration directory in the layout used by the Python
github-user-management tool into the layout gomgr reads.

Every organization directory under --from is converted: org_name becomes
app.org, org_owners becomes org.owners, and each entry in a teams file becomes
its own teams/<slug>.yaml with member/maintainer/repos renamed to
members/maintainers/repositories. Organization-level team defaults are inlined
onto each team, since gomgr has no defaults block.

Nothing is written until every organization converts, and anything the source
says that the output does not is reported. The command exits 2 when the
conversion lost something, so a pipeline cannot mistake a lossy run for a clean
one. It never contacts GitHub and carries no credentials across: set app_id and
private_key on the result deliberately.`,
	Example: `  gomgr migrate --from ./python-orgs --to ./gomgr-orgs
  gomgr migrate --from ./python-orgs --to ./gomgr-orgs --on-missing-permission=drop`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		res, err := migrate.Run(migrate.Options{
			From:                migrateFrom,
			To:                  migrateTo,
			Force:               migrateForce,
			OnMissingPermission: migrate.MissingPermission(migrateOnMissing),
		})
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		p := func(format string, a ...any) { _, _ = fmt.Fprintf(out, format, a...) }

		teams, grants, nested := 0, 0, 0
		for _, o := range res.Orgs {
			p("%-40s org=%-30s teams=%-3d grants=%-3d nested=%d\n", o.Dir, o.Org, o.Teams, o.Grants, o.Nested)
			for _, d := range o.Dropped {
				p("    dropped:  %s\n", d)
			}
			for _, u := range o.Unmapped {
				p("    unmapped: %s\n", u)
			}
			teams += o.Teams
			grants += o.Grants
			nested += o.Nested
		}
		p("\nConverted %d organizations, %d teams (%d nested), %d repository grants into %s\n",
			len(res.Orgs), teams, nested, grants, migrateTo)

		if res.Lossy() {
			p("\nSome of the source did not survive the conversion — see the lines above.\n")
			// Not returned as an error: the conversion happened and the files
			// are on disk. A distinct exit code lets a pipeline branch on it
			// without having to read this output.
			os.Exit(2)
		}
		p("Run `gomgr validate -c <org-dir>` on the result before applying it.\n")
		return nil
	},
}

func init() {
	migrateCmd.Flags().StringVar(&migrateFrom, "from", "", "Directory holding the Python-layout organization directories (required)")
	migrateCmd.Flags().StringVar(&migrateTo, "to", "", "Directory to write the converted configuration into (required)")
	migrateCmd.Flags().BoolVar(&migrateForce, "force", false, "Write into --to even if it is not empty")
	migrateCmd.Flags().StringVar(&migrateOnMissing, "on-missing-permission", string(migrate.MissingPermissionRefuse),
		"What to do with a repository entry that names no permission: refuse or drop")
	_ = migrateCmd.MarkFlagRequired("from")
	_ = migrateCmd.MarkFlagRequired("to")
	rootCmd.AddCommand(migrateCmd)
}
