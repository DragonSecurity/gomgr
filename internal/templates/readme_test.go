package templates

import (
	"strings"
	"testing"
)

func TestGenerateReadme(t *testing.T) {
	org := "TestOrg"
	repo := "test-repo"

	content, err := GenerateReadme(org, repo)
	if err != nil {
		t.Fatalf("GenerateReadme failed: %v", err)
	}

	// Verify content has expected elements
	expectedElements := []string{
		"# test-repo",
		"git clone git@github.com:TestOrg/test-repo.git",
		"git remote add origin git@github.com:TestOrg/test-repo.git",
		"Quick setup",
		"README, LICENSE, and .gitignore",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(content, expected) {
			t.Errorf("Generated README missing expected element: %q", expected)
		}
	}

	// Verify the repo name appears multiple times
	if strings.Count(content, repo) < 3 {
		t.Errorf("Expected repo name to appear at least 3 times, found %d", strings.Count(content, repo))
	}

	// Verify the org name appears multiple times
	if strings.Count(content, org) < 3 {
		t.Errorf("Expected org name to appear at least 3 times, found %d", strings.Count(content, org))
	}
}

func TestGenerateReadmeConsistency(t *testing.T) {
	// Test that generating the same README twice produces the same result
	org := "ConsistentOrg"
	repo := "consistent-repo"

	content1, err1 := GenerateReadme(org, repo)
	content2, err2 := GenerateReadme(org, repo)

	if err1 != nil || err2 != nil {
		t.Fatalf("GenerateReadme failed: %v, %v", err1, err2)
	}

	if content1 != content2 {
		t.Error("GenerateReadme should produce consistent results")
	}
}
