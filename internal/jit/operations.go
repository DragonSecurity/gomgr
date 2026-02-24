package jit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/google/go-github/v83/github"
)

// GrantTeamAccess grants a user access to a team (temporary or permanent)
func GrantTeamAccess(ctx context.Context, client *gh.Client, org, team, user, role string, duration time.Duration) error {
	if role == "" {
		role = "member"
	}
	role = strings.ToLower(role)
	if role != "member" && role != "maintainer" {
		return fmt.Errorf("invalid role %q: must be 'member' or 'maintainer'", role)
	}

	_, _, err := client.REST.Teams.AddTeamMembershipBySlug(ctx, org, team, user, &github.TeamAddTeamMembershipOptions{
		Role: role,
	})
	if err != nil {
		return fmt.Errorf("add user %s to team %s: %w", user, team, err)
	}

	return nil
}

// RevokeTeamAccess removes a user from a team
func RevokeTeamAccess(ctx context.Context, client *gh.Client, org, team, user string) error {
	_, err := client.REST.Teams.RemoveTeamMembershipBySlug(ctx, org, team, user)
	if err != nil {
		return fmt.Errorf("remove user %s from team %s: %w", user, team, err)
	}
	return nil
}

// GrantRepoAccess grants a user direct access to a repository (temporary or permanent)
func GrantRepoAccess(ctx context.Context, client *gh.Client, org, repo, user, permission string) error {
	if permission == "" {
		permission = "push"
	}
	permission = strings.ToLower(permission)

	// Validate permission
	validPerms := map[string]bool{
		"pull":     true,
		"triage":   true,
		"push":     true,
		"maintain": true,
		"admin":    true,
	}
	if !validPerms[permission] {
		return fmt.Errorf("invalid permission %q: must be one of pull, triage, push, maintain, admin", permission)
	}

	_, _, err := client.REST.Repositories.AddCollaborator(ctx, org, repo, user, &github.RepositoryAddCollaboratorOptions{
		Permission: permission,
	})
	if err != nil {
		return fmt.Errorf("add user %s to repo %s with permission %s: %w", user, repo, permission, err)
	}

	return nil
}

// RevokeRepoAccess removes a user's direct access to a repository
func RevokeRepoAccess(ctx context.Context, client *gh.Client, org, repo, user string) error {
	_, err := client.REST.Repositories.RemoveCollaborator(ctx, org, repo, user)
	if err != nil {
		return fmt.Errorf("remove user %s from repo %s: %w", user, repo, err)
	}
	return nil
}

// CleanupExpiredGrants removes access for expired grants
func CleanupExpiredGrants(ctx context.Context, client *gh.Client, state *State) error {
	expired := state.GetExpired()
	if len(expired) == 0 {
		return nil
	}

	var errs []error
	for _, grant := range expired {
		var err error
		switch grant.Type {
		case "team":
			err = RevokeTeamAccess(ctx, client, grant.Org, grant.Target, grant.User)
		case "repo":
			err = RevokeRepoAccess(ctx, client, grant.Org, grant.Target, grant.User)
		default:
			err = fmt.Errorf("unknown grant type: %s", grant.Type)
		}

		if err != nil {
			errs = append(errs, fmt.Errorf("failed to revoke grant %s: %w", grant.ID, err))
		} else {
			// Remove from state if successful
			if err := state.RemoveGrant(grant.ID); err != nil {
				errs = append(errs, fmt.Errorf("failed to remove grant %s from state: %w", grant.ID, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}
