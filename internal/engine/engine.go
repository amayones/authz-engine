// Package engine berisi logic inti pengambilan keputusan authorization.
package engine

import (
	"context"
	"fmt"

	"github.com/amayones/authz-engine/internal/model"
	"github.com/amayones/authz-engine/internal/store"
)

type Engine struct {
	store  store.Store
	schema model.RelationSchema
}

func New(s store.Store) *Engine {
	return &Engine{store: s}
}

func (e *Engine) CreateRole(ctx context.Context, name string, permissions []model.Permission) error {
	if name == "" {
		return fmt.Errorf("engine: nama role tidak boleh kosong")
	}
	return e.store.CreateRole(ctx, model.Role{Name: name, Permissions: permissions})
}

func (e *Engine) AssignRole(ctx context.Context, subjectID, roleName string) error {
	return e.store.AssignRole(ctx, subjectID, roleName)
}

func (e *Engine) RevokeRole(ctx context.Context, subjectID, roleName string) error {
	return e.store.RevokeRole(ctx, subjectID, roleName)
}

// Can mengevaluasi apakah request diizinkan: pertama cek RBAC (permission
// cocok), lalu kalau role punya Conditions, cek juga syarat ABAC terhadap
// Context yang dikirim di request.
func (e *Engine) Can(ctx context.Context, req model.AccessRequest) (bool, error) {
	roleNames, err := e.store.GetSubjectRoles(ctx, req.SubjectID)
	if err != nil {
		return false, fmt.Errorf("engine: gagal ambil role subject: %w", err)
	}

	needed := req.Permission()

	for _, roleName := range roleNames {
		role, err := e.store.GetRole(ctx, roleName)
		if err != nil {
			continue
		}

		hasPermission := false
		for _, perm := range role.Permissions {
			if perm == needed || perm == model.Permission(req.Resource+":*") {
				hasPermission = true
				break
			}
		}
		if !hasPermission {
			continue
		}

		if e.conditionsMet(role.Conditions, req.Context) {
			return true, nil
		}
	}

	return false, nil
}

// conditionsMet mengecek apakah semua condition pada role terpenuhi oleh
// context yang dikirim di request. Role tanpa condition otomatis lolos
// (perilaku RBAC murni).
func (e *Engine) conditionsMet(conditions []model.Condition, ctxAttrs model.Attributes) bool {
	for _, cond := range conditions {
		value, exists := ctxAttrs[cond.AttrKey]
		if !exists || value != cond.AttrValue {
			return false
		}
	}
	return true
}
