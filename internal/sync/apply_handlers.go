package sync

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
	"github.com/google/go-github/v83/github"
)

// extractDetails performs a safe type assertion on ch.Details to map[string]any.
func extractDetails(ch util.Change) (map[string]any, error) {
	d, ok := ch.Details.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid details for %s:%s: expected map[string]any, got %T", ch.Scope, ch.Action, ch.Details)
	}
	return d, nil
}

func detailString(d map[string]any, key string) string {
	if v, ok := d[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprint(v)
	}
	return ""
}

func detailBool(d map[string]any, key string) bool {
	if v, ok := d[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
		return fmt.Sprint(v) == "true"
	}
	return false
}

func applyTeamCreate(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	name := detailString(d, "name")
	var privacyPtr, descPtr *string
	if pv := detailString(d, "privacy"); pv != "" {
		privacyPtr = github.Ptr(pv)
	}
	if dv := detailString(d, "description"); dv != "" {
		descPtr = github.Ptr(dv)
	}
	newTeam := github.NewTeam{Name: name, Privacy: privacyPtr, Description: descPtr}
	_, _, err = c.REST.Teams.CreateTeam(ctx, org, newTeam)
	if err != nil {
		return fmt.Errorf("create team %q: %w", name, err)
	}
	return nil
}

func applyTeamUpdate(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	slug := detailString(d, "slug")
	name := detailString(d, "name")
	newTeam := github.NewTeam{Name: name}
	if dv := detailString(d, "description"); dv != "" {
		newTeam.Description = github.Ptr(dv)
	}
	if pv := detailString(d, "privacy"); pv != "" {
		newTeam.Privacy = github.Ptr(pv)
	}
	_, _, err = c.REST.Teams.EditTeamBySlug(ctx, org, slug, newTeam, false)
	if err != nil {
		return fmt.Errorf("update team %q: %w", slug, err)
	}
	return nil
}

func applyTeamDelete(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	slug := detailString(d, "slug")
	_, err = c.REST.Teams.DeleteTeamBySlug(ctx, org, slug)
	if err != nil {
		return fmt.Errorf("delete team %q in org %q: %w", slug, org, err)
	}
	return nil
}

func applyTeamMemberEnsure(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, ok := ch.Details.(teamMemberChange)
	if !ok {
		return fmt.Errorf("invalid details for team-member:ensure: expected teamMemberChange, got %T", ch.Details)
	}
	_, _, err := c.REST.Teams.AddTeamMembershipBySlug(ctx, d.Org, d.Slug, d.User, &github.TeamAddTeamMembershipOptions{Role: d.Role})
	if err != nil {
		return fmt.Errorf("add %q as %q to %q in org %q: %w", d.User, d.Role, d.Slug, d.Org, err)
	}
	return nil
}

func applyRepoEnsure(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	name := detailString(d, "name")
	private := true
	if v, ok := d["private"]; ok {
		if b, isBool := v.(bool); isBool {
			private = b
		} else {
			private = fmt.Sprint(v) != "false"
		}
	}
	isTemplate := detailBool(d, "template")

	// Check if this repo should be created from a template
	if templateRef := detailString(d, "from"); templateRef != "" {
		templateOrg, templateRepo := parseTemplateRef(templateRef, org)

		// Create repository from template
		_, _, err := c.REST.Repositories.CreateFromTemplate(ctx, templateOrg, templateRepo, &github.TemplateRepoRequest{
			Name:    github.Ptr(name),
			Owner:   github.Ptr(org),
			Private: github.Ptr(private),
		})
		if err != nil {
			var ghErr *github.ErrorResponse
			if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == 422 {
				// already exists race
			} else {
				return fmt.Errorf("create repo %s/%s from template %s/%s: %w", org, name, templateOrg, templateRepo, err)
			}
		}
	} else {
		// Create regular repository
		_, _, err := c.REST.Repositories.Create(ctx, org, &github.Repository{
			Name:                github.Ptr(name),
			Private:             github.Ptr(private),
			IsTemplate:          github.Ptr(isTemplate),
			AllowAutoMerge:      github.Ptr(true),
			AllowMergeCommit:    github.Ptr(false),
			DeleteBranchOnMerge: github.Ptr(true),
			HasIssues:           github.Ptr(true),
		})
		if err != nil {
			var ghErr *github.ErrorResponse
			if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == 422 {
				// already exists race
			} else {
				return fmt.Errorf("create repo %s/%s: %w", org, name, err)
			}
		}
	}
	return nil
}

