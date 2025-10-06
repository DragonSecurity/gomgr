package sync

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/util"
	"github.com/google/go-github/v75/github"
)

type teamMemberChange struct {
	Org  string
	Slug string
	User string
	Role string // "member" or "maintainer"
}

// ---- planning ----

func planTeams(ctx context.Context, c *gh.Client, cfg *config.Root, st *State) ([]util.Change, map[string]config.TeamConfig, error) {
	var out []util.Change
	desired := map[string]config.TeamConfig{}

	// build desired map
	for _, t := range cfg.Team {
		slug := t.Slug
		if slug == "" && t.Name != "" {
			slug = strings.ToLower(strings.ReplaceAll(t.Name, " ", "-"))
		}
		if slug == "" {
			continue
		}
		t.Slug = slug
		desired[slug] = t
	}

	// list actual teams
	var actual []*github.Team
	opt := &github.ListOptions{PerPage: 100}
	for {
		ts, resp, err := c.REST.Teams.ListTeams(ctx, st.Org, opt)
		if err != nil {
			return nil, nil, err
		}
		actual = append(actual, ts...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	actualBySlug := map[string]*github.Team{}
	for _, t := range actual {
		actualBySlug[t.GetSlug()] = t
	}

	for slug, want := range desired {
		if _, ok := actualBySlug[slug]; !ok {
			out = append(out, util.Change{
				Scope:  "team",
				Target: slug,
				Action: "create",
				Details: map[string]any{
					"org":         st.Org,
					"name":        want.Name,
					"privacy":     want.Privacy,
					"description": want.Description,
				},
			})
			continue
		}
		// TODO: compare & update description/privacy/parents as needed
	}
	return out, desired, nil
}

func planTeamMembership(ctx context.Context, c *gh.Client, cfg *config.Root, st *State, desiredBySlug map[string]config.TeamConfig) ([]util.Change, error) {
	var out []util.Change
	org := st.Org

	for slug, want := range desiredBySlug {
		// actual role map
		got := map[string]string{}
		// maintainers
		mopts := &github.TeamListTeamMembersOptions{Role: "maintainer", ListOptions: github.ListOptions{PerPage: 100}}
		for {
			users, resp, err := c.REST.Teams.ListTeamMembersBySlug(ctx, org, slug, mopts)
			if err != nil {
				var ghErr *github.ErrorResponse
				if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
					break
				}
				return nil, err
			}
			for _, u := range users {
				got[strings.ToLower(u.GetLogin())] = "maintainer"
			}
			if resp.NextPage == 0 {
				break
			}
			mopts.Page = resp.NextPage
		}
		// members
		opts := &github.TeamListTeamMembersOptions{Role: "member", ListOptions: github.ListOptions{PerPage: 100}}
		for {
			users, resp, err := c.REST.Teams.ListTeamMembersBySlug(ctx, org, slug, opts)
			if err != nil {
				var ghErr *github.ErrorResponse
				if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
					break
				}
				return nil, err
			}
			for _, u := range users {
				if _, ok := got[strings.ToLower(u.GetLogin())]; !ok {
					got[strings.ToLower(u.GetLogin())] = "member"
				}
			}
			if resp.NextPage == 0 {
				break
			}
			opts.Page = resp.NextPage
		}

		// desired role map
		wantRole := map[string]string{}
		for _, u := range want.Maintainers {
			wantRole[strings.ToLower(u)] = "maintainer"
		}
		for _, u := range want.Members {
			if _, ok := wantRole[strings.ToLower(u)]; !ok {
				wantRole[strings.ToLower(u)] = "member"
			}
		}

		for user, role := range wantRole {
			if got[user] == role {
				continue
			}
			out = append(out, util.Change{
				Scope:   "team-member",
				Target:  slug,
				Action:  "ensure",
				Details: teamMemberChange{Org: org, Slug: slug, User: user, Role: role},
			})
		}
		// (optional) removals left for later
	}
	return out, nil
}

