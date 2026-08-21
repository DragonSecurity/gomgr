package gh

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/go-github/v90/github"
)

// accountTypeOrganization is what GitHub calls an installation account that is
// an organization, as opposed to a user.
const accountTypeOrganization = "Organization"

// Installation is one account a GitHub App is installed on.
type Installation struct {
	// Org is the account's login, lowercased, because every other org name in
	// gomgr is compared lowercased. Named Org because that is what it is for
	// every installation gomgr can act on — see IsOrg for the ones it cannot.
	Org string
	// ID is the installation ID, which is what the token endpoints want.
	ID int64
	// RepositorySelection is "all" or "selected". An app installed on a
	// selected subset can be installed on the org and still be unable to see
	// the repository a config names, so the two are not the same answer.
	RepositorySelection string
	// IsOrg reports whether this installation is on an organization rather
	// than a user account.
	//
	// An app can be installed on a personal account, and gomgr can never act on
	// one: auth resolves through GetOrganizationInstallation, which answers 404
	// for a user. Such an installation is still worth reporting — it is real,
	// and it carries the same permissions against that account's repositories —
	// but it must not be offered as somewhere a configuration could point.
	IsOrg bool
}

// ListInstallations returns every organization the authenticated GitHub App is
// installed on, sorted by org.
//
// This answers to the app, not to one of its installations, so it needs a
// client from AppClient — an installation token cannot see its siblings, which
// is the whole reason this is worth a separate call.
//
// It needs no enterprise scope. That matters: it is the one honest answer gomgr
// can give to "which organizations does this reach" without a credential that
// could also rewrite them.
func ListInstallations(ctx context.Context, appClient *github.Client) ([]Installation, error) {
	var out []Installation
	opts := &github.ListOptions{PerPage: 100}
	for {
		installs, resp, err := appClient.Apps.ListInstallations(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("list app installations: %w", err)
		}
		for _, in := range installs {
			login := in.GetAccount().GetLogin()
			if login == "" {
				continue
			}
			out = append(out, Installation{
				Org:                 strings.ToLower(login),
				ID:                  in.GetID(),
				RepositorySelection: in.GetRepositorySelection(),
				IsOrg:               in.GetAccount().GetType() == accountTypeOrganization,
			})
		}
		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Org < out[j].Org })
	return out, nil
}
