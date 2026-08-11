// Package mssql berisi implementasi store.Store menggunakan SQL Server
// sebagai backend persistent.
package mssql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/amayones/authz-engine/internal/model"
	"github.com/amayones/authz-engine/internal/store"
	_ "github.com/microsoft/go-mssqldb"
)

type MSSQLStore struct {
	db *sql.DB
}

// New membuka koneksi ke SQL Server. connString formatnya:
// "sqlserver://user:password@host:port?database=dbname"
// PoolConfig mengatur batas connection pool. Nilai default (lewat New)
// aman untuk service kecil-menengah; sesuaikan lewat NewWithConfig kalau
// traffic sudah lebih besar.
type PoolConfig struct {
	MaxOpenConns    int           // total koneksi aktif maksimum ke SQL Server
	MaxIdleConns    int           // koneksi menganggur yang tetap disimpan (reuse)
	ConnMaxLifetime time.Duration // paksa recycle koneksi setelah durasi ini
	ConnMaxIdleTime time.Duration // tutup koneksi idle setelah durasi ini
}

// DefaultPoolConfig cocok untuk service dengan traffic rendah-menengah
// (puluhan hingga ratusan request/detik). Naikkan MaxOpenConns kalau
// metrics nanti menunjukkan banyak request menunggu koneksi (lihat
// db.Stats().WaitCount).
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

// New membuka koneksi ke SQL Server dengan pool config default.
func New(connString string) (*MSSQLStore, error) {
	return NewWithConfig(connString, DefaultPoolConfig())
}

// NewWithConfig sama seperti New, tapi pool-nya bisa dikustomisasi —
// dipakai kalau service butuh tuning khusus (misal batch job yang
// butuh banyak koneksi paralel, atau service kecil yang mau hemat).
func NewWithConfig(connString string, cfg PoolConfig) (*MSSQLStore, error) {
	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, fmt.Errorf("mssql: gagal buka koneksi: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("mssql: gagal ping database: %w", err)
	}

	return &MSSQLStore{db: db}, nil
}

func (s *MSSQLStore) Close() error {
	return s.db.Close()
}

// Stats mengembalikan statistik pool saat ini — berguna untuk expose
// lewat endpoint /health atau metrics nanti (lihat berapa banyak
// koneksi sedang dipakai, berapa yang menunggu, dll).
func (s *MSSQLStore) Stats() sql.DBStats {
	return s.db.Stats()
}

// --- RoleStore ---

func (s *MSSQLStore) CreateRole(ctx context.Context, role model.Role) error {
	permJSON, err := json.Marshal(role.Permissions)
	if err != nil {
		return fmt.Errorf("mssql: gagal encode permissions: %w", err)
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO roles (name, permissions) VALUES (@p1, @p2)`,
		role.Name, string(permJSON))
	if err != nil {
		return fmt.Errorf("mssql: gagal insert role: %w", err)
	}
	return nil
}

func (s *MSSQLStore) GetRole(ctx context.Context, name string) (model.Role, error) {
	var permJSON string
	err := s.db.QueryRowContext(ctx,
		`SELECT permissions FROM roles WHERE name = @p1`, name,
	).Scan(&permJSON)

	if err == sql.ErrNoRows {
		return model.Role{}, store.ErrNotFound
	}
	if err != nil {
		return model.Role{}, fmt.Errorf("mssql: gagal query role: %w", err)
	}

	var perms []model.Permission
	if err := json.Unmarshal([]byte(permJSON), &perms); err != nil {
		return model.Role{}, fmt.Errorf("mssql: gagal decode permissions: %w", err)
	}
	return model.Role{Name: name, Permissions: perms}, nil
}

func (s *MSSQLStore) UpdateRole(ctx context.Context, role model.Role) error {
	permJSON, err := json.Marshal(role.Permissions)
	if err != nil {
		return fmt.Errorf("mssql: gagal encode permissions: %w", err)
	}
	result, err := s.db.ExecContext(ctx,
		`UPDATE roles SET permissions = @p1 WHERE name = @p2`,
		string(permJSON), role.Name)
	if err != nil {
		return fmt.Errorf("mssql: gagal update role: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (s *MSSQLStore) DeleteRole(ctx context.Context, name string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM roles WHERE name = @p1`, name)
	if err != nil {
		return fmt.Errorf("mssql: gagal delete role: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (s *MSSQLStore) ListRoles(ctx context.Context) ([]model.Role, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT name, permissions FROM roles`)
	if err != nil {
		return nil, fmt.Errorf("mssql: gagal list roles: %w", err)
	}
	defer rows.Close()

	var roles []model.Role
	for rows.Next() {
		var name, permJSON string
		if err := rows.Scan(&name, &permJSON); err != nil {
			return nil, fmt.Errorf("mssql: gagal scan row: %w", err)
		}
		var perms []model.Permission
		if err := json.Unmarshal([]byte(permJSON), &perms); err != nil {
			return nil, fmt.Errorf("mssql: gagal decode permissions: %w", err)
		}
		roles = append(roles, model.Role{Name: name, Permissions: perms})
	}
	return roles, nil
}

// --- SubjectStore ---

func (s *MSSQLStore) AssignRole(ctx context.Context, subjectID, roleName string) error {
	_, err := s.db.ExecContext(ctx, `
		IF NOT EXISTS (SELECT 1 FROM subject_roles WHERE subject_id = @p1 AND role_name = @p2)
		INSERT INTO subject_roles (subject_id, role_name) VALUES (@p1, @p2)`,
		subjectID, roleName)
	if err != nil {
		return fmt.Errorf("mssql: gagal assign role: %w", err)
	}
	return nil
}

func (s *MSSQLStore) RevokeRole(ctx context.Context, subjectID, roleName string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM subject_roles WHERE subject_id = @p1 AND role_name = @p2`,
		subjectID, roleName)
	if err != nil {
		return fmt.Errorf("mssql: gagal revoke role: %w", err)
	}
	return nil
}

func (s *MSSQLStore) GetSubjectRoles(ctx context.Context, subjectID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT role_name FROM subject_roles WHERE subject_id = @p1`, subjectID)
	if err != nil {
		return nil, fmt.Errorf("mssql: gagal query subject roles: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("mssql: gagal scan row: %w", err)
		}
		roles = append(roles, name)
	}
	return roles, nil
}

// --- AttributeStore ---

func (s *MSSQLStore) SetAttribute(ctx context.Context, subjectID, key, value string) error {
	_, err := s.db.ExecContext(ctx, `
		MERGE subject_attributes AS target
		USING (SELECT @p1 AS subject_id, @p2 AS attr_key, @p3 AS attr_value) AS src
		ON target.subject_id = src.subject_id AND target.attr_key = src.attr_key
		WHEN MATCHED THEN UPDATE SET attr_value = src.attr_value
		WHEN NOT MATCHED THEN INSERT (subject_id, attr_key, attr_value)
			VALUES (src.subject_id, src.attr_key, src.attr_value);`,
		subjectID, key, value)
	if err != nil {
		return fmt.Errorf("mssql: gagal set attribute: %w", err)
	}
	return nil
}

func (s *MSSQLStore) GetAttributes(ctx context.Context, subjectID string) (model.Attributes, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT attr_key, attr_value FROM subject_attributes WHERE subject_id = @p1`, subjectID)
	if err != nil {
		return nil, fmt.Errorf("mssql: gagal query attributes: %w", err)
	}
	defer rows.Close()

	attrs := make(model.Attributes)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("mssql: gagal scan row: %w", err)
		}
		attrs[k] = v
	}
	return attrs, nil
}

func (s *MSSQLStore) DeleteAttribute(ctx context.Context, subjectID, key string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM subject_attributes WHERE subject_id = @p1 AND attr_key = @p2`,
		subjectID, key)
	if err != nil {
		return fmt.Errorf("mssql: gagal delete attribute: %w", err)
	}
	return nil
}

