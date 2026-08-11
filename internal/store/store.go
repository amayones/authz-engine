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

// RelationStore mengelola relation tuple untuk ReBAC.
type RelationStore interface {
	WriteTuple(ctx context.Context, t model.RelationTuple) error
	DeleteTuple(ctx context.Context, t model.RelationTuple) error

	// ReadTuples mengambil semua tuple pada object dengan relation
	// tertentu — dipakai engine untuk traversal graph.
	ReadTuples(ctx context.Context, object, relation string) ([]model.RelationTuple, error)
}

// APIKeyStore mengelola kredensial client untuk HTTP API layer.
type APIKeyStore interface {
	// CreateAPIKey menyimpan hash dari key, bukan key mentahnya.
	CreateAPIKey(ctx context.Context, keyHash string, key model.APIKey) error
	// GetAPIKeyByHash dipakai saat autentikasi tiap request masuk.
	GetAPIKeyByHash(ctx context.Context, keyHash string) (model.APIKey, error)
	RevokeAPIKey(ctx context.Context, keyHash string) error
}

type AuditStore interface {
	RecordAudit(ctx context.Context, entry model.AuditEntry) error
}

// Store adalah gabungan RoleStore + SubjectStore + AttributeStore.
// Ini yang dipakai oleh engine.
type Store interface {
	RoleStore
	SubjectStore
	AttributeStore
	RelationStore
	APIKeyStore
	AuditStore
}
