package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// runCmd invokes rootCmd with the given args, capturing anything written to
// either the cobra Out/Err sinks or os.Stdout (several commands use fmt.Println
// directly). The caller gets stdout, stderr, and the command's error back.
//
// Tests that use this helper must not run in parallel — rootCmd is a package
// singleton with shared flag state.
func runCmd(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	cfgDir = ""
	debug = false
	dryRun = false
	timeout = 10 * time.Minute
	auditLog = false
	teamName = ""
	outFile = ""
	resetFlagsChanged(rootCmd)

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	rootCmd.SetOut(outBuf)
	rootCmd.SetErr(errBuf)

	origStdout := os.Stdout
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("pipe: %v", pipeErr)
	}
	os.Stdout = w

	rootCmd.SetArgs(args)
	execErr := rootCmd.Execute()

	_ = w.Close()
	os.Stdout = origStdout
	var stdoutBuf bytes.Buffer
	_, _ = io.Copy(&stdoutBuf, r)

	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})

	return stdoutBuf.String() + outBuf.String(), errBuf.String(), execErr
}

// resetFlagsChanged walks a cobra command tree and clears the Changed flag on
// every pflag so required-flag detection works across sequential Execute calls.
func resetFlagsChanged(c *cobra.Command) {
	clear := func(f *pflag.Flag) { f.Changed = false }
	c.Flags().VisitAll(clear)
	c.PersistentFlags().VisitAll(clear)
	for _, child := range c.Commands() {
		resetFlagsChanged(child)
	}
}

// writeConfigDir builds a minimal but valid gomgr config tree under dir and
// returns the dir path. The team file references a repo called "api".
func writeConfigDir(t *testing.T, dir string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "teams"), 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(path, contents string) {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(dir, "app.yaml"), "org: testorg\n")
	write(filepath.Join(dir, "org.yaml"), "owners:\n  - alice\n")
	write(filepath.Join(dir, "teams", "backend.yaml"), `name: Backend
slug: backend
privacy: closed
maintainers:
  - alice
members:
  - bob
repositories:
  api:
    permission: push
`)
	return dir
}

func TestVersionCommand(t *testing.T) {
	stdout, _, err := runCmd(t, "version")
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if !strings.Contains(stdout, "Version:") {
		t.Errorf("expected 'Version:' in output, got %q", stdout)
	}
}

func TestValidate_MissingConfigFlag(t *testing.T) {
	_, _, err := runCmd(t, "validate")
	if err == nil {
		t.Fatal("expected error when --config is missing")
	}
	if !strings.Contains(err.Error(), "--config") {
		t.Errorf("expected error to mention --config, got %v", err)
	}
}

func TestValidate_NonexistentConfigDir(t *testing.T) {
	_, _, err := runCmd(t, "validate", "-c", "/nonexistent/path/gomgr-test")
	if err == nil {
		t.Fatal("expected error for nonexistent config dir")
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	dir := writeConfigDir(t, t.TempDir())
	stdout, _, err := runCmd(t, "validate", "-c", dir)
	if err != nil {
		t.Fatalf("validate failed on good config: %v", err)
	}
	if !strings.Contains(stdout, "Configuration is valid") {
		t.Errorf("expected success message, got %q", stdout)
	}
}

func TestValidate_InvalidConfig_EmptyOrg(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.yaml"), []byte("org: \"\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "org.yaml"), []byte("owners: []\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err := runCmd(t, "validate", "-c", dir)
	if err == nil {
		t.Fatal("expected validation error for empty org")
	}
	if !strings.Contains(err.Error(), "org") {
		t.Errorf("expected error to mention 'org', got %v", err)
	}
}

func TestValidate_InvalidConfig_BadTeamPrivacy(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "teams"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile := func(path, contents string) {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	writeFile(filepath.Join(dir, "app.yaml"), "org: testorg\n")
	writeFile(filepath.Join(dir, "org.yaml"), "owners:\n  - alice\n")
	writeFile(filepath.Join(dir, "teams", "bad.yaml"), "name: Bad\nprivacy: open\n")

	_, _, err := runCmd(t, "validate", "-c", dir)
	if err == nil {
		t.Fatal("expected validation error for bad privacy")
	}
	if !strings.Contains(err.Error(), "privacy") {
		t.Errorf("expected error to mention 'privacy', got %v", err)
	}
}

func TestSetupTeam_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	_, _, err := runCmd(t, "setup-team", "-c", dir, "-n", "Backend Team")
	if err != nil {
		t.Fatalf("setup-team failed: %v", err)
	}
	// slug should be lowercase with dashes.
	expected := filepath.Join(dir, "teams", "backend-team.yaml")
	if _, err := os.Stat(expected); err != nil {
		t.Fatalf("expected team file at %s, got: %v", expected, err)
	}
	b, err := os.ReadFile(expected)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "Backend Team") {
		t.Errorf("expected team file to contain team name, got %q", string(b))
	}
}

func TestSetupTeam_MissingNameFlag(t *testing.T) {
	dir := t.TempDir()
	_, _, err := runCmd(t, "setup-team", "-c", dir)
	if err == nil {
		t.Fatal("expected error when --name is missing")
	}
}

func TestSetupTeam_ExplicitOutFile(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "frontend.yaml")
	_, _, err := runCmd(t, "setup-team", "-c", dir, "-n", "Frontend", "-f", out)
	if err != nil {
		t.Fatalf("setup-team failed: %v", err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("expected file at %s, got: %v", out, err)
	}
}

func TestSync_MissingConfigFlag(t *testing.T) {
	_, _, err := runCmd(t, "sync")
	if err == nil {
		t.Fatal("expected error when --config is missing")
	}
	if !strings.Contains(err.Error(), "--config") {
		t.Errorf("expected error to mention --config, got %v", err)
	}
}

func TestRoot_UnknownCommand(t *testing.T) {
	_, _, err := runCmd(t, "does-not-exist")
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}