// --- RelationStore ---

func (s *MSSQLStore) WriteTuple(ctx context.Context, t model.RelationTuple) error {
	_, err := s.db.ExecContext(ctx, `
		IF NOT EXISTS (SELECT 1 FROM relation_tuples WHERE object_id = @p1 AND relation = @p2 AND subject = @p3)
		INSERT INTO relation_tuples (object_id, relation, subject) VALUES (@p1, @p2, @p3)`,
		t.Object, t.Relation, t.Subject)
	if err != nil {
		return fmt.Errorf("mssql: gagal write tuple: %w", err)
	}
	return nil
}

func (s *MSSQLStore) DeleteTuple(ctx context.Context, t model.RelationTuple) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM relation_tuples WHERE object_id = @p1 AND relation = @p2 AND subject = @p3`,
		t.Object, t.Relation, t.Subject)
	if err != nil {
		return fmt.Errorf("mssql: gagal delete tuple: %w", err)
	}
	return nil
}

func (s *MSSQLStore) ReadTuples(ctx context.Context, object, relation string) ([]model.RelationTuple, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT object_id, relation, subject FROM relation_tuples WHERE object_id = @p1 AND relation = @p2`,
		object, relation)
	if err != nil {
		return nil, fmt.Errorf("mssql: gagal read tuples: %w", err)
	}
	defer rows.Close()

	var tuples []model.RelationTuple
	for rows.Next() {
		var t model.RelationTuple
		if err := rows.Scan(&t.Object, &t.Relation, &t.Subject); err != nil {
			return nil, fmt.Errorf("mssql: gagal scan row: %w", err)
		}
		tuples = append(tuples, t)
	}
	return tuples, nil
}

// --- APIKeyStore ---

func (s *MSSQLStore) CreateAPIKey(ctx context.Context, keyHash string, key model.APIKey) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO api_keys (key_hash, client_name, rate_limit_rpm, is_active)
		 VALUES (@p1, @p2, @p3, @p4)`,
		keyHash, key.ClientName, key.RateLimitRPM, key.IsActive)
	if err != nil {
		return fmt.Errorf("mssql: gagal insert api key: %w", err)
	}
	return nil
}

func (s *MSSQLStore) GetAPIKeyByHash(ctx context.Context, keyHash string) (model.APIKey, error) {
	var key model.APIKey
	err := s.db.QueryRowContext(ctx,
		`SELECT id, client_name, rate_limit_rpm, is_active FROM api_keys WHERE key_hash = @p1`,
		keyHash,
	).Scan(&key.ID, &key.ClientName, &key.RateLimitRPM, &key.IsActive)
	if err == sql.ErrNoRows {
		return model.APIKey{}, store.ErrNotFound
	}
	if err != nil {
		return model.APIKey{}, fmt.Errorf("mssql: gagal query api key: %w", err)
	}
	return key, nil
}

func (s *MSSQLStore) RevokeAPIKey(ctx context.Context, keyHash string) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE api_keys SET is_active = 0 WHERE key_hash = @p1`, keyHash)
	if err != nil {
		return fmt.Errorf("mssql: gagal revoke api key: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return store.ErrNotFound
	}
	return nil
}
