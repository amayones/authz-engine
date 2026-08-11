// Package engine berisi logic inti pengambilan keputusan authorization.
package engine

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/amayones/authz-engine/internal/cache"
	"github.com/amayones/authz-engine/internal/model"
	"github.com/amayones/authz-engine/internal/store"
)

type Engine struct {
	store  store.Store
	schema model.RelationSchema
	cache  *cache.Cache
}

func New(s store.Store) *Engine {
	return &Engine{store: s}
}

// NewWithCache membuat engine dengan caching aktif, ttl menentukan berapa
// lama satu decision dianggap valid sebelum dihitung ulang.
func NewWithCache(s store.Store, ttl time.Duration) *Engine {
	return &Engine{store: s, cache: cache.New(ttl)}
}

func (e *Engine) CreateRole(ctx context.Context, name string, permissions []model.Permission) error {
	if name == "" {
		return fmt.Errorf("engine: nama role tidak boleh kosong")
	}
	return e.store.CreateRole(ctx, model.Role{Name: name, Permissions: permissions})
}

func (e *Engine) AssignRole(ctx context.Context, subjectID, roleName string) error {
	if err := e.store.AssignRole(ctx, subjectID, roleName); err != nil {
		return err
	}
	e.invalidateSubject(subjectID)
	return nil
}

func (e *Engine) RevokeRole(ctx context.Context, subjectID, roleName string) error {
	if err := e.store.RevokeRole(ctx, subjectID, roleName); err != nil {
		return err
	}
	e.invalidateSubject(subjectID)
	return nil
}

// invalidateSubject membuang semua cache entry milik satu subject
// (baik dari Can() maupun CheckRelation).
func (e *Engine) invalidateSubject(subjectID string) {
	if e.cache == nil {
		return
	}
	e.cache.InvalidatePrefix("can:" + subjectID)
	// Untuk rel:, subject ada di posisi ketiga, jadi prefix tidak cukup —
	// paling aman & simpel: clear seluruh cache saat ada perubahan role.
	e.cache.Clear()
}

// Can adalah titik keputusan tunggal untuk semua model authorization.
// Evaluasi berjalan dengan logika OR — kalau salah satu jalur berikut
// mengizinkan, request diizinkan:
//
//  1. RBAC + ABAC: subject punya role dengan permission yang cocok,
//     dan (kalau ada) semua Condition role tersebut terpenuhi oleh
//     req.Context.
//  2. ReBAC: kalau req.Object diisi, cek apakah subject punya relasi
//     bernama req.Action terhadap req.Object (lewat direct match,
//     hierarki schema, atau userset/grup).
//
// Kalau req.Object kosong, hanya jalur RBAC+ABAC yang dievaluasi
// (perilaku sama seperti sebelumnya, tidak breaking).
func (e *Engine) Can(ctx context.Context, req model.AccessRequest) (bool, error) {
	key := canCacheKey(req)
	if e.cache != nil {
		if val, found := e.cache.Get(key); found {
			return val, nil
		}
	}

	allowed, err := e.evaluateCan(ctx, req)
	if err != nil {
		return false, err
	}

	if e.cache != nil {
		e.cache.Set(key, allowed)
	}
	return allowed, nil
}

func (e *Engine) evaluateCan(ctx context.Context, req model.AccessRequest) (bool, error) {
	// Jalur 1: RBAC + ABAC
	rbacAllowed, err := e.evaluateRBAC(ctx, req)
	if err != nil {
		return false, err
	}
	if rbacAllowed {
		return true, nil
	}

	// Jalur 2: ReBAC — hanya dievaluasi kalau Object diisi.
	if req.Object != "" {
		rebacAllowed, err := e.checkRelation(ctx, req.Object, req.Action, req.SubjectID, make(map[string]bool))
		if err != nil {
			return false, fmt.Errorf("engine: gagal evaluasi ReBAC: %w", err)
		}
		if rebacAllowed {
			return true, nil
		}
	}

	return false, nil
}

// evaluateRBAC berisi logic RBAC+ABAC murni (dipisah dari evaluateCan
// supaya masing-masing jalur tetap mudah dibaca dan ditest terpisah).
func (e *Engine) evaluateRBAC(ctx context.Context, req model.AccessRequest) (bool, error) {
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

func (e *Engine) conditionsMet(conditions []model.Condition, ctxAttrs model.Attributes) bool {
	for _, cond := range conditions {
		value, exists := ctxAttrs[cond.AttrKey]
		if !exists || value != cond.AttrValue {
			return false
		}
	}
	return true
}

// canCacheKey membangun key unik per kombinasi subject+resource+action+
// object+context. Context diurutkan supaya key deterministik.
func canCacheKey(req model.AccessRequest) string {
	var b strings.Builder
	b.WriteString("can:")
	b.WriteString(req.SubjectID)
	b.WriteString("|")
	b.WriteString(req.Resource)
	b.WriteString("|")
	b.WriteString(req.Action)
	b.WriteString("|")
	b.WriteString(req.Object)

	keys := make([]string, 0, len(req.Context))
	for k := range req.Context {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteString("|")
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(req.Context[k])
	}
	return b.String()
}

// CreateRoleWithConditions sama seperti CreateRole, tapi mendukung
// pengisian Conditions (ABAC) sekaligus saat pembuatan role.
func (e *Engine) CreateRoleWithConditions(ctx context.Context, role model.Role) error {
	if role.Name == "" {
		return fmt.Errorf("engine: nama role tidak boleh kosong")
	}
	return e.store.CreateRole(ctx, role)
}

// SetAttribute mengatur satu atribut milik subject, dipakai untuk ABAC.
func (e *Engine) SetAttribute(ctx context.Context, subjectID, key, value string) error {
	if err := e.store.SetAttribute(ctx, subjectID, key, value); err != nil {
		return err
	}
	if e.cache != nil {
		e.cache.Clear()
	}
	return nil
}