func applyTeamRepoGrant(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	slug := detailString(d, "slug")
	repo := detailString(d, "repo")
	perm := normalizePermission(detailString(d, "permission"))
	_, err = c.REST.Teams.AddTeamRepoBySlug(ctx, org, slug, org, repo, &github.TeamAddTeamRepoOptions{Permission: perm})
	if err != nil {
		return fmt.Errorf("grant %q on %s/%s to %q: %w", perm, org, repo, slug, err)
	}
	return nil
}

func applyRepoFileEnsure(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	repo := detailString(d, "repo")
	path := detailString(d, "path")
	content := []byte(detailString(d, "content"))
	message := detailString(d, "message")
	branch := detailString(d, "branch")
	file, _, resp, err := c.REST.Repositories.GetContents(ctx, org, repo, path, &github.RepositoryContentGetOptions{Ref: branch})
	if err != nil && (resp == nil || resp.StatusCode != http.StatusNotFound) {
		return fmt.Errorf("check %s/%s:%s: %w", org, repo, path, err)
	}
	if file == nil {
		_, _, err := c.REST.Repositories.CreateFile(ctx, org, repo, path, &github.RepositoryContentFileOptions{
			Message: github.Ptr(message),
			Content: content,
			Branch:  github.Ptr(branch),
		})
		if err != nil {
			// Handle race condition: If repository was created from template,
			// files may exist even though GetContents returned nil.
			var ghErr *github.ErrorResponse
			if errors.As(err, &ghErr) && ghErr.Response != nil {
				isRaceCondition := (ghErr.Response.StatusCode == 422 && containsErrorMessage(ghErr, errTermSHA, errTermSHANotSupplied)) ||
					(ghErr.Response.StatusCode == 409 && containsErrorMessage(ghErr, errTermRefExists))

				if !isRaceCondition {
					return fmt.Errorf("create file %s in %s/%s: %w", path, org, repo, err)
				}
				// File already exists (likely from template), which is what we want - skip error
			} else {
				return fmt.Errorf("create file %s in %s/%s: %w", path, org, repo, err)
			}
		}
	}
	return nil
}

func applyRepoTopicsEnsure(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	repo := detailString(d, "repo")

	// Handle topics - may come as []string or []any from planning
	var topicsRaw []string
	if v, ok := d["topics"]; ok {
		switch topics := v.(type) {
		case []string:
			topicsRaw = topics
		case []any:
			for _, t := range topics {
				if tStr, ok := t.(string); ok {
					topicsRaw = append(topicsRaw, tStr)
				}
			}
		default:
			return fmt.Errorf("invalid type for topics for %s/%s: %T", org, repo, v)
		}
	}

	_, _, err = c.REST.Repositories.ReplaceAllTopics(ctx, org, repo, topicsRaw)
	if err != nil {
		return fmt.Errorf("set topics on %s/%s: %w", org, repo, err)
	}
	return nil
}

func applyRepoTemplateEnsure(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	repo := detailString(d, "repo")

	_, _, err = c.REST.Repositories.Edit(ctx, org, repo, &github.Repository{
		IsTemplate: github.Ptr(true),
	})
	if err != nil {
		return fmt.Errorf("mark repo %s/%s as template: %w", org, repo, err)
	}
	return nil
}

func applyRepoPinEnsure(_ context.Context, _ *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	repo := detailString(d, "repo")

	util.Warnf("Skipping pin for %s/%s: GitHub API does not support pinning to organization profiles", org, repo)
	return nil
}

func applyRepoDelete(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	repo := detailString(d, "repo")
	_, err = c.REST.Repositories.Delete(ctx, org, repo)
	if err != nil {
		return fmt.Errorf("delete repo %q in org %q: %w", repo, org, err)
	}
	return nil
}

func applyOrgMemberRemove(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	user := detailString(d, "user")
	_, err = c.REST.Organizations.RemoveOrgMembership(ctx, user, org)
	if err != nil {
		return fmt.Errorf("remove member %q from org %q: %w", user, org, err)
	}
	return nil
}

func normalizePermission(p string) string {
	switch strings.ToLower(p) {
	case "read", "pull":
		return "pull"
	case "triage":
		return "triage"
	case "write", "push":
		return "push"
	case "maintain":
		return "maintain"
	case "admin":
		return "admin"
	default:
		return p
	}
}
