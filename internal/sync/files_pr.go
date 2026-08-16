package sync

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v90/github"

	"github.com/DragonSecurity/gomgr/internal/gh"
	"github.com/DragonSecurity/gomgr/internal/util"
)

// scopeRepoFilePR is the change scope for a file that has to reach its branch
// through a pull request rather than a direct commit.
const scopeRepoFilePR = "repo-file-pr"

// applyRepoFilePullRequest lands one file change through a branch and a pull
// request, because a ruleset the configuration declares would reject a direct
// push to the target branch.
//
// One head branch is reused per repository, so several files in the same sync
// become several commits on one pull request rather than one pull request each.
// The change marked "final" is the one that turns auto-merge on, so GitHub
// cannot merge the pull request while gomgr is still adding commits to it.
func applyRepoFilePullRequest(ctx context.Context, c *gh.Client, ch util.Change) error {
	d, err := extractDetails(ch)
	if err != nil {
		return err
	}
	org := detailString(d, "org")
	repo := detailString(d, "repo")
	path := detailString(d, "path")
	content := []byte(detailString(d, "content"))
	message := detailString(d, "message")
	base := detailString(d, "branch")
	head := detailString(d, "head_branch")
	mergeMethod := detailString(d, "merge_method")

	if err := ensureHeadBranch(ctx, c, org, repo, base, head); err != nil {
		return err
	}
	if err := writeFileOnBranch(ctx, c, org, repo, path, head, message, content); err != nil {
		return err
	}

	pr, err := ensurePullRequest(ctx, c, org, repo, base, head, message)
	if err != nil {
		return err
	}

	if detailBool(d, "final") && detailBool(d, "auto_merge") {
		if err := enableAutoMerge(ctx, c, pr.GetNodeID(), mergeMethod); err != nil {
			// Auto-merge is a repository setting and can also be refused when a
			// branch has no required checks to wait for. Neither is a reason to
			// fail the sync: the pull request exists and carries the change.
			util.Warnf("%s/%s PR #%d: could not enable auto-merge (%v); the pull request is open and waiting",
				org, repo, pr.GetNumber(), err)
		}
	}
	util.Infof("%s/%s PR #%d: %s", org, repo, pr.GetNumber(), pr.GetHTMLURL())
	return nil
}

// ensureHeadBranch creates the pull request's head branch from base when it is
// not already there. An existing branch is left alone: it usually carries an
// open pull request from an earlier run that this change should join.
func ensureHeadBranch(ctx context.Context, c *gh.Client, org, repo, base, head string) error {
	if _, _, err := c.REST.Git.GetRef(ctx, org, repo, "refs/heads/"+head); err == nil {
		return nil
	} else if !isNotFound(err) {
		return fmt.Errorf("check branch %s on %s/%s: %w", head, org, repo, err)
	}

	baseRef, _, err := c.REST.Git.GetRef(ctx, org, repo, "refs/heads/"+base)
	if err != nil {
		return fmt.Errorf("read branch %s on %s/%s: %w", base, org, repo, err)
	}
	_, _, err = c.REST.Git.CreateRef(ctx, org, repo, github.CreateRef{
		Ref: "refs/heads/" + head,
		SHA: baseRef.GetObject().GetSHA(),
	})
	if err != nil && !isRefAlreadyExists(err) {
		return fmt.Errorf("create branch %s on %s/%s: %w", head, org, repo, err)
	}
	return nil
}

// writeFileOnBranch creates or updates the file on the head branch, and does
// nothing when the branch already carries the wanted content — a sync that has
// already pushed this change should not add an empty commit on the next run.
func writeFileOnBranch(ctx context.Context, c *gh.Client, org, repo, path, branch, message string, content []byte) error {
	file, _, resp, err := c.REST.Repositories.GetContents(ctx, org, repo, path,
		&github.RepositoryContentGetOptions{Ref: branch})
	if err != nil && (resp == nil || resp.StatusCode != http.StatusNotFound) {
		return fmt.Errorf("read %s/%s:%s on %s: %w", org, repo, path, branch, err)
	}

	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(message),
		Content: content,
		Branch:  github.Ptr(branch),
	}
	if file == nil {
		if _, _, err := c.REST.Repositories.CreateFile(ctx, org, repo, path, opts); err != nil {
			return fmt.Errorf("create %s on %s/%s@%s: %w", path, org, repo, branch, err)
		}
		return nil
	}

	current, err := file.GetContent()
	if err != nil {
		return fmt.Errorf("decode %s on %s/%s@%s: %w", path, org, repo, branch, err)
	}
	if current == string(content) {
		return nil
	}
	opts.SHA = github.Ptr(file.GetSHA())
	if _, _, err := c.REST.Repositories.UpdateFile(ctx, org, repo, path, opts); err != nil {
		return fmt.Errorf("update %s on %s/%s@%s: %w", path, org, repo, branch, err)
	}
	return nil
}

// ensurePullRequest returns the open pull request from head into base, opening
// one if there is not already one to add to.
func ensurePullRequest(ctx context.Context, c *gh.Client, org, repo, base, head, message string) (*github.PullRequest, error) {
	existing, _, err := c.REST.PullRequests.List(ctx, org, repo, &github.PullRequestListOptions{
		State: "open",
		Head:  org + ":" + head,
		Base:  base,
	})
	if err != nil {
		return nil, fmt.Errorf("list pull requests on %s/%s: %w", org, repo, err)
	}
	if len(existing) > 0 {
		return existing[0], nil
	}

	title, body := splitCommitMessage(message)
	pr, _, err := c.REST.PullRequests.Create(ctx, org, repo, github.CreatePullRequest{
		Title: github.Ptr(title),
		Head:  head,
		Base:  base,
		Body:  github.Ptr(body),
	})
	if err != nil {
		return nil, fmt.Errorf("open pull request on %s/%s: %w", org, repo, err)
	}
	return pr, nil
}

// splitCommitMessage turns a commit message into a pull request title and body.
// The subject becomes the title; the rest, sign-off trailer included, becomes
// the body, so the merge commit still carries the trailer a DCO rule wants.
func splitCommitMessage(message string) (title, body string) {
	subject, rest, found := strings.Cut(message, "\n")
	title = strings.TrimSpace(subject)
	if title == "" {
		title = "chore: sync managed files"
	}
	if !found {
		return title, "Opened by gomgr to apply managed file content.\n"
	}
	return title, "Opened by gomgr to apply managed file content.\n\n" + strings.TrimSpace(rest) + "\n"
}

// enableAutoMerge asks GitHub to merge the pull request once its required
// checks pass. Auto-merge is only exposed through GraphQL.
func enableAutoMerge(ctx context.Context, c *gh.Client, prNodeID, mergeMethod string) error {
	if prNodeID == "" {
		return errors.New("pull request has no node ID")
	}
	const mutation = `
mutation($pullRequestId: ID!, $mergeMethod: PullRequestMergeMethod!) {
  enablePullRequestAutoMerge(input: {pullRequestId: $pullRequestId, mergeMethod: $mergeMethod}) {
    clientMutationId
  }
}`
	return c.DoGraphQL(ctx, mutation, map[string]any{
		"pullRequestId": prNodeID,
		"mergeMethod":   strings.ToUpper(mergeMethod),
	}, nil)
}

// isRefAlreadyExists reports whether err is GitHub refusing to create a ref
// that another run created a moment ago.
func isRefAlreadyExists(err error) bool {
	var ghErr *github.ErrorResponse
	if !errors.As(err, &ghErr) || ghErr.Response == nil {
		return false
	}
	return containsErrorMessage(ghErr, errTermRefExists, errTermNameExists)
}
