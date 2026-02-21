package sync

import (
	"context"
	"fmt"
	"strings"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
	"github.com/google/go-github/v83/github"
)

// customRoleChange represents a custom role modification
type customRoleChange struct {
	Org         string
	ID          int64
	Name        string
	Description string
	BaseRole    string
	Permissions []string
}

// planCustomRoles determines what custom repository role changes are needed
func planCustomRoles(ctx context.Context, c *gh.Client, cfg *config.Root, st *State) ([]util.Change, error) {
	var out []util.Change
	org := st.Org

	if len(cfg.Org.CustomRoles) == 0 {
		// No custom roles configured
		return out, nil
	}

	// Fetch existing custom roles from GitHub
	existingRolesResp, _, err := c.REST.Organizations.ListCustomRepoRoles(ctx, org)
	if err != nil {
		// If the org doesn't have custom roles enabled (not Enterprise Cloud),
		// return an error with helpful context
		return out, fmt.Errorf("list custom repo roles: %w (note: custom roles require GitHub Enterprise Cloud)", err)
	}

	// Build map of existing roles by name (lowercase for case-insensitive comparison)
	existingByName := make(map[string]*github.CustomRepoRoles)
	for _, role := range existingRolesResp.CustomRepoRoles {
		if role.Name != nil {
			existingByName[strings.ToLower(*role.Name)] = role
		}
	}

	// Track state
	st.CurrentCustomRoles = len(existingRolesResp.CustomRepoRoles)
	st.DesiredCustomRoles = len(cfg.Org.CustomRoles)

	// Plan changes for each desired role
	for _, desiredRole := range cfg.Org.CustomRoles {
		roleName := desiredRole.Name
		roleNameLower := strings.ToLower(roleName)

		existingRole, exists := existingByName[roleNameLower]

		if !exists {
			// Create new role
			out = append(out, util.Change{
				Scope:  "custom-role",
				Target: roleName,
				Action: "create",
				Details: customRoleChange{
					Org:         org,
					Name:        roleName,
					Description: desiredRole.Description,
					BaseRole:    desiredRole.BaseRole,
					Permissions: desiredRole.Permissions,
				},
			})
		} else {
			// Check if update is needed
			needsUpdate := false

			// Check description changes
			existingDesc := ""
			if existingRole.Description != nil {
				existingDesc = *existingRole.Description
			}
			if existingDesc != desiredRole.Description {
				needsUpdate = true
			}

			// Check base role changes
			if existingRole.BaseRole != nil && *existingRole.BaseRole != desiredRole.BaseRole {
				needsUpdate = true
			}

			// Check permission changes
			if !permissionsEqual(existingRole.Permissions, desiredRole.Permissions) {
				needsUpdate = true
			}

			if needsUpdate {
				out = append(out, util.Change{
					Scope:  "custom-role",
					Target: roleName,
					Action: "update",
					Details: customRoleChange{
						Org:         org,
						ID:          existingRole.GetID(),
						Name:        roleName,
						Description: desiredRole.Description,
						BaseRole:    desiredRole.BaseRole,
						Permissions: desiredRole.Permissions,
					},
				})
			}
		}
	}

	return out, nil
}

// planCustomRoleCleanups determines which custom roles should be deleted
func planCustomRoleCleanups(ctx context.Context, c *gh.Client, cfg *config.Root, st *State) ([]util.Change, []string, error) {
	var out []util.Change
	var warnings []string
	org := st.Org

	if !cfg.App.DeleteUnmanagedCustomRoles && !cfg.App.DryWarnings.WarnUnmanagedCustomRoles {
		return out, warnings, nil
	}

	// Fetch existing custom roles
	existingRolesResp, _, err := c.REST.Organizations.ListCustomRepoRoles(ctx, org)
	if err != nil {
		// If custom roles aren't available, skip cleanup
		return out, warnings, nil
	}

	// Build set of desired role names (case-insensitive)
	desiredNames := make(map[string]bool)
	for _, role := range cfg.Org.CustomRoles {
		desiredNames[strings.ToLower(role.Name)] = true
	}

	// Find unmanaged roles
	var unmanagedRoles []string
	for _, role := range existingRolesResp.CustomRepoRoles {
		if role.Name == nil {
			continue
		}
		roleName := *role.Name
		if !desiredNames[strings.ToLower(roleName)] {
			unmanagedRoles = append(unmanagedRoles, roleName)
			if cfg.App.DeleteUnmanagedCustomRoles {
				out = append(out, util.Change{
					Scope:  "custom-role",
					Target: roleName,
					Action: "delete",
					Details: customRoleChange{
						Org:  org,
						ID:   role.GetID(),
						Name: roleName,
					},
				})
			}
		}
	}

	if cfg.App.DryWarnings.WarnUnmanagedCustomRoles && len(unmanagedRoles) > 0 {
		warnings = append(warnings, fmt.Sprintf("Found %d unmanaged custom repository roles: %v", len(unmanagedRoles), unmanagedRoles))
	}

	return out, warnings, nil
}

// applyCustomRoleChanges handles creating, updating, and deleting custom roles
func applyCustomRoleChanges(ctx context.Context, c *gh.Client, changes []util.Change) error {
	for _, ch := range changes {
		if !strings.HasPrefix(ch.Scope, "custom-role") {
			continue
		}

		d, ok := ch.Details.(customRoleChange)
		if !ok {
			return fmt.Errorf("invalid details for custom-role change")
		}

		switch ch.Scope + ":" + ch.Action {
		case "custom-role:create":
			opts := &github.CreateOrUpdateCustomRepoRoleOptions{
				Name:        github.Ptr(d.Name),
				BaseRole:    github.Ptr(d.BaseRole),
				Permissions: d.Permissions,
			}
			if d.Description != "" {
				opts.Description = github.Ptr(d.Description)
			}

			_, _, err := c.REST.Organizations.CreateCustomRepoRole(ctx, d.Org, opts)
			if err != nil {
				return fmt.Errorf("create custom role %q: %w", d.Name, err)
			}

		case "custom-role:update":
			opts := &github.CreateOrUpdateCustomRepoRoleOptions{
				Name:        github.Ptr(d.Name),
				BaseRole:    github.Ptr(d.BaseRole),
				Permissions: d.Permissions,
			}
			if d.Description != "" {
				opts.Description = github.Ptr(d.Description)
			}

			_, _, err := c.REST.Organizations.UpdateCustomRepoRole(ctx, d.Org, d.ID, opts)
			if err != nil {
				return fmt.Errorf("update custom role %q (ID %d): %w", d.Name, d.ID, err)
			}

		case "custom-role:delete":
			_, err := c.REST.Organizations.DeleteCustomRepoRole(ctx, d.Org, d.ID)
			if err != nil {
				return fmt.Errorf("delete custom role %q (ID %d): %w", d.Name, d.ID, err)
			}
		}
	}

	return nil
}

// permissionsEqual checks if two permission lists are equivalent
func permissionsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aSet := make(map[string]bool)
	for _, p := range a {
		aSet[p] = true
	}
	for _, p := range b {
		if !aSet[p] {
			return false
		}
	}
	return true
}
