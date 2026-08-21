package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/DragonSecurity/gomgr/internal/config"
	"github.com/DragonSecurity/gomgr/internal/gh"
)

func drift(t *testing.T, installs []gh.Installation, dirs []config.OrgDir) string {
	t.Helper()
	var buf bytes.Buffer
	if err := reportInstallationDrift(&buf, installs, dirs); err != nil {
		t.Fatalf("reportInstallationDrift: %v", err)
	}
	return buf.String()
}

// An org someone installed the app on that nobody wrote config for. Nothing is
// applied to it, and without this nothing says so.
func TestDriftReportsInstalledButUnconfigured(t *testing.T) {
	out := drift(t,
		[]gh.Installation{{Org: "alpha", ID: 1, IsOrg: true}, {Org: "orphan", ID: 2, IsOrg: true}},
		[]config.OrgDir{{Dir: "orgs/alpha", Org: "alpha"}},
	)

	if !strings.Contains(out, "orphan") || !strings.Contains(out, "no config directory") {
		t.Errorf("the unconfigured org should be called out:\n%s", out)
	}
	if !strings.Contains(out, "nothing configures: orphan") {
		t.Errorf("expected a summary line naming it:\n%s", out)
	}
}

// The inverse: a config directory for an org the app cannot reach. Today that
// fails at authentication, late and with nothing pointing at the cause.
func TestDriftReportsConfiguredButUnreachable(t *testing.T) {
	out := drift(t,
		[]gh.Installation{{Org: "alpha", ID: 1, IsOrg: true}},
		[]config.OrgDir{{Dir: "orgs/alpha", Org: "alpha"}, {Dir: "orgs/ghost", Org: "ghost"}},
	)

	if !strings.Contains(out, "ghost") || !strings.Contains(out, "NOT installed") {
		t.Errorf("the unreachable org should be called out:\n%s", out)
	}
	if !strings.Contains(out, "fails at authentication") {
		t.Errorf("expected the consequence to be stated:\n%s", out)
	}
}

func TestDriftReportsCleanState(t *testing.T) {
	out := drift(t,
		[]gh.Installation{{Org: "alpha", ID: 1, IsOrg: true}},
		[]config.OrgDir{{Dir: "orgs/alpha", Org: "alpha"}},
	)

	if !strings.Contains(out, "No drift") {
		t.Errorf("a matching pair is not drift:\n%s", out)
	}
	for _, bad := range []string{"NOT installed", "no config directory"} {
		if strings.Contains(out, bad) {
			t.Errorf("clean state should report nothing wrong, found %q:\n%s", bad, out)
		}
	}
}

func TestDriftReportsANamelessConfigDir(t *testing.T) {
	out := drift(t,
		[]gh.Installation{{Org: "alpha", ID: 1, IsOrg: true}},
		[]config.OrgDir{{Dir: "orgs/alpha", Org: "alpha"}, {Dir: "orgs/broken", Org: ""}},
	)

	if !strings.Contains(out, "orgs/broken") || !strings.Contains(out, "names no org") {
		t.Errorf("a config directory with no org is drift too:\n%s", out)
	}
}

// Two directories naming one org is a real mistake — two configs racing to
// define the same organization — so both paths get shown rather than one.
func TestDriftShowsEveryDirectoryForAnOrg(t *testing.T) {
	out := drift(t,
		[]gh.Installation{{Org: "alpha", ID: 1, IsOrg: true}},
		[]config.OrgDir{{Dir: "orgs/alpha", Org: "alpha"}, {Dir: "orgs/alpha-copy", Org: "alpha"}},
	)

	if !strings.Contains(out, "orgs/alpha,") && !strings.Contains(out, "orgs/alpha-copy") {
		t.Errorf("both directories should appear:\n%s", out)
	}
}

// A user account cannot have a config directory, so reporting it as missing one
// would be an instruction to do the impossible. It is still shown: the
// installation is real and carries the app's permissions against that account.
func TestDriftReportsUserAccountsSeparately(t *testing.T) {
	out := drift(t,
		[]gh.Installation{
			{Org: "alpha", ID: 1, IsOrg: true},
			{Org: "aperson", ID: 2, IsOrg: false},
		},
		[]config.OrgDir{{Dir: "orgs/alpha", Org: "alpha"}},
	)

	if !strings.Contains(out, "aperson") {
		t.Errorf("the user installation is real and should be shown:\n%s", out)
	}
	if strings.Contains(out, "aperson") && strings.Contains(out, "nothing configures: aperson") {
		t.Errorf("a user account is not unconfigured drift:\n%s", out)
	}
	if !strings.Contains(out, "user account(s) the app is installed on") {
		t.Errorf("expected a separate summary line for user accounts:\n%s", out)
	}
	if !strings.Contains(out, "Remove it if nothing needs it") {
		t.Errorf("the blast-radius note is the reason to show it at all:\n%s", out)
	}
}

// With only an org installation and a matching directory, nothing about user
// accounts should appear at all.
func TestDriftSaysNothingAboutUserAccountsWhenThereAreNone(t *testing.T) {
	out := drift(t,
		[]gh.Installation{{Org: "alpha", ID: 1, IsOrg: true}},
		[]config.OrgDir{{Dir: "orgs/alpha", Org: "alpha"}},
	)
	if strings.Contains(out, "user account") {
		t.Errorf("no user installations, so no note:\n%s", out)
	}
	if !strings.Contains(out, "No drift") {
		t.Errorf("expected a clean report:\n%s", out)
	}
}

// The command is App-only, and it should say why rather than emit a confusing
// 401 from an endpoint a PAT can never answer.
func TestInstallationsRefusesPATAuthWithAnExplanation(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "ghp_irrelevant")
	t.Setenv("GITHUB_APP_ID", "")
	t.Setenv("GITHUB_APP_PRIVATE_KEY", "")

	_, _, err := runCmd(t, "installations")
	if err == nil {
		t.Fatal("expected an error explaining that this needs App credentials")
	}
	msg := err.Error()
	if !strings.Contains(msg, "GitHub App credentials") {
		t.Errorf("the message should name what is missing: %v", err)
	}
	if !strings.Contains(msg, "GITHUB_TOKEN cannot answer this") {
		t.Errorf("the message should explain why a PAT is not a fallback: %v", err)
	}
}
