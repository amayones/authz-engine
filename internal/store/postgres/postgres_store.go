// Package postgres berisi implementasi store.Store menggunakan PostgreSQL
// sebagai backend persistent. Ini alternatif dari mssql untuk deployment
// di platform PaaS gratis (Render, Railway, Fly.io, dll).
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/amayones/authz-engine/internal/model"
	"github.com/amayones/authz-engine/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStore struct {
	db *sql.DB
}

// PoolConfig mengatur batas connection pool.
type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

// New membuka koneksi ke PostgreSQL. connString formatnya:
// "postgres://user:password@host:5432/dbname?sslmode=require"
func New(connString string) (*PostgresStore, error) {
	return NewWithConfig(connString, DefaultPoolConfig())
}

func NewWithConfig(connString string, cfg PoolConfig) (*PostgresStore, error) {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, fmt.Errorf("postgres: gagal buka koneksi: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("postgres: gagal ping database: %w", err)
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

func (s *PostgresStore) Stats() sql.DBStats {
	return s.db.Stats()
}

// --- RoleStore ---

func (s *PostgresStore) CreateRole(ctx context.Context, role model.Role) error {
	permJSON, err := json.Marshal(role.Permissions)
	if err != nil {
		return fmt.Errorf("postgres: gagal encode permissions: %w", err)
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO roles (name, permissions) VALUES ($1, $2)`,
		role.Name, string(permJSON))
	if err != nil {
		return fmt.Errorf("postgres: gagal insert role: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetRole(ctx context.Context, name string) (model.Role, error) {
	var permJSON string
	err := s.db.QueryRowContext(ctx,
		`SELECT permissions FROM roles WHERE name = $1`, name,
	).Scan(&permJSON)

	if err == sql.ErrNoRows {
		return model.Role{}, store.ErrNotFound
	}
	if err != nil {
		return model.Role{}, fmt.Errorf("postgres: gagal query role: %w", err)
	}

	var perms []model.Permission
	if err := json.Unmarshal([]byte(permJSON), &perms); err != nil {
		return model.Role{}, fmt.Errorf("postgres: gagal decode permissions: %w", err)
	}
	return model.Role{Name: name, Permissions: perms}, nil
}

func (s *PostgresStore) UpdateRole(ctx context.Context, role model.Role) error {
	permJSON, err := json.Marshal(role.Permissions)
	if err != nil {
		return fmt.Errorf("postgres: gagal encode permissions: %w", err)
	}
	result, err := s.db.ExecContext(ctx,
		`UPDATE roles SET permissions = $1 WHERE name = $2`,
		string(permJSON), role.Name)
	if err != nil {
		return fmt.Errorf("postgres: gagal update role: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (s *PostgresStore) DeleteRole(ctx context.Context, name string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM roles WHERE name = $1`, name)
	if err != nil {
		return fmt.Errorf("postgres: gagal delete role: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (s *PostgresStore) ListRoles(ctx context.Context) ([]model.Role, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT name, permissions FROM roles`)
	if err != nil {
		return nil, fmt.Errorf("postgres: gagal list roles: %w", err)
	}
	defer rows.Close()

	var roles []model.Role
	for rows.Next() {
		var name, permJSON string
		if err := rows.Scan(&name, &permJSON); err != nil {
			return nil, fmt.Errorf("postgres: gagal scan row: %w", err)
		}
		var perms []model.Permission
		if err := json.Unmarshal([]byte(permJSON), &perms); err != nil {
			return nil, fmt.Errorf("postgres: gagal decode permissions: %w", err)
		}
		roles = append(roles, model.Role{Name: name, Permissions: perms})
	}
	return roles, nil
}

// --- SubjectStore ---

func (s *PostgresStore) AssignRole(ctx context.Context, subjectID, roleName string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO subject_roles (subject_id, role_name) VALUES ($1, $2)
		ON CONFLICT (subject_id, role_name) DO NOTHING`,
		subjectID, roleName)
	if err != nil {
		return fmt.Errorf("postgres: gagal assign role: %w", err)
	}
	return nil
}

func (s *PostgresStore) RevokeRole(ctx context.Context, subjectID, roleName string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM subject_roles WHERE subject_id = $1 AND role_name = $2`,
		subjectID, roleName)
	if err != nil {
		return fmt.Errorf("postgres: gagal revoke role: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetSubjectRoles(ctx context.Context, subjectID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT role_name FROM subject_roles WHERE subject_id = $1`, subjectID)
	if err != nil {
		return nil, fmt.Errorf("postgres: gagal query subject roles: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("postgres: gagal scan row: %w", err)
		}
		roles = append(roles, name)
	}
	return roles, nil
}

// --- AttributeStore ---

func (s *PostgresStore) SetAttribute(ctx context.Context, subjectID, key, value string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO subject_attributes (subject_id, attr_key, attr_value)
		VALUES ($1, $2, $3)
		ON CONFLICT (subject_id, attr_key) DO UPDATE SET attr_value = EXCLUDED.attr_value`,
		subjectID, key, value)
	if err != nil {
		return fmt.Errorf("postgres: gagal set attribute: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetAttributes(ctx context.Context, subjectID string) (model.Attributes, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT attr_key, attr_value FROM subject_attributes WHERE subject_id = $1`, subjectID)
	if err != nil {
		return nil, fmt.Errorf("postgres: gagal query attributes: %w", err)
	}
	defer rows.Close()

	attrs := make(model.Attributes)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("postgres: gagal scan row: %w", err)
		}
		attrs[k] = v
	}
	return attrs, nil
}

func (s *PostgresStore) DeleteAttribute(ctx context.Context, subjectID, key string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM subject_attributes WHERE subject_id = $1 AND attr_key = $2`,
		subjectID, key)
	if err != nil {
		return fmt.Errorf("postgres: gagal delete attribute: %w", err)
	}
	return nil
}

// --- RelationStore ---

func (s *PostgresStore) WriteTuple(ctx context.Context, t model.RelationTuple) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO relation_tuples (object_id, relation, subject) VALUES ($1, $2, $3)
		ON CONFLICT (object_id, relation, subject) DO NOTHING`,
		t.Object, t.Relation, t.Subject)
	if err != nil {
		return fmt.Errorf("postgres: gagal write tuple: %w", err)
	}
	return nil
}

func (s *PostgresStore) DeleteTuple(ctx context.Context, t model.RelationTuple) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM relation_tuples WHERE object_id = $1 AND relation = $2 AND subject = $3`,
		t.Object, t.Relation, t.Subject)
	if err != nil {
		return fmt.Errorf("postgres: gagal delete tuple: %w", err)
	}
	return nil
}

func (s *PostgresStore) ReadTuples(ctx context.Context, object, relation string) ([]model.RelationTuple, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT object_id, relation, subject FROM relation_tuples WHERE object_id = $1 AND relation = $2`,
		object, relation)
	if err != nil {
		return nil, fmt.Errorf("postgres: gagal read tuples: %w", err)
	}
	defer rows.Close()

	var tuples []model.RelationTuple
	for rows.Next() {
		var t model.RelationTuple
		if err := rows.Scan(&t.Object, &t.Relation, &t.Subject); err != nil {
			return nil, fmt.Errorf("postgres: gagal scan row: %w", err)
		}
		tuples = append(tuples, t)
	}
	return tuples, nil
}

