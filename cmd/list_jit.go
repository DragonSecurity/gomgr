package cmd

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/DragonSecurity/gomgr/internal/jit"
	"github.com/DragonSecurity/gomgr/internal/util"
	"github.com/spf13/cobra"
)

var (
	listStateDir string
)

var listJitCmd = &cobra.Command{
	Use:   "list-jit",
	Short: "List active JIT access grants",
	Long: `List all active Just-In-Time (JIT) access grants.

Shows currently active grants with their expiration times.

Example:
  gomgr list-jit`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if debug {
			util.EnableDebug()
		}

		// Initialize JIT state
		state, err := jit.NewState(listStateDir)
		if err != nil {
			return fmt.Errorf("init JIT state: %w", err)
		}

		// Get all grants
		grants := state.GetAll()
		if len(grants) == 0 {
			log.Println("No active JIT grants")
			return nil
		}

		// Sort by expiration time
		sort.Slice(grants, func(i, j int) bool {
			return grants[i].ExpiresAt.Before(grants[j].ExpiresAt)
		})

		now := time.Now()
		log.Printf("Active JIT Grants (%d total):\n", len(grants))
		log.Println("=====================================")

		for _, g := range grants {
			status := "active"
			timeLeft := g.ExpiresAt.Sub(now)
			if timeLeft < 0 {
				status = "EXPIRED"
				timeLeft = -timeLeft
			}

			log.Printf("\nID: %s", g.ID)
			log.Printf("Type: %s", g.Type)
			log.Printf("Org: %s", g.Org)
			log.Printf("Target: %s", g.Target)
			log.Printf("User: %s", g.User)
			log.Printf("Permission: %s", g.Permission)
			log.Printf("Granted: %s", g.GrantedAt.Format("2006-01-02 15:04:05"))
			log.Printf("Expires: %s", g.ExpiresAt.Format("2006-01-02 15:04:05"))
			if status == "EXPIRED" {
				log.Printf("Status: %s (expired %s ago)", status, timeLeft.Round(time.Minute))
			} else {
				log.Printf("Status: %s (expires in %s)", status, timeLeft.Round(time.Minute))
			}
		}

		log.Println("\n=====================================")
		log.Printf("Use 'gomgr cleanup-jit --org <org>' to clean up expired grants")

		return nil
	},
}

func init() {
	listJitCmd.Flags().StringVar(&listStateDir, "state-dir", "", "Directory to store JIT state (default: ~/.gomgr)")
	rootCmd.AddCommand(listJitCmd)
}
