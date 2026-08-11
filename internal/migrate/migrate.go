// Package migrate membungkus golang-migrate supaya bisa dipanggil
// langsung dari kode Go saat startup server, bukan cuma lewat CLI.
package migrate

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlserver"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Up menjalankan semua migration yang belum diterapkan. migrationsPath
// biasanya "file://migrations" (relatif dari working directory).
func Up(connString, migrationsPath string) error {
	m, err := migrate.New(migrationsPath, connString)
	if err != nil {
		return fmt.Errorf("migrate: gagal inisialisasi: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate: gagal jalankan migration: %w", err)
	}
	return nil
}