// --- APIKeyStore ---

func (s *PostgresStore) CreateAPIKey(ctx context.Context, keyHash string, key model.APIKey) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO api_keys (key_hash, client_name, rate_limit_rpm, is_active)
		 VALUES ($1, $2, $3, $4)`,
		keyHash, key.ClientName, key.RateLimitRPM, key.IsActive)
	if err != nil {
		return fmt.Errorf("postgres: gagal insert api key: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetAPIKeyByHash(ctx context.Context, keyHash string) (model.APIKey, error) {
	var key model.APIKey
	err := s.db.QueryRowContext(ctx,
		`SELECT id, client_name, rate_limit_rpm, is_active FROM api_keys WHERE key_hash = $1`,
		keyHash,
	).Scan(&key.ID, &key.ClientName, &key.RateLimitRPM, &key.IsActive)
	if err == sql.ErrNoRows {
		return model.APIKey{}, store.ErrNotFound
	}
	if err != nil {
		return model.APIKey{}, fmt.Errorf("postgres: gagal query api key: %w", err)
	}
	return key, nil
}

func (s *PostgresStore) RevokeAPIKey(ctx context.Context, keyHash string) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE api_keys SET is_active = false WHERE key_hash = $1`, keyHash)
	if err != nil {
		return fmt.Errorf("postgres: gagal revoke api key: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (s *PostgresStore) RecordAudit(ctx context.Context, entry model.AuditEntry) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO audit_log (actor, action, target, detail) VALUES ($1, $2, $3, $4)`,
		entry.Actor, entry.Action, entry.Target, entry.Detail)
	if err != nil {
		return fmt.Errorf("postgres: gagal insert audit log: %w", err)
	}
	return nil
}