func planRepoPerms(ctx context.Context, c *gh.Client, cfg *config.Root, st *State) ([]util.Change, error) {
	var out []util.Change
	org := st.Org

	existing := map[string]bool{}
	opt := &github.RepositoryListByOrgOptions{ListOptions: github.ListOptions{PerPage: 100}, Type: "all"}
	for {
		repos, resp, err := c.REST.Repositories.ListByOrg(ctx, org, opt)
		if err != nil {
			return nil, err
		}
		for _, r := range repos {
			existing[strings.ToLower(r.GetName())] = true
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	for _, t := range cfg.Team {
		slug := t.Slug
		if slug == "" {
			slug = strings.ToLower(strings.ReplaceAll(t.Name, " ", "-"))
		}
		for repo, perm := range t.Repositories {
			r := strings.ToLower(repo)
			if !existing[r] && cfg.App.CreateRepo {
				out = append(out, util.Change{Scope: "repo", Target: r, Action: "ensure", Details: map[string]any{"org": org, "name": repo, "private": true}})
				existing[r] = true
			}
			out = append(out, util.Change{Scope: "team-repo", Target: slug + "/" + r, Action: "grant", Details: map[string]any{"org": org, "slug": slug, "repo": repo, "permission": perm}})
			if cfg.App.AddDefaultReadme {
				readmeContent := fmt.Sprintf(
					"# %s\n\n"+
						"Quick setup — if you’ve done this kind of thing before  \n"+
						"or clone directly:  \n\n"+
						"```bash\n"+
						"git clone git@github.com:%s/%s.git\n"+
						"```\n\n"+
						"Get started by creating a new file or uploading an existing one.  \n"+
						"We recommend every repository include a README, LICENSE, and .gitignore.\n\n"+
						"…or create a new repository on the command line\n\n"+
						"```bash\n"+
						"echo \"# %s\" >> README.md\n"+
						"git init\n"+
						"git add README.md\n"+
						"git commit -m \"first commit\"\n"+
						"git branch -M main\n"+
						"git remote add origin git@github.com:%s/%s.git\n"+
						"git push -u origin main\n"+
						"```\n\n"+
						"…or push an existing repository from the command line\n\n"+
						"```bash\n"+
						"git remote add origin git@github.com:%s/%s.git\n"+
						"git branch -M main\n"+
						"git push -u origin main\n"+
						"```\n",
					repo, org, repo, repo, org, repo, org, repo,
				)
				out = append(out, util.Change{
					Scope:  "repo-file",
					Target: r + ":README.md",
					Action: "ensure",
					Details: map[string]any{
						"org":     org,
						"repo":    repo,
						"path":    "README.md",
						"content": readmeContent,
						"message": "chore: add default README",
						"branch":  "main",
					},
				})
			}
			if cfg.App.AddRenovateConfig && cfg.App.RenovateConfig != "" {
				out = append(out, util.Change{
					Scope:  "repo-file",
					Target: r + ":.github/renovate.json",
					Action: "ensure",
					Details: map[string]any{
						"org":     org,
						"repo":    repo,
						"path":    ".github/renovate.json",
						"content": cfg.App.RenovateConfig,
						"message": "chore: add Renovate config",
						"branch":  "main",
					},
				})
			}
		}
	}
	return out, nil
}

func planCleanups(ctx context.Context, c *gh.Client, cfg *config.Root, st *State, desired map[string]config.TeamConfig) ([]util.Change, error) {
	var out []util.Change
	org := st.Org
	if cfg.App.DeleteUnconfiguredTeams {
		var actual []*github.Team
		opt := &github.ListOptions{PerPage: 100}
		for {
			ts, resp, err := c.REST.Teams.ListTeams(ctx, org, opt)
			if err != nil {
				return nil, err
			}
			actual = append(actual, ts...)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}
		for _, at := range actual {
			if _, ok := desired[at.GetSlug()]; !ok {
				out = append(out, util.Change{Scope: "team", Target: at.GetSlug(), Action: "delete", Details: map[string]any{"org": org, "slug": at.GetSlug()}})
			}
		}
	}

	if cfg.App.RemoveMembersWithoutTeam {
		// list all org members
		memOpt := &github.ListMembersOptions{
			Role: "member",
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		}
		var members []*github.User
		for {
			us, resp, err := c.REST.Organizations.ListMembers(ctx, org, memOpt)
			if err != nil {
				return nil, err
			}
			members = append(members, us...)
			if resp.NextPage == 0 {
				break
			}
			memOpt.Page = resp.NextPage
		}
		// compute members who are in any team
		inAnyTeam := map[string]bool{}
		teamOpt := &github.ListOptions{PerPage: 100}
		for {
			ts, resp, err := c.REST.Teams.ListTeams(ctx, org, teamOpt)
			if err != nil {
				return nil, err
			}
			for _, t := range ts {
				page := &github.TeamListTeamMembersOptions{Role: "all", ListOptions: github.ListOptions{PerPage: 100}}
				for {
					us, r2, err := c.REST.Teams.ListTeamMembersBySlug(ctx, org, t.GetSlug(), page)
					if err != nil {
						return nil, err
					}
					for _, u := range us {
						inAnyTeam[strings.ToLower(u.GetLogin())] = true
					}
					if r2.NextPage == 0 {
						break
					}
					page.Page = r2.NextPage
				}
			}
			if resp.NextPage == 0 {
				break
			}
			teamOpt.Page = resp.NextPage
		}
		for _, u := range members {
			login := strings.ToLower(u.GetLogin())
			if !inAnyTeam[login] {
				out = append(out, util.Change{Scope: "org-member", Target: login, Action: "remove", Details: map[string]any{"org": org, "user": login}})
			}
		}
	}
	return out, nil
}

// ---- apply ----

func applyChanges(ctx context.Context, c *gh.Client, changes []util.Change) error {
	precedence := map[string]int{
		"team:create":        10,
		"repo:ensure":        10,
		"team-repo:grant":    20,
		"team-member:ensure": 30,
		"repo-file:ensure":   40,
		"team:delete":        90,
	}

	sort.Slice(changes, func(i, j int) bool {
		ai := changes[i].Scope + ":" + changes[i].Action
		aj := changes[j].Scope + ":" + changes[j].Action
		return precedence[ai] < precedence[aj]
	})

	for _, ch := range changes {
		switch ch.Scope + ":" + ch.Action {
		case "team:create":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			name := fmt.Sprint(d["name"])
			var privacyPtr, descPtr *string
			if v, ok := d["privacy"]; ok && fmt.Sprint(v) != "" {
				pv := fmt.Sprint(v)
				privacyPtr = github.Ptr(pv)
			}
			if v, ok := d["description"]; ok && fmt.Sprint(v) != "" {
				dv := fmt.Sprint(v)
				descPtr = github.Ptr(dv)
			}
			newTeam := github.NewTeam{Name: name, Privacy: privacyPtr, Description: descPtr}
			_, _, err := c.REST.Teams.CreateTeam(ctx, org, newTeam)
			if err != nil {
				return fmt.Errorf("create team %q: %w", name, err)
			}

		case "team:delete":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			slug := fmt.Sprint(d["slug"])
			_, err := c.REST.Teams.DeleteTeamBySlug(ctx, org, slug)
			if err != nil {
				return fmt.Errorf("delete team %s: %w", slug, err)
			}

		case "team-member:ensure":
			d, ok := ch.Details.(teamMemberChange)
			if !ok {
				return fmt.Errorf("invalid details for team-member:ensure")
			}
			_, _, err := c.REST.Teams.AddTeamMembershipBySlug(ctx, d.Org, d.Slug, d.User, &github.TeamAddTeamMembershipOptions{Role: d.Role})
			if err != nil {
				return fmt.Errorf("add %s as %s to %s: %w", d.User, d.Role, d.Slug, err)
			}

		case "repo:ensure":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			name := fmt.Sprint(d["name"])
			private := true
			if v, ok := d["private"]; ok {
				private = fmt.Sprint(v) != "false"
			}
			_, _, err := c.REST.Repositories.Create(ctx, org, &github.Repository{
				Name:                github.Ptr(name),
				Private:             github.Ptr(private),
				AllowAutoMerge:      github.Ptr(true),
				AllowMergeCommit:    github.Ptr(false),
				DeleteBranchOnMerge: github.Ptr(true),
			})
			if err != nil {
				var ghErr *github.ErrorResponse
				if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == 422 {
					// already exists race
				} else {
					return fmt.Errorf("create repo %s/%s: %w", org, name, err)
				}
			}

		case "team-repo:grant":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			slug := fmt.Sprint(d["slug"])
			repo := fmt.Sprint(d["repo"])
			perm := normalizePermission(fmt.Sprint(d["permission"]))
			_, err := c.REST.Teams.AddTeamRepoBySlug(ctx, org, slug, org, repo, &github.TeamAddTeamRepoOptions{Permission: perm})
			if err != nil {
				return fmt.Errorf("grant %s on %s/%s to %s: %w", perm, org, repo, slug, err)
			}

		case "repo-file:ensure":
			d, _ := ch.Details.(map[string]any)
			org := fmt.Sprint(d["org"])
			repo := fmt.Sprint(d["repo"])
			path := fmt.Sprint(d["path"])
			content := []byte(fmt.Sprint(d["content"]))
			message := fmt.Sprint(d["message"])
			branch := fmt.Sprint(d["branch"])
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
					return fmt.Errorf("create file %s in %s/%s: %w", path, org, repo, err)
				}
			} else {
				// optional: update if differs (skipped for now)
			}

		default:
			// no-op for unhandled changes
		}
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
		return "pull"
	}
}
