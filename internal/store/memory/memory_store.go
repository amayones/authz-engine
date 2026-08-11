// Package memory berisi implementasi store.Store yang menyimpan semua data
// di memori (map + mutex). Cocok untuk development dan testing.
package memory

import (
	"context"
	"sync"

	"github.com/amayones/authz-engine/internal/model"
	"github.com/amayones/authz-engine/internal/store"
)

type MemoryStore struct {
	mu sync.RWMutex

	roles             map[string]model.Role
	subjectRoles      map[string]map[string]bool
	subjectAttributes map[string]model.Attributes
	relationTuples    []model.RelationTuple
}

func New() *MemoryStore {
	return &MemoryStore{
		roles:             make(map[string]model.Role),
		subjectRoles:      make(map[string]map[string]bool),
		subjectAttributes: make(map[string]model.Attributes),
	}
}

// --- RoleStore ---

func (s *MemoryStore) CreateRole(_ context.Context, role model.Role) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.roles[role.Name]; exists {
		return store.ErrAlreadyExists
	}
	s.roles[role.Name] = role
	return nil
}

func (s *MemoryStore) GetRole(_ context.Context, name string) (model.Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	role, exists := s.roles[name]
	if !exists {
		return model.Role{}, store.ErrNotFound
	}
	return role, nil
}

func (s *MemoryStore) UpdateRole(_ context.Context, role model.Role) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.roles[role.Name]; !exists {
		return store.ErrNotFound
	}
	s.roles[role.Name] = role
	return nil
}

func (s *MemoryStore) DeleteRole(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.roles[name]; !exists {
		return store.ErrNotFound
	}
	delete(s.roles, name)

	for _, roleSet := range s.subjectRoles {
		delete(roleSet, name)
	}
	return nil
}

func (s *MemoryStore) ListRoles(_ context.Context) ([]model.Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	roles := make([]model.Role, 0, len(s.roles))
	for _, r := range s.roles {
		roles = append(roles, r)
	}
	return roles, nil
}

// --- SubjectStore ---

func (s *MemoryStore) AssignRole(_ context.Context, subjectID, roleName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.roles[roleName]; !exists {
		return store.ErrNotFound
	}
	if s.subjectRoles[subjectID] == nil {
		s.subjectRoles[subjectID] = make(map[string]bool)
	}
	s.subjectRoles[subjectID][roleName] = true
	return nil
}

func (s *MemoryStore) RevokeRole(_ context.Context, subjectID, roleName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if roleSet, exists := s.subjectRoles[subjectID]; exists {
		delete(roleSet, roleName)
	}
	return nil
}

func (s *MemoryStore) GetSubjectRoles(_ context.Context, subjectID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	roleSet, exists := s.subjectRoles[subjectID]
	if !exists {
		return []string{}, nil
	}
	roles := make([]string, 0, len(roleSet))
	for name := range roleSet {
		roles = append(roles, name)
	}
	return roles, nil
}

// --- AttributeStore ---

func (s *MemoryStore) SetAttribute(_ context.Context, subjectID, key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.subjectAttributes[subjectID] == nil {
		s.subjectAttributes[subjectID] = make(model.Attributes)
	}
	s.subjectAttributes[subjectID][key] = value
	return nil
}

func (s *MemoryStore) GetAttributes(_ context.Context, subjectID string) (model.Attributes, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	attrs, exists := s.subjectAttributes[subjectID]
	if !exists {
		return model.Attributes{}, nil
	}
	result := make(model.Attributes, len(attrs))
	for k, v := range attrs {
		result[k] = v
	}
	return result, nil
}

func (s *MemoryStore) DeleteAttribute(_ context.Context, subjectID, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if attrs, exists := s.subjectAttributes[subjectID]; exists {
		delete(attrs, key)
	}
	return nil
}

func (s *MemoryStore) WriteTuple(_ context.Context, t model.RelationTuple) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.relationTuples {
		if existing == t {
			return nil // idempotent, sudah ada
		}
	}
	s.relationTuples = append(s.relationTuples, t)
	return nil
}

func (s *MemoryStore) DeleteTuple(_ context.Context, t model.RelationTuple) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filtered := s.relationTuples[:0]
	for _, existing := range s.relationTuples {
		if existing != t {
			filtered = append(filtered, existing)
		}
	}
	s.relationTuples = filtered
	return nil
}

func (s *MemoryStore) ReadTuples(_ context.Context, object, relation string) ([]model.RelationTuple, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []model.RelationTuple
	for _, t := range s.relationTuples {
		if t.Object == object && t.Relation == relation {
			result = append(result, t)
		}
	}
	return result, nil
}
