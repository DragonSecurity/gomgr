package sync

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v90/github"

	"github.com/DragonSecurity/gomgr/internal/gh"
)

// allTeamMembers returns every team in the organization and its members,
// keyed by team slug then by lowercased login. Maintainers are members here:
// the distinction matters when deciding someone's role in a team and not at
// all when deciding what a team entitles them to.
//
// The result is cached on State. It costs a request per team, and both member
// cleanup and collaborator enforcement want the same walk.
func (s *State) allTeamMembers(ctx context.Context, c *gh.Client) (map[string]map[string]bool, error) {
	if s.teamMembers != nil {
		return s.teamMembers, nil
	}
	out := make(map[string]map[string]bool, len(s.ActualTeams))
	for _, t := range s.ActualTeams {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		slug := t.GetSlug()
		logins := map[string]bool{}
		opt := &github.TeamListTeamMembersOptions{
			Role:        "all",
			ListOptions: github.ListOptions{PerPage: defaultPerPage},
		}
		if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
			opt.ListOptions = *opts
			users, resp, err := c.REST.Teams.ListTeamMembersBySlug(ctx, s.Org, slug, opt)
			if err != nil {
				return nil, err
			}
			for _, u := range users {
				logins[strings.ToLower(u.GetLogin())] = true
			}
			return resp, nil
		}); err != nil {
			return nil, fmt.Errorf("list members of team %q: %w", slug, err)
		}
		out[slug] = logins
	}
	s.teamMembers = out
	return out, nil
}

// allTeamRepos returns every team's repository permissions, keyed by team slug
// then by lowercased repository name.
//
// A team that has just been deleted out from under the run answers 404; that is
// not a reason to fail a plan, so it contributes nothing and the walk goes on.
func (s *State) allTeamRepos(ctx context.Context, c *gh.Client) (map[string]map[string]string, error) {
	if s.teamRepos != nil {
		return s.teamRepos, nil
	}
	out := make(map[string]map[string]string, len(s.ActualTeams))
	for _, t := range s.ActualTeams {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		slug := t.GetSlug()
		perms := map[string]string{}
		opt := &github.ListOptions{PerPage: defaultPerPage}
		if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
			*opt = *opts
			repos, resp, err := c.REST.Teams.ListTeamReposBySlug(ctx, s.Org, slug, opt)
			if err != nil {
				var ghErr *github.ErrorResponse
				if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
					return &github.Response{}, nil
				}
				return nil, err
			}
			for _, r := range repos {
				perms[strings.ToLower(r.GetName())] = extractRepoPerm(r)
			}
			return resp, nil
		}); err != nil {
			return nil, fmt.Errorf("list repositories of team %q: %w", slug, err)
		}
		out[slug] = perms
	}
	s.teamRepos = out
	return out, nil
}

// orgMembers returns every member of the organization, lowercased. Owners are
// included: they are members too, and a caller that needs to tell them apart
// has fetchOrgOwners for that.
func (s *State) orgMembers(ctx context.Context, c *gh.Client) (map[string]bool, error) {
	opt := &github.ListMembersOptions{
		Role:        "all",
		ListOptions: github.ListOptions{PerPage: defaultPerPage},
	}
	out := map[string]bool{}
	if err := paginate(func(opts *github.ListOptions) (*github.Response, error) {
		opt.ListOptions = *opts
		users, resp, err := c.REST.Organizations.ListMembers(ctx, s.Org, opt)
		if err != nil {
			return nil, err
		}
		for _, u := range users {
			out[strings.ToLower(u.GetLogin())] = true
		}
		return resp, nil
	}); err != nil {
		return nil, fmt.Errorf("list members of org %q: %w", s.Org, err)
	}
	return out, nil
}

// permNone is GitHub's name for an organization default that grants nothing.
const permNone = "none"

// orgDefaultRepoPermission returns the base permission every organization
// member holds on every repository: permNone, read, write or admin.
//
// Without it, a direct grant of push in an organization whose default is
// already write reads as excess access when it confers nothing at all.
func (s *State) orgDefaultRepoPermission(ctx context.Context, c *gh.Client) (string, error) {
	if s.defaultRepoPermission != "" {
		return s.defaultRepoPermission, nil
	}
	o, _, err := c.REST.Organizations.Get(ctx, s.Org)
	if err != nil {
		return "", fmt.Errorf("read default repository permission of org %q: %w", s.Org, err)
	}
	perm := o.GetDefaultRepoPermission()
	if perm == "" || perm == permNone {
		perm = permNone
	}
	s.defaultRepoPermission = perm
	return perm, nil
}
