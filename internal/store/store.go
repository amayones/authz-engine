// Package store mendefinisikan kontrak (interface) untuk penyimpanan data
// authorization: role, permission, assignment, dan attribute.
package store

import (
	"context"
	"errors"

	"github.com/amayones/authz-engine/internal/model"
)

var ErrNotFound = errors.New("store: data not found")
var ErrAlreadyExists = errors.New("store: data already exists")

// RoleStore mengelola definisi role beserta permission-nya.
type RoleStore interface {
	CreateRole(ctx context.Context, role model.Role) error
	GetRole(ctx context.Context, name string) (model.Role, error)
	UpdateRole(ctx context.Context, role model.Role) error
	DeleteRole(ctx context.Context, name string) error
	ListRoles(ctx context.Context) ([]model.Role, error)
}

// SubjectStore mengelola assignment role ke subject (user/service).
type SubjectStore interface {
	AssignRole(ctx context.Context, subjectID, roleName string) error
	RevokeRole(ctx context.Context, subjectID, roleName string) error
	GetSubjectRoles(ctx context.Context, subjectID string) ([]string, error)
}

// AttributeStore mengelola atribut milik subject, dipakai untuk ABAC.
type AttributeStore interface {
	SetAttribute(ctx context.Context, subjectID, key, value string) error
	GetAttributes(ctx context.Context, subjectID string) (model.Attributes, error)
	DeleteAttribute(ctx context.Context, subjectID, key string) error
}

// Store adalah gabungan RoleStore + SubjectStore + AttributeStore.
// Ini yang dipakai oleh engine.
type Store interface {
	RoleStore
	SubjectStore
	AttributeStore
}
