package jit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewState(t *testing.T) {
	tmpDir := t.TempDir()

	state, err := NewState(tmpDir)
	if err != nil {
		t.Fatalf("NewState() error = %v", err)
	}

	if state == nil {
		t.Fatal("NewState() returned nil")
	}

	// Check that state file was created
	expectedPath := filepath.Join(tmpDir, "jit-state.json")
	if state.path != expectedPath {
		t.Errorf("state.path = %q, want %q", state.path, expectedPath)
	}
}

func TestAddGrant(t *testing.T) {
	tmpDir := t.TempDir()
	state, err := NewState(tmpDir)
	if err != nil {
		t.Fatalf("NewState() error = %v", err)
	}

	grant := &Grant{
		Type:       "team",
		User:       "testuser",
		Org:        "testorg",
		Target:     "developers",
		Permission: "member",
		GrantedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(1 * time.Hour),
	}

	err = state.AddGrant(grant)
	if err != nil {
		t.Fatalf("AddGrant() error = %v", err)
	}

	// Check that grant was added
	if grant.ID == "" {
		t.Error("AddGrant() did not set grant ID")
	}

	// Check that grant is in state
	all := state.GetAll()
	if len(all) != 1 {
		t.Errorf("GetAll() returned %d grants, want 1", len(all))
	}
}

func TestGetExpired(t *testing.T) {
	tmpDir := t.TempDir()
	state, err := NewState(tmpDir)
	if err != nil {
		t.Fatalf("NewState() error = %v", err)
	}

	now := time.Now()

	// Add expired grant
	expiredGrant := &Grant{
		Type:       "team",
		User:       "testuser1",
		Org:        "testorg",
		Target:     "developers",
		Permission: "member",
		GrantedAt:  now.Add(-2 * time.Hour),
		ExpiresAt:  now.Add(-1 * time.Hour),
	}
	state.AddGrant(expiredGrant)

	// Add active grant
	activeGrant := &Grant{
		Type:       "repo",
		User:       "testuser2",
		Org:        "testorg",
		Target:     "myrepo",
		Permission: "push",
		GrantedAt:  now,
		ExpiresAt:  now.Add(1 * time.Hour),
	}
	state.AddGrant(activeGrant)

	// Get expired grants
	expired := state.GetExpired()
	if len(expired) != 1 {
		t.Errorf("GetExpired() returned %d grants, want 1", len(expired))
	}

	if len(expired) > 0 && expired[0].User != "testuser1" {
		t.Errorf("GetExpired() returned wrong grant, got user %q, want %q", expired[0].User, "testuser1")
	}
}

func TestRemoveGrant(t *testing.T) {
	tmpDir := t.TempDir()
	state, err := NewState(tmpDir)
	if err != nil {
		t.Fatalf("NewState() error = %v", err)
	}

	grant := &Grant{
		Type:       "team",
		User:       "testuser",
		Org:        "testorg",
		Target:     "developers",
		Permission: "member",
		GrantedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(1 * time.Hour),
	}

	err = state.AddGrant(grant)
	if err != nil {
		t.Fatalf("AddGrant() error = %v", err)
	}

	grantID := grant.ID

	// Remove grant
	err = state.RemoveGrant(grantID)
	if err != nil {
		t.Fatalf("RemoveGrant() error = %v", err)
	}

	// Check that grant was removed
	all := state.GetAll()
	if len(all) != 0 {
		t.Errorf("GetAll() returned %d grants after removal, want 0", len(all))
	}
}

func TestStatePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	state1, err := NewState(tmpDir)
	if err != nil {
		t.Fatalf("NewState() error = %v", err)
	}

	grant := &Grant{
		Type:       "team",
		User:       "testuser",
		Org:        "testorg",
		Target:     "developers",
		Permission: "member",
		GrantedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(1 * time.Hour),
	}

	err = state1.AddGrant(grant)
	if err != nil {
		t.Fatalf("AddGrant() error = %v", err)
	}

	// Create new state instance (should load from disk)
	state2, err := NewState(tmpDir)
	if err != nil {
		t.Fatalf("NewState() error = %v", err)
	}

	// Check that grant was loaded
	all := state2.GetAll()
	if len(all) != 1 {
		t.Errorf("GetAll() returned %d grants after reload, want 1", len(all))
	}

	if len(all) > 0 && all[0].User != "testuser" {
		t.Errorf("loaded grant has user %q, want %q", all[0].User, "testuser")
	}
}

func TestStateFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	state, err := NewState(tmpDir)
	if err != nil {
		t.Fatalf("NewState() error = %v", err)
	}

	grant := &Grant{
		Type:       "team",
		User:       "testuser",
		Org:        "testorg",
		Target:     "developers",
		Permission: "member",
		GrantedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(1 * time.Hour),
	}

	err = state.AddGrant(grant)
	if err != nil {
		t.Fatalf("AddGrant() error = %v", err)
	}

	// Check file permissions
	info, err := os.Stat(state.path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	mode := info.Mode().Perm()
	expected := os.FileMode(0600)
	if mode != expected {
		t.Errorf("state file has permissions %v, want %v", mode, expected)
	}
}
