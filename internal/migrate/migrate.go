// Package migrate membungkus golang-migrate supaya bisa dipanggil
// langsung dari kode Go saat startup server, bukan cuma lewat CLI.
package migrate

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlserver"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Driver adalah tipe database yang didukung untuk migration.
type Driver string

const (
	DriverSQLServer Driver = "sqlserver"
	DriverPostgres  Driver = "postgres"
)

// MigrationPath mengembalikan path folder migration untuk driver tertentu.
// Driver postgres menggunakan folder "migrations/postgres", driver sqlserver
// menggunakan folder "migrations" yang sudah ada sebelumnya.
func MigrationPath(driver Driver) string {
	if driver == DriverPostgres {
		return "file://migrations/postgres"
	}
	return "file://migrations"
}

// Up menjalankan semua migration yang belum diterapkan. driver menentukan
// folder migration mana yang dipakai (lihat MigrationPath).
func Up(connString string, driver Driver) error {
	m, err := migrate.New(MigrationPath(driver), connString)
	if err != nil {
		return fmt.Errorf("migrate: gagal inisialisasi: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate: gagal jalankan migration: %w", err)
	}
	return nil
}
