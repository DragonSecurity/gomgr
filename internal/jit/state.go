package jit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Grant represents a temporary access grant
type Grant struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`        // "team" or "repo"
	User       string    `json:"user"`
	Org        string    `json:"org"`
	Target     string    `json:"target"`      // team slug or repo name
	Permission string    `json:"permission"`  // role for teams, permission for repos
	ExpiresAt  time.Time `json:"expires_at"`
	GrantedAt  time.Time `json:"granted_at"`
}

// State manages JIT access grants
type State struct {
	mu     sync.RWMutex
	path   string
	grants map[string]*Grant
}

// NewState creates a new JIT state manager
func NewState(stateDir string) (*State, error) {
	if stateDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home dir: %w", err)
		}
		stateDir = filepath.Join(home, ".gomgr")
	}

	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, fmt.Errorf("create state dir: %w", err)
	}

	statePath := filepath.Join(stateDir, "jit-state.json")
	s := &State{
		path:   statePath,
		grants: make(map[string]*Grant),
	}

	// Load existing state
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return s, nil
}

// AddGrant adds a new JIT grant
func (s *State) AddGrant(grant *Grant) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if grant.ID == "" {
		grant.ID = fmt.Sprintf("%s-%s-%s-%d", grant.Type, grant.Org, grant.Target, time.Now().Unix())
	}

	s.grants[grant.ID] = grant
	return s.save()
}

// GetExpired returns all expired grants
func (s *State) GetExpired() []*Grant {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var expired []*Grant
	now := time.Now()
	for _, g := range s.grants {
		if g.ExpiresAt.Before(now) {
			expired = append(expired, g)
		}
	}
	return expired
}

// RemoveGrant removes a grant by ID
func (s *State) RemoveGrant(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.grants, id)
	return s.save()
}

// GetAll returns all active grants
func (s *State) GetAll() []*Grant {
	s.mu.RLock()
	defer s.mu.RUnlock()

	all := make([]*Grant, 0, len(s.grants))
	for _, g := range s.grants {
		all = append(all, g)
	}
	return all
}

// load reads state from disk
func (s *State) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	var grants []*Grant
	if err := json.Unmarshal(data, &grants); err != nil {
		return fmt.Errorf("unmarshal state: %w", err)
	}

	for _, g := range grants {
		s.grants[g.ID] = g
	}

	return nil
}

// save writes state to disk
func (s *State) save() error {
	grants := make([]*Grant, 0, len(s.grants))
	for _, g := range s.grants {
		grants = append(grants, g)
	}

	data, err := json.MarshalIndent(grants, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	return os.WriteFile(s.path, data, 0600)
}